package marionette

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"io"
	"log"
	"sort"
	"sync"
)

const (
	CellHeaderSize = 25
)

const (
	NORMAL        = 0x1
	END_OF_STREAM = 0x2
	NEGOTIATE     = 0x3
)

// Cell represents a single unit of data sent between the client & server.
//
// This cell is associated with a specific stream and the encoder/decoders
// handle ordering based on sequence id.
type Cell struct {
	Type       int    // Record type (NORMAL, END_OF_STREAM)
	Payload    []byte // Data
	Length     int    // Size of marshaled data, if specified.
	StreamID   int    // Associated stream
	SequenceID int    // Record number within stream
	UUID       int    // MAR format identifier
	InstanceID int    // MAR instance identifier
}

// NewCell returns a new instance of Cell.
func NewCell(uuid, instanceID, streamID, sequenceID, length, typ int) *Cell {
	assert(streamID != 0)
	return &Cell{
		Type:       typ,
		SequenceID: sequenceID,
		Length:     length,
		StreamID:   streamID,
		UUID:       uuid,
		InstanceID: instanceID,
	}
}

// Compare returns -1 if c has a lower sequence than other, 1 if c has a higher
// sequence than other, and 0 if both cells have the same sequence.
func (c *Cell) Compare(other *Cell) int {
	if c.SequenceID < other.SequenceID {
		return -1
	} else if c.SequenceID > other.SequenceID {
		return 1
	}
	return 0
}

// Equal returns true if the payload, stream, sequence, uuid, and instance are the same.
func (c *Cell) Equal(other *Cell) bool {
	return bytes.Equal(c.Payload, other.Payload) &&
		c.StreamID == other.StreamID &&
		c.UUID == other.UUID &&
		c.InstanceID == other.InstanceID &&
		c.SequenceID == other.SequenceID
}

// Size returns the marshaled size of the cell, in bytes.
func (c *Cell) Size() int {
	return CellHeaderSize + len(c.Payload) + c.paddingN()
}

// paddingN returns the length of padding, in bytes, if a length is specified.
// If no length is provided or the length is smaller than Size() then 0 is returned.
func (c *Cell) paddingN() int {
	n := c.Length - len(c.Payload) - CellHeaderSize
	if n < 0 {
		return 0
	}
	return n
}

// MarshalBinary returns a byte slice with an encoded cell.
func (c *Cell) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, c.Size()))
	binary.Write(buf, binary.BigEndian, uint32(c.Size()))
	binary.Write(buf, binary.BigEndian, uint32(len(c.Payload)))
	binary.Write(buf, binary.BigEndian, uint32(c.UUID))
	binary.Write(buf, binary.BigEndian, uint32(c.InstanceID))
	binary.Write(buf, binary.BigEndian, uint32(c.StreamID))
	binary.Write(buf, binary.BigEndian, uint32(c.SequenceID))
	binary.Write(buf, binary.BigEndian, uint8(c.Type))
	buf.Write(c.Payload)
	buf.Write(make([]byte, c.paddingN()))

	assert(buf.Len() == CellHeaderSize+len(c.Payload)+c.paddingN())

	return buf.Bytes(), nil
}

// UnmarshalBinary decodes a cell from binary-encoded data.
func (c *Cell) UnmarshalBinary(data []byte) (err error) {
	r := bytes.NewReader(data)

	// Read cell & payload size.
	var sz, payloadN, u32 uint32
	if err := binary.Read(r, binary.BigEndian, &sz); err != nil {
		return err
	} else if err := binary.Read(r, binary.BigEndian, &payloadN); err != nil {
		return err
	}

	// Read model uuid.
	if err := binary.Read(r, binary.BigEndian, &u32); err != nil {
		return err
	}
	c.UUID = int(u32)

	// Read model instance id.
	if err := binary.Read(r, binary.BigEndian, &u32); err != nil {
		return err
	}
	c.InstanceID = int(u32)

	// Read stream id.
	if err := binary.Read(r, binary.BigEndian, &u32); err != nil {
		return err
	}
	c.StreamID = int(u32)

	// Read sequence id.
	if err := binary.Read(r, binary.BigEndian, &u32); err != nil {
		return err
	}
	c.SequenceID = int(u32)

	// Read cell type.
	var u8 uint8
	if err := binary.Read(r, binary.BigEndian, &u8); err != nil {
		return err
	}
	c.Type = int(u8)

	// Read payload.
	c.Payload = make([]byte, payloadN)
	if _, err := r.Read(c.Payload); err != nil {
		return err
	}

	return nil
}

