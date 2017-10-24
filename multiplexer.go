package marionette

import (
	"math/rand"
	"sort"
)

type Stream struct {
	stream_id int
	incoming  *IncomingBuffer
	outgoing  *OutgoingBuffer
	srv_queue *TwistedDeferredQueue
	host      *Steam
	buffer    string
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
		s.host.terminate(s.stream_id)
	}
}

func (s *Stream) get_stream_id() int {
	return s.stream_id
}

func (s *Stream) push(data) {
	s.multiplexer_outgoing_.push(s.stream_id_, data)
}

func (s *Stream) pop() {
	buffer := s.buffer
	s.buffer = ""
	return buffer
}

func (s *Stream) peek() {
	return s.buffer
}

type OutgoingBuffer struct {
	mu                 sync.RWMutex
	fifo_              map[int]string
	terminate_         map[int]struct{}
	streams_with_data_ map[int]struct{}
	sequence_nums      map[int]int
}

func NewOutgoingBuffer() *OutgoingBuffer {
	return &OutgoingBuffer{
		fifo_:              make(map[int]string),
		terminate_:         make(map[int]struct{}),
		streams_with_data_: make(map[int]struct{}),
		sequence_nums:      make(map[int]int),
	}
}

func (buf *OutgoingBuffer) push(stream_id int, s string) bool {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	if !buf.fifo_.get(stream_id) {
		buf.fifo_[stream_id] = ""
	}
	buf.fifo_[stream_id] += s

	if s != "" {
		buf.streams_with_data_[stream_id] = struct{}{}
	}

	return true
}

func (buf *OutgoingBuffer) pop(model_uuid int, model_instance_id, n int) *Cell {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	assert(model_uuid != 0)
	assert(model_instance_id != nil)

	var cell_obj *Cell

	var stream_id int
	if a := buf.terminatedStreamsWithData(); len(a) > 0 {
		stream_id = a[rand.Intn(len(a))]
	}

	if i := buf.sequence_nums[stream_id]; i == 0 {
		buf.sequence_nums[stream_id] = 1
	}

	if stream_id == 0 {
		sequence_id = 1
	} else {
		sequence_id = buf.sequence_nums[stream_id]
		buf.sequence_nums[stream_id] += 1
	}

	// Determine if we should terminate the stream
	if buf.fifo_[stream_id] == "" && buf.terminated(stream_id) {
		cell_obj = NewCell(model_uuid, model_instance_id, stream_id, sequence_id, n, marionette_tg.record_layer.END_OF_STREAM)
		delete(buf.terminate_, stream_id)
		delete(buf.fifo_, stream_id)
		delete(buf.sequence_nums, stream_id)
		return cell_obj
	}

	if n > 0 {
		if buf.has_data(stream_id) {
			cell_obj = Cell(model_uuid, model_instance_id, stream_id, sequence_id, n)
			payload_length := (n - PAYLOAD_HEADER_SIZE_IN_BITS) / 8
			payload := buf.fifo_[stream_id][:payload_length]
			buf.fifo_[stream_id] = buf.fifo_[stream_id][payload_length:]
			cell_obj.set_payload(payload)
		} else {
			cell_obj = NewCell(model_uuid, model_instance_id, 0, sequence_id, n)
		}
	} else {
		if buf.has_data(stream_id) {
			cell_obj = NewCell(model_uuid, model_instance_id, stream_id, sequence_id)
			payload_length := len(buf.fifo_[stream_id])
			payload := buf.fifo_[stream_id][:payload_length]
			buf.fifo_[stream_id] = buf.fifo_[stream_id][payload_length:]
			cell_obj.set_payload(payload)
		}
	}

	if buf.fifo_[stream_id] == "" {
		delete(buf.streams_with_data_, stream_id)
	}

	return cell_obj
}

