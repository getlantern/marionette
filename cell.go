package marionette

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"sort"
	"sync"
)

const (
	PAYLOAD_HEADER_SIZE_IN_BITS  = 200
	PAYLOAD_HEADER_SIZE_IN_BYTES = PAYLOAD_HEADER_SIZE_IN_BITS / 8
)

const (
	NORMAL        = 0x1
	END_OF_STREAM = 0x2
	NEGOTIATE     = 0x3
)

type Cell struct {
	Type            int
	Payload         []byte
	SequenceID      int
	CellLength      int
	StreamID        int
	ModelUUID       int
	ModelInstanceID int
}

func NewCell(modelUUID, modelInstanceID, streamID, sequenceID, length, typ int) *Cell {
	assert(streamID != 0)
	return &Cell{
		Type:            typ,
		SequenceID:      sequenceID,
		CellLength:      length,
		StreamID:        streamID,
		ModelUUID:       modelUUID,
		ModelInstanceID: modelInstanceID,
	}
}

func (c *Cell) Compare(other *Cell) int {
	if c.SequenceID < other.SequenceID {
		return -1
	} else if c.SequenceID > other.SequenceID {
		return 1
	}
	return 0
}

func (c *Cell) Equal(other *Cell) bool {
	return bytes.Equal(c.Payload, other.Payload) &&
		c.StreamID == other.StreamID &&
		c.ModelUUID == other.ModelUUID &&
		c.ModelInstanceID == other.ModelInstanceID &&
		c.SequenceID == other.SequenceID
}

func (c *Cell) Size() int {
	return PAYLOAD_HEADER_SIZE_IN_BYTES + len(c.Payload) + c.paddingN()
}

func (c *Cell) paddingN() int {
	return (c.CellLength / 8) - len(c.Payload) - PAYLOAD_HEADER_SIZE_IN_BYTES
}

func (c *Cell) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(c.Size()))
	binary.Write(&buf, binary.BigEndian, uint32(len(c.Payload)))
	binary.Write(&buf, binary.BigEndian, uint32(c.ModelUUID))
	binary.Write(&buf, binary.BigEndian, uint32(c.ModelInstanceID))
	binary.Write(&buf, binary.BigEndian, uint32(c.StreamID))
	binary.Write(&buf, binary.BigEndian, uint32(c.SequenceID))
	binary.Write(&buf, binary.BigEndian, uint32(c.Type))
	buf.Write(c.Payload)
	buf.Write(make([]byte, c.paddingN()))

	assert(buf.Len() == PAYLOAD_HEADER_SIZE_IN_BYTES+len(c.Payload)+c.paddingN())

	return buf.Bytes(), nil
}

func (c *Cell) UnmarshalBinary(data []byte) (err error) {
	r := bytes.NewReader(data)

	// Read cell & payload size.
	var sz, payloadN, v uint32
	if err := binary.Read(r, binary.BigEndian, &sz); err != nil {
		return err
	} else if err := binary.Read(r, binary.BigEndian, &payloadN); err != nil {
		return err
	}

	// Read model uuid.
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return err
	}
	c.ModelUUID = int(v)

	// Read model instance id.
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return err
	}
	c.ModelInstanceID = int(v)

	// Read stream id.
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return err
	}
	c.StreamID = int(v)

	// Read sequence id.
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return err
	}
	c.SequenceID = int(v)

	// Read cell type.
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return err
	}
	c.Type = int(v)

	// Read payload.
	c.Payload = make([]byte, payloadN)
	if _, err := r.Read(c.Payload); err != nil {
		return err
	}

	return nil
}

type CellEncoder struct {
	mu      sync.RWMutex
	streams map[int]*encoderStream
}

// has_data_for_any_stream
func (enc *CellEncoder) ChooseStreamIDWithData() int {
	enc.mu.RLock()
	defer enc.mu.RUnlock()

	streamIDs := enc.streamIDsWithData()
	if len(streamIDs) == 0 {
		return 0
	}
	return streamIDs[rand.Intn(len(streamIDs))]
}

// streamIDsWithData returns a list of stream ids which have data.
func (enc *CellEncoder) streamIDsWithData() []int {
	var a []int
	for streamID, stream := range enc.streams {
		if len(stream.buf) > 0 {
			a = append(a, streamID)
		}
	}
	return a
}

// operableStreamIDs returns a list of stream id which have data or are marked terminated.
func (enc *CellEncoder) operableStreams() []*encoderStream {
	var a []*encoderStream
	for _, stream := range enc.streams {
		if len(stream.buf) > 0 || stream.terminated {
			a = append(a, stream)
		}
	}
	return a
}

func (enc *CellEncoder) Push(streamID int, data []byte) {
	enc.mu.Lock()
	defer enc.mu.Unlock()
	stream := enc.streams[streamID]
	if stream == nil {
		stream = &encoderStream{id: streamID}
		enc.streams[streamID] = stream
	}
	stream.buf = append(stream.buf, data...)
}

func (enc *CellEncoder) Pop(modelUUID int, modelInstanceID int, n int) *Cell {
	enc.mu.Lock()
	defer enc.mu.Unlock()

	assert(modelUUID != 0)
	assert(modelInstanceID != 0)

	var stream *encoderStream
	if streams := enc.operableStreams(); len(streams) > 0 {
		stream = streams[rand.Intn(len(streams))]
	}

	var sequenceID int
	if stream != nil {
		stream.seq++
		sequenceID = stream.seq
	} else {
		sequenceID = 1
	}

	// Determine if we should terminate the stream
	if stream != nil && len(stream.buf) == 0 && stream.terminated {
		delete(enc.streams, stream.id)
		return NewCell(modelUUID, modelInstanceID, stream.id, sequenceID, n, END_OF_STREAM)
	}

	if n > 0 {
		if len(stream.buf) > 0 {
			payloadN := (n - PAYLOAD_HEADER_SIZE_IN_BITS) / 8
			payload := stream.buf[:payloadN]
			stream.buf = stream.buf[payloadN:]

			cell := NewCell(modelUUID, modelInstanceID, stream.id, sequenceID, n, NORMAL)
			cell.Payload = payload
			return cell
		} else {
			return NewCell(modelUUID, modelInstanceID, 0, sequenceID, n, NORMAL)
		}
	} else {
		if len(stream.buf) > 0 {
			payloadN := len(stream.buf)
			payload := stream.buf[:payloadN]
			stream.buf = stream.buf[payloadN:]

			cell := NewCell(modelUUID, modelInstanceID, stream.id, sequenceID, 0, NORMAL)
			cell.Payload = payload
			return cell
		}
	}

	return nil
}

func (enc *CellEncoder) Peek(streamID int) []byte {
	enc.mu.RLock()
	defer enc.mu.RUnlock()
	stream := enc.streams[streamID]
	if stream == nil {
		return nil
	}
	return stream.buf
}

func (enc *CellEncoder) has_data(streamID int) bool {
	enc.mu.RLock()
	defer enc.mu.RUnlock()
	stream := enc.streams[streamID]
	return stream != nil && len(stream.buf) > 0
}

func (enc *CellEncoder) Terminate(streamID int) {
	enc.mu.RLock()
	defer enc.mu.RUnlock()
	stream := enc.streams[streamID]
	if stream != nil {
		stream.terminated = true
	}
}

type encoderStream struct {
	id         int
	buf        []byte
	terminated bool
	seq        int
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
