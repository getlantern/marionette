package marionette

import (
	"math/rand"
	"sync"
)

// NOTE: StreamBufferSet == BufferOutgoing

// StreamBufferSet represents a set of stream buffers.
type StreamBufferSet struct {
	mu      sync.RWMutex
	streams map[int]*streamBuffer
}

// NewStreamBufferSet returns a new instance of StreamBufferSet.
func NewStreamBufferSet() *StreamBufferSet {
	return &StreamBufferSet{
		streams: make(map[int]*streamBuffer),
	}
}

// Push appends data to a stream's buffer.
func (s *StreamBufferSet) Push(streamID int, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stream := s.streams[streamID]
	if stream == nil {
		stream = &streamBuffer{id: streamID}
		s.streams[streamID] = stream
	}
	stream.buf = append(stream.buf, data...)
}

// Pop returns a cell containing data for a random stream.
func (s *StreamBufferSet) Pop(uuid, instanceID, n int, blocking bool) *Cell {
	s.mu.Lock()
	defer s.mu.Unlock()

	assert(uuid != 0)
	assert(instanceID != 0)

	// Choose a stream with data.
	var stream *streamBuffer
	if streams := s.operableStreams(); len(streams) > 0 {
		stream = streams[rand.Intn(len(streams))]
	} else if blocking {
		return NewCell(uuid, instanceID, 0, 1, n, END_OF_STREAM)
	} else {
		return nil
	}

	// Determine the amount of data to read.
	if len(stream.buf) > n {
		n = len(stream.buf)
	}
	if n += CellHeaderSize; n > MaxCellLength {
		n = MaxCellLength
	}

	// Determine next sequence.
	sequenceID := stream.nextSeq()

	// End stream if there's no more data and it's marked as terminated.
	if len(stream.buf) == 0 && stream.terminated {
		delete(s.streams, stream.id)
		return NewCell(uuid, instanceID, stream.id, sequenceID, n, END_OF_STREAM)
	}

	// Build cell.
	cell := NewCell(uuid, instanceID, stream.id, sequenceID, n, NORMAL)

	// Determine payload.
	payloadN := n - CellHeaderSize
	if payloadN < 0 {
		payloadN = 0
	} else if payloadN > len(stream.buf) {
		payloadN = len(stream.buf)
	}
	cell.Payload, stream.buf = stream.buf[:payloadN], stream.buf[payloadN:]

	return cell
}

func (s *StreamBufferSet) Terminate(streamID int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stream := s.streams[streamID]
	if stream != nil {
		stream.terminated = true
	}
}

// operableStreams returns a list of stream id which have data or are marked terminated.
func (s *StreamBufferSet) operableStreams() []*streamBuffer {
	var a []*streamBuffer
	for _, stream := range s.streams {
		if len(stream.buf) > 0 || stream.terminated {
			a = append(a, stream)
		}
	}
	return a
}

type streamBuffer struct {
	id         int
	buf        []byte
	terminated bool
	seq        int
}

func (b *streamBuffer) nextSeq() int {
	b.seq++
	return b.seq
}
