package marionette

import (
	"math/rand"
	"sync"
)

type StreamSet struct {
	mu      sync.Mutex
	streams map[int]*Stream

	// Callback invoked whenever a stream is created.
	OnNewStream func(*Stream)
}

// NewStreamSet returns a new instance of StreamSet.
func NewStreamSet() *StreamSet {
	return &StreamSet{
		streams: make(map[int]*Stream),
	}
}

// Create returns a new stream.
func (ss *StreamSet) Create() *Stream {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.create()
}

func (ss *StreamSet) create() *Stream {
	stream := newStream(int(rand.Int31()))
	ss.streams[stream.id] = stream

	// Execute callback, if exists.
	if ss.OnNewStream != nil {
		ss.OnNewStream(stream)
	}

	return stream
}

// AddCell pushes a cell onto a stream's read queue.
// If the stream doesn't exist then it is created.
func (ss *StreamSet) AddCell(cell *Cell) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	stream := ss.streams[cell.StreamID]
	if stream == nil {
		stream = ss.create()
	}
	stream.AddCell(cell)
}

// GenerateCell returns a cell containing data for a random stream.
func (ss *StreamSet) GenerateCell(n int) *Cell {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Choose a random stream with data.
	var stream *Stream
	for _, s := range ss.streams {
		if s.WriteBufferLen() > 0 && !s.Closed() {
			stream = s
			break
		}
	}
	if stream == nil {
		return nil
	}

	// Generate cell from stream. Remove from set if at the end.
	cell := stream.GenerateCell(n)
	if cell.Type == END_OF_STREAM {
		delete(ss.streams, stream.ID())
	}
	return cell
}
