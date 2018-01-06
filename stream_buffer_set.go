package marionette

/*
import (
	"math/rand"
	"sort"
	"sync"
	"time"
)

// NOTE: StreamBufferSet == BufferOutgoing

// StreamBufferSet represents a set of stream buffers.
type StreamBufferSet struct {
	mu      sync.RWMutex
	streams map[int]*streamBuffer

	Rand *rand.Rand
}

// NewStreamBufferSet returns a new instance of StreamBufferSet.
func NewStreamBufferSet() *StreamBufferSet {
	return &StreamBufferSet{
		streams: make(map[int]*streamBuffer),
		Rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
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
	sort.Sort(streamBuffers(a))
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

type streamBuffers []*streamBuffer

func (a streamBuffers) Len() int           { return len(a) }
func (a streamBuffers) Less(i, j int) bool { return a[i].id < a[j].id }
func (a streamBuffers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
*/
