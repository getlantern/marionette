package marionette

import (
	"errors"
	"io"
	"net"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	// ErrStreamClosed is returned enqueuing cells or writing data to a closed stream.
	// Dequeuing cells and reading data will be available until pending data is exhausted.
	ErrStreamClosed = errors.New("marionette: stream closed")

	// ErrWriteTooLarge is returned when a Write() is larger than the buffer.
	ErrWriteTooLarge = errors.New("marionette: write too large")
)

// Ensure type implements interface.
var _ net.Conn = &Stream{}

// Stream represents a readable and writable connection for plaintext data.
// Data is injected into the stream using cells which provide ordering and payload data.
// Implements the net.Conn interface.
type Stream struct {
	mu  sync.RWMutex
	id  int
	seq int

	once    sync.Once
	closed  bool
	closing chan struct{}

	localAddr  net.Addr
	remoteAddr net.Addr

	rbuf, wbuf []byte
	rqueue     []*Cell
	rnotify    chan struct{}
	wnotify    chan struct{}
}

func NewStream(id int) *Stream {
	return &Stream{
		id:      id,
		rbuf:    make([]byte, 0, MaxCellLength),
		wbuf:    make([]byte, 0, MaxCellLength),
		closing: make(chan struct{}),
		rnotify: make(chan struct{}),
		wnotify: make(chan struct{}),
	}
}

// ID returns the stream id.
func (s *Stream) ID() int { return s.id }

// ReadNotify returns a channel that receives a notification when a new read is available.
func (s *Stream) ReadNotify() <-chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rnotify
}

func (s *Stream) notifyRead() {
	close(s.rnotify)
	s.rnotify = make(chan struct{})
}

// WriteNotify returns a channel that receives a notification when a new write is available.
func (s *Stream) WriteNotify() <-chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wnotify
}

func (s *Stream) notifyWrite() {
	close(s.wnotify)
	s.wnotify = make(chan struct{})
}

// Read reads n bytes from the stream.
func (s *Stream) Read(b []byte) (n int, err error) {
	for {
		// Attempt to read from the buffer. Exit if bytes read or error.
		s.mu.Lock()
		if n, err = s.read(b); n != 0 || err != nil {
			s.mu.Unlock()
			return n, err
		} else if n == 0 && len(s.rqueue) == 0 && s.closed {
			s.mu.Unlock()
			return 0, io.EOF
		}
		notify := s.rnotify

		s.processReadQueue()
		s.mu.Unlock()

		// Wait for notification of new read buffer bytes.
		select {
		case <-s.closing:
		case <-notify:
		}
	}
}

func (s *Stream) read(b []byte) (n int, err error) {
	if len(s.rbuf) == 0 {
		return 0, nil
	}

	// Copy bytes to caller.
	n = len(b)
	if n > len(s.rbuf) {
		n = len(s.rbuf)
	}
	copy(b, s.rbuf)

	// Remove bytes from buffer.
	copy(s.rbuf, s.rbuf[n:])
	s.rbuf = s.rbuf[:len(s.rbuf)-n]

	return n, nil
}

// ReadBufferLen returns the number of bytes in the read buffer.
func (s *Stream) ReadBufferLen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.rbuf)
}

// Write appends b to the write buffer. This method will continue to try until
// the entire byte slice is written atomically to the buffer.
func (s *Stream) Write(b []byte) (n int, err error) {
	for {
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return 0, ErrStreamClosed
		} else if n, err = s.write(b); n != 0 || err != nil {
			s.notifyWrite()
			s.mu.Unlock()
			return n, err
		}
		notify := s.wnotify
		s.mu.Unlock()

		// Wait for a change in the write buffer.
		select {
		case <-s.closing:
		case <-notify:
		}
	}
}

func (s *Stream) write(b []byte) (n int, err error) {
	if len(b) > cap(s.wbuf) {
		return 0, ErrWriteTooLarge
	} else if len(b) > cap(s.wbuf)-len(s.wbuf) {
		return 0, nil // not enough space
	}

	// Copy bytes to the end of the write buffer.
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

// Enqueue pushes a cell's payload on to the stream if it is the next sequence.
// Out of sequence cells are added to the queue and are read after earlier cells.
func (s *Stream) Enqueue(cell *Cell) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStreamClosed
	}

	// If sequence is out of order then add to queue and exit.
	if cell.SequenceID < s.seq {
		s.logger().Debug("duplicate cell", zap.Int("sequence_id", cell.SequenceID))
		return nil // duplicate cell
	}

	// Add to queue & sort.
	s.rqueue = append(s.rqueue, cell)
	sort.Sort(Cells(s.rqueue))

	s.processReadQueue()

	return nil
}

func (s *Stream) processReadQueue() {
	// Read all consecutive cells onto the buffer.
	var notify bool
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
		notify = true

		// Shift cell off queue and increment sequence.
		s.rqueue[0] = nil
		s.rqueue = s.rqueue[1:]
		s.seq++
	}

	// Notify of read buffer change.
	if notify {
		s.notifyRead()
	}
}

// Dequeue reads n bytes from the write buffer and encodes it as a cell.
func (s *Stream) Dequeue(n int) *Cell {
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
	if payloadN > 0 {
		cell.Payload = make([]byte, payloadN)
		copy(cell.Payload, s.wbuf[:payloadN])

		// Remove payload bytes from buffer.
		remaining := len(s.wbuf) - payloadN
		copy(s.wbuf[:remaining], s.wbuf[payloadN:len(s.wbuf)])
		s.wbuf = s.wbuf[:remaining]

		// Send notification that write buffer has changed.
		s.notifyWrite()
	}

	return cell
}

// Close marks the stream as closed.
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.once.Do(func() {
		s.closed = true
		close(s.closing)
	})
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