func (buf *OutgoingBuffer) peek(stream_id int) string {
	var retval string
	buf.mu.Lock()
	defer buf.mu.Unlock()
	return buf.fifo_[stream_id]
}

func (buf *OutgoingBuffer) has_data(stream_id) bool {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	return len(buf.fifo_[stream_id]) > 0
}

func (buf *OutgoingBuffer) has_data_for_any_stream(self) int {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	// Ignore if there are no streams with data.
	if len(buf.streams_with_data_) {
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

func (buf *OutgoingBuffer) terminate(stream_id) {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.terminate_[stream_id] = struct{}{}
}

func (buf *OutgoingBuffer) terminatedStreamsWithData() []int {
	// Perform a union between streams with data and terminated streams.
	m := make(map[int]struct{})
	for k := range buf.streams_with_data_ {
		if buf.terminate_[k] {
			m[k]
		}
	}
	for k := range buf.terminate_ {
		if buf.streams_with_data_[k] {
			m[k]
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

type IncomingBuffer struct {
	mu          mu.RWMutex
	fifo_       string
	fifo_len_   int
	output_q    map[int][]*Cell // ?
	curr_seq_id map[int]int
	has_data_   bool
	callback_   func(*Cell)
}

func (buf *IncomingBuffer) addCallback(callback func()) {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	buf.callback_ = callback
}

func (buf *IncomingBuffer) dequeue(cell_stream_id int) {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	remove_keys := make(map[int]struct{})
	for len(buf.output_q[cell_stream_id]) > 0 && buf.output_q[cell_stream_id][0].get_seq_id() == buf.curr_seq_id[cell_stream_id] {
		cell_obj := heapq.heappop(buf.output_q[cell_stream_id])
		buf.curr_seq_id[cell_stream_id] += 1

		log.Printf("Stream %d Dequeue ID %d", cell_stream_id, cell_obj.get_seq_id())

		if cell_obj.get_cell_type() == marionette_tg.record_layer.END_OF_STREAM {
			log.Printf("Removing Stream %d", cell_stream_id)
			remove_keys[cell_stream_id] = struct{}{}
		}

		buf.callback_(cell_obj) // callFromThread()
	}

	for key := range remove_keys {
		delete(buf.output_q, key)
		delete(buf.curr_seq_id, key)
	}
}

func (buf *IncomingBuffer) enqueue(cell_obj *Cell, cell_stream_id int) {
	buf.mu.Lock()
	defer buf.mu.Unlock()
	if _, ok := buf.output_q[cell_stream_id]; !ok {
		buf.output_q[cell_stream_id] = make([]*Cell, 0)
		buf.curr_seq_id[cell_stream_id] = 1
	}
	heapq.heappush(buf.output_q[cell_stream_id], cell_obj)
	log.msg("Stream %d Enqueue ID %d", cell_stream_id, cell_obj.get_seq_id())
}

func (buf *IncomingBuffer) push(s string) bool {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.fifo_ += s
	buf.fifo_len_ += len(s)

	if buf.callback_ {
		for cell_obj := buf.pop(); cell_obj != nil; cell_obj = buf.pop() {
			cell_stream_id := cell_obj.get_stream_id()
			if cell_stream_id > 0 {
				buf.enqueue(cell_obj, cell_stream_id)
				buf.dequeue(cell_stream_id)
			} else {
				buf.callback_(cell_obj) // callFromThread
			}
		}
	}

	return true
}

func (buf *IncomingBuffer) pop() *Cell {
	buf.mu.Lock()
	defer buf.mu.Unlock()

	if len(buf.fifo_) < 8 {
		return nil
	}

	cell_len := bytes_to_long(str(buf.fifo_[:4]))
	cell_obj := unserialize(buf.fifo_[:cell_len])
	buf.fifo_ = buf.fifo_[cell_len:]
	buf.fifo_len_ -= cell_len
	if buf.fifo_len_ < 0 {
		buf.fifo_len_ = 0
	}

	return cell_obj
}