// NOTE: CellDecoder == BufferIncoming

type CellDecoder struct {
	mu      sync.RWMutex
	r       io.Reader
	buf     []byte
	streams map[int]*cellDecoderStream
}

func NewCellDecoder(r io.Reader) *CellDecoder {
	return &CellDecoder{
		r:       r,
		streams: make(map[int]*cellDecoderStream),
	}
}

// Decode decodes the next available in-order cell from any stream.
func (dec *CellDecoder) Decode(cell *Cell) error {
	for {
		// Wait until we have enough bytes for at least one record.
		if err := dec.fillBuffer(); err != nil {
			return err
		}

		// Decode cells into heaps.
		if err := dec.decodeToHeaps(); err != nil {
			return err
		}

		// Return next in-order cell, if available. Otherwise retry.
		if other := dec.pop(); other != nil {
			*cell = *other
			return nil
		}
	}
}

// fillBuffer reads enough data from the reader to fill buffer with at least one cell.
func (dec *CellDecoder) fillBuffer() error {
	for {
		// Exit once we have enough data in the buffer.
		if dec.isBufferFull() {
			return nil
		}

		// Read next available bytes.
		b := make([]byte, 4096)
		if n, err := dec.r.Read(b); err != nil {
			return err
		} else {
			dec.buf = append(dec.buf, b[:n]...)
		}
	}
}

// decodeToHeaps decodes all cells in the buffer to per-stream heaps.
func (dec *CellDecoder) decodeToHeaps() error {
	for {
		// Exit if there's not enough data available.
		if !dec.isBufferFull() {
			return nil
		}

		// Slice next record off the buffer.
		n := binary.BigEndian.Uint32(dec.buf[:4])
		data := dec.buf[:n]
		dec.buf = dec.buf[n:]

		// Unmarshal into cell.
		var cell Cell
		if err := cell.UnmarshalBinary(data); err != nil {
			return err
		}

		// Append to new or existing stream.
		if stream := dec.streams[cell.StreamID]; stream == nil {
			stream = &cellDecoderStream{sequenceID: 1, queue: cellHeap{&cell}}
			heap.Init(&stream.queue)
			dec.streams[cell.StreamID] = stream
		} else {
			stream.sequenceID++
			heap.Push(&stream.queue, &cell)
		}
	}
}

// pop returns the next available in-order cell.
func (dec *CellDecoder) pop() *Cell {
	if len(dec.streams) == 0 {
		return nil
	}

	// Find first stream with an available cell.
	for _, streamID := range dec.streamIDs() {
		stream := dec.streams[streamID]

		// Skip stream if no cells in queue or next cell is out-of-order.
		if len(stream.queue) == 0 || stream.sequenceID != stream.queue[0].SequenceID {
			continue
		}

		// Pop next cell and increment next expected sequence id.
		cell := heap.Pop(&stream.queue).(*Cell)
		stream.sequenceID++

		log.Printf("Stream %d Dequeue ID %d", streamID, cell.SequenceID)

		if cell.Type == END_OF_STREAM {
			log.Printf("Removing Stream %d", streamID)
			delete(dec.streams, streamID)
		}
		return cell
	}
	return nil
}

// streamIDs returns a list of ordered available stream ids.
func (dec *CellDecoder) streamIDs() []int {
	a := make([]int, 0, len(dec.streams))
	for streamID := range dec.streams {
		a = append(a, streamID)
	}
	sort.Ints(a)
	return a
}

// isBufferFull returns true if there is enough data to deserialize at least one cell.
func (dec *CellDecoder) isBufferFull() bool {
	return len(dec.buf) >= 4 && len(dec.buf) >= int(binary.BigEndian.Uint32(dec.buf[:4]))
}

type cellDecoderStream struct {
	sequenceID int
	queue      cellHeap
}

type cellHeap []*Cell

func (q cellHeap) Len() int           { return len(q) }
func (q cellHeap) Less(i, j int) bool { return q[i].Compare(q[j]) == -1 }
func (q cellHeap) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }

func (q *cellHeap) Push(x interface{}) {
	*q = append(*q, x.(*Cell))
}

func (q *cellHeap) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[0 : n-1]
	return item
}
