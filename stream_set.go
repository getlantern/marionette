package marionette

import (
	"math/rand"
	"net"
	"sync"
)

type StreamSet struct {
	mu      sync.Mutex
	streams map[int]*Stream

	// Network address information injected into each new stream.
	LocalAddr  net.Addr
	RemoteAddr net.Addr

	// Callback invoked whenever a stream is created.
	OnNewStream func(*Stream)
}

// NewStreamSet returns a new instance of StreamSet.
func NewStreamSet() *StreamSet {
	return &StreamSet{
		streams: make(map[int]*Stream),
	}
}

// Close closes all streams in the set.
func (ss *StreamSet) Close() (err error) {
	for _, stream := range ss.streams {
		if e := stream.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

// Stream returns a stream by id.
func (ss *StreamSet) Stream(id int) *Stream {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.streams[id]
}

// Create returns a new stream with a random stream id.
func (ss *StreamSet) Create() *Stream {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.create(0)
}

func (ss *StreamSet) create(id int) *Stream {
	if id == 0 {
		id = int(rand.Int31() + 1)
	}

	stream := NewStream(id)
	stream.localAddr = ss.LocalAddr
	stream.remoteAddr = ss.RemoteAddr
	ss.streams[stream.id] = stream

	// Execute callback, if exists.
	if ss.OnNewStream != nil {
		ss.OnNewStream(stream)
	}

	return stream
}

// Enqueue pushes a cell onto a stream's read queue.
// If the stream doesn't exist then it is created.
func (ss *StreamSet) Enqueue(cell *Cell) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Ignore empty cells.
	if cell.StreamID == 0 {
		return nil
	}

	// Create or find stream and enqueue cell.
	stream := ss.streams[cell.StreamID]
	if stream == nil {
		stream = ss.create(cell.StreamID)
	}
	return stream.Enqueue(cell)
}

// Dequeue returns a cell containing data for a random stream's write buffer.
func (ss *StreamSet) Dequeue(n int) *Cell {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Choose a random stream with data.
	var stream *Stream
	for _, s := range ss.streams {
		if s.WriteBufferLen() > 0 || s.Closed() {
			stream = s
			break
		}
	}

	// If there is no stream with data then send an empty
	if stream == nil {
		return nil
	}

	// Generate cell from stream. Remove from set if at the end.
	cell := stream.Dequeue(n)
	if cell.Type == END_OF_STREAM {
		delete(ss.streams, stream.ID())
	}
	return cell
}
