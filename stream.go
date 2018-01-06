package marionette

import (
	"io"
	"net"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Stream struct {
	mu     sync.RWMutex
	id     int
	seq    int
	closed bool

	localAddr  net.Addr
	remoteAddr net.Addr

	rbuf, wbuf   []byte
	rqueue       []*Cell
	rchan, wchan chan []byte
}

func newStream(id int) *Stream {
	return &Stream{
		id:    id,
		rbuf:  make([]byte, 0, MaxCellLength),
		wbuf:  make([]byte, 0, MaxCellLength),
		rchan: make(chan []byte),
		wchan: make(chan []byte),
	}
}

// ID returns the stream id.
func (s *Stream) ID() int { return s.id }

// Read reads n bytes from the stream.
func (s *Stream) Read(b []byte) (n int, err error) {
	// TODO: Loop and wait for read buffer to fill. Sleep in between.
	println("dbg/read!!!", len(s.rbuf))

	// Wait for data if buffer is empty.
	if len(s.rbuf) == 0 {
		println("dbg/read.123.aaa")
		data, ok := <-s.rchan
		println("dbg/read.123", len(data), ok)
		if !ok {
			return 0, io.EOF
		}
		s.mu.Lock()
		s.rbuf = s.rbuf[:len(data)]
		copy(s.rbuf, data)
		s.mu.Unlock()
	}

	// Copy data to caller.
	n = len(b)
	if n > len(s.rbuf) {
		n = len(s.rbuf)
	}
	copy(b, s.rbuf)

	println("dbg/read.DONE", n, len(b), string(b))
	return n, nil
}

// ReadBufferLen returns the number of bytes in the read buffer.
func (s *Stream) ReadBufferLen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.rbuf)
}

// Write appends b to the write buffer.
func (s *Stream) Write(b []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Wait until there is room in the write buffer.

	s.wbuf = s.wbuf[:len(s.wbuf)+len(b)]
	copy(s.wbuf[len(s.wbuf)-len(b):], b)

	return len(b), nil
}

// WriteBufferLen returns the number of bytes in the write buffer.
func (s *Stream) WriteBufferLen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.wbuf)
}

// AddCell pushes a cell's payload on to the stream if it is the next sequence.
// Out of sequence cells are added to the queue and are read after earlier cells.
func (s *Stream) AddCell(cell *Cell) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If sequence is out of order then add to queue and exit.
	if cell.SequenceID < s.seq {
		return // duplicate cell
	}

	// Add to queue & sort.
	s.rqueue = append(s.rqueue, cell)
	sort.Sort(Cells(s.rqueue))

	// Read all consecutive cells onto the buffer.
	for len(s.rqueue) > 0 {
		cell := s.rqueue[0]
		if cell.SequenceID != s.seq {
			break // out-of-order
		} else if len(cell.Payload) > cap(s.rbuf)-len(s.rbuf) {
			break // not enough space on buffer
		}

		// Extend buffer and copy cell payload.
		s.rbuf = s.rbuf[:len(s.rbuf)+len(cell.Payload)]
		copy(s.rbuf[len(s.rbuf)-len(cell.Payload):], cell.Payload)

		// Shift cell off queue and increment sequence.
		s.rqueue[0] = nil
		s.rqueue = s.rqueue[1:]
		s.seq++
	}

	println("dbg/add.cell", string(cell.Payload), len(s.rbuf))
}

// GenerateCell reads n bytes from the write buffer and encodes it as a cell.
func (s *Stream) GenerateCell(n int) *Cell {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Determine the amount of data to read.
	if len(s.wbuf) > n {
		n = len(s.wbuf)
	}
	if n += CellHeaderSize; n > MaxCellLength {
		n = MaxCellLength
	}

	// Determine next sequence.
	sequenceID := s.seq
	s.seq++

	// End stream if there's no more data and it's marked as closed.
	if len(s.wbuf) == 0 && s.closed {
		return NewCell(s.id, sequenceID, n, END_OF_STREAM)
	}

	// Build cell.
	cell := NewCell(s.id, sequenceID, n, NORMAL)

	// Determine payload size.
	payloadN := n - CellHeaderSize
	if payloadN > len(s.wbuf) {
		payloadN = len(s.wbuf)
	}

	// Copy buffer to payload
	cell.Payload = make([]byte, payloadN)
	copy(cell.Payload, s.wbuf[:payloadN])

	// Remove payload bytes from buffer.
	remaining := len(s.wbuf) - payloadN
	copy(s.wbuf[:remaining], s.wbuf[payloadN:len(s.wbuf)])
	s.wbuf = s.wbuf[:remaining]

	return cell
}

// Close marks the stream as closed.
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

// Closed returns true if the stream has been closed.
func (s *Stream) Closed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

func (c *Stream) LocalAddr() net.Addr  { return c.localAddr }
func (c *Stream) RemoteAddr() net.Addr { return c.remoteAddr }

func (c *Stream) SetDeadline(t time.Time) error      { return nil }
func (c *Stream) SetReadDeadline(t time.Time) error  { return nil }
func (c *Stream) SetWriteDeadline(t time.Time) error { return nil }

func (s *Stream) logger() *zap.Logger {
	return Logger.With(zap.Int("stream_id", s.id))
}
