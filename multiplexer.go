package marionette

import (
	"math/rand"
	"sort"
	"sync"
)

type Stream struct {
	stream_id int
	incoming  *IncomingBuffer
	outgoing  *OutgoingBuffer
	srv_queue *TwistedDeferredQueue
	host      *Stream
	buffer    []byte
}

func NewStream(incoming *IncomingBuffer, outgoing *OutgoingBuffer, stream_id int, srv_queue *TwistedDeferredQueue) *Stream {
	return &Stream{
		incoming:  incoming,
		outgoing:  outgoing,
		stream_id: stream_id,
		srv_queue: srv_queue,
	}
}

func (s *Stream) terminate() {
	s.outgoing.terminate(s.stream_id)
	if s.host != nil {
		s.host.terminate()
	}
}

func (s *Stream) get_stream_id() int {
	return s.stream_id
}

func (s *Stream) push(data []byte) {
	s.outgoing.push(s.stream_id, data)
}

func (s *Stream) pop() []byte {
	buffer := s.buffer
	s.buffer = nil
	return buffer
}

func (s *Stream) peek() []byte {
	return s.buffer
}

type IncomingBuffer struct{}

type OutgoingBuffer struct {
	mu                 sync.RWMutex
	fifo_              map[int][]byte
	terminate_         map[int]struct{}
	streams_with_data_ map[int]struct{}
	sequence_nums      map[int]int
}

func NewOutgoingBuffer() *OutgoingBuffer {
	return &OutgoingBuffer{
		fifo_:              make(map[int][]byte),
		terminate_:         make(map[int]struct{}),
		streams_with_data_: make(map[int]struct{}),
		sequence_nums:      make(map[int]int),
	}
}

func (buf *OutgoingBuffer) push(stream_id int, data []byte) bool {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.fifo_[stream_id] = append(buf.fifo_[stream_id], data...)

	if len(data) != 0 {
		buf.streams_with_data_[stream_id] = struct{}{}
	}

	return true
}

func (buf *OutgoingBuffer) pop(model_uuid int, model_instance_id, n int) *Cell {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	assert(model_uuid != 0)
	assert(model_instance_id != 0)

	var cell_obj *Cell

	var stream_id int
	if a := buf.terminatedStreamsWithData(); len(a) > 0 {
		stream_id = a[rand.Intn(len(a))]
	}

	if i := buf.sequence_nums[stream_id]; i == 0 {
		buf.sequence_nums[stream_id] = 1
	}

	var sequence_id int
	if stream_id == 0 {
		sequence_id = 1
	} else {
		sequence_id = buf.sequence_nums[stream_id]
		buf.sequence_nums[stream_id] += 1
	}

	// Determine if we should terminate the stream
	if len(buf.fifo_[stream_id]) == 0 && buf.terminated(stream_id) {
		cell_obj = NewCell(model_uuid, model_instance_id, stream_id, sequence_id, n, END_OF_STREAM)
		delete(buf.terminate_, stream_id)
		delete(buf.fifo_, stream_id)
		delete(buf.sequence_nums, stream_id)
		return cell_obj
	}

	if n > 0 {
		if buf.has_data(stream_id) {
			cell_obj = NewCell(model_uuid, model_instance_id, stream_id, sequence_id, n, NORMAL)
			payload_length := (n - PAYLOAD_HEADER_SIZE_IN_BITS) / 8
			payload := buf.fifo_[stream_id][:payload_length]
			buf.fifo_[stream_id] = buf.fifo_[stream_id][payload_length:]
			cell_obj.Payload = payload
		} else {
			cell_obj = NewCell(model_uuid, model_instance_id, 0, sequence_id, n, NORMAL)
		}
	} else {
		if buf.has_data(stream_id) {
			cell_obj = NewCell(model_uuid, model_instance_id, stream_id, sequence_id, 0, NORMAL)
			payload_length := len(buf.fifo_[stream_id])
			payload := buf.fifo_[stream_id][:payload_length]
			buf.fifo_[stream_id] = buf.fifo_[stream_id][payload_length:]
			cell_obj.Payload = payload
		}
	}

	if len(buf.fifo_[stream_id]) == 0 {
		delete(buf.streams_with_data_, stream_id)
	}

	return cell_obj
}

func (buf *OutgoingBuffer) peek(stream_id int) []byte {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	return buf.fifo_[stream_id]
}

func (buf *OutgoingBuffer) has_data(stream_id int) bool {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	return len(buf.fifo_[stream_id]) > 0
}

func (buf *OutgoingBuffer) has_data_for_any_stream() int {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	// Ignore if there are no streams with data.
	if len(buf.streams_with_data_) == 0 {
		return 0
	}

	// Convert to a slice.
	a := make([]int, 0, len(buf.streams_with_data_))
	for k := range buf.streams_with_data_ {
		a = append(a, k)
	}
	sort.Ints(a)

	return a[rand.Intn(len(a))]
}

func (buf *OutgoingBuffer) terminate(stream_id int) {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.terminate_[stream_id] = struct{}{}
}

func (buf *OutgoingBuffer) terminatedStreamsWithData() []int {
	// Perform a union between streams with data and terminated streams.
	m := make(map[int]struct{})
	for k := range buf.streams_with_data_ {
		if _, ok := buf.terminate_[k]; ok {
			m[k] = struct{}{}
		}
	}
	for k := range buf.terminate_ {
		if _, ok := buf.streams_with_data_[k]; ok {
			m[k] = struct{}{}
		}
	}

	// Convert set to a slice and sort.
	a := make([]int, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	sort.Ints(a)

	return a
}

func (buf *OutgoingBuffer) terminated(stream_id int) bool {
	_, ok := buf.terminate_[stream_id]
	return ok
}

// TEMP
type TwistedDeferredQueue struct{}
