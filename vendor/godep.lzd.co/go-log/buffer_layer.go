package log

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"godep.lzd.co/go-log/internal"
)

type bufferLayer struct {
	writer io.WriteCloser
	stderr bool

	pool                sync.Pool
	createdBuffersCount uint64
	totalLoggedLength   uint64
	totalLoggedCount    uint64
	skippedBuf          uint64

	queue      chan *bytes.Buffer
	sent       uint64
	lost       uint64
	bufferSize int64
	inBuffer   int64

	flush int64

	metrics *internal.LogMetrics
}

func newBufferLayer(w io.WriteCloser, stderr bool, bufferSize int64, workerCount int) *bufferLayer {
	layer := &bufferLayer{
		writer:     w,
		queue:      make(chan *bytes.Buffer, 1e6),
		stderr:     stderr,
		bufferSize: bufferSize,
	}

	layer.pool = sync.Pool{
		New: func() interface{} {
			atomic.AddUint64(&layer.createdBuffersCount, 1)
			return new(bytes.Buffer)
		},
	}

	for i := 0; i < workerCount; i++ {
		go layer.worker()
	}

	return layer
}

func (l *bufferLayer) WriteFrom(f io.WriterTo) error {
	buf := l.getBufferFromPool()

	_, err := f.WriteTo(buf)
	if err != nil {
		l.onMsgWriteFailure(buf, err)
		return err
	}

	l.metrics.MsgSize().Observe(float64(buf.Len()))

	err = l.queueWrite(buf)
	if err != nil {
		l.onMsgWriteFailure(buf, err)
		return err
	}

	return nil
}

func (l *bufferLayer) Close() error {
	/* leak worker goroutine */
	return l.writer.Close()
}

var errQueueFull = errors.New("Queue is out of capacity")
var errMemory = errors.New("Memory limit is reached ")

func (l *bufferLayer) queueWrite(buf *bytes.Buffer) error {
	c := int64(buf.Cap())
	if atomic.LoadInt64(&l.inBuffer)+c > l.bufferSize {
		return errMemory
	}

	select {
	case l.queue <- buf:
		atomic.AddInt64(&l.inBuffer, c)
		l.metrics.BufferMessagesCount().Inc()
		l.metrics.BufferSize().Add(float64(c))
	default:
		return errQueueFull
	}

	return nil
}

func (l *bufferLayer) Flush(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	atomic.StoreInt64(&l.flush, atomic.LoadInt64(&l.inBuffer))
	for atomic.LoadInt64(&l.flush) > 0 {
		select {
		case <-time.After(10 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (l *bufferLayer) worker() {
	for buf := range l.queue {
		c := buf.Cap()
		l.write(buf)
		atomic.AddInt64(&l.inBuffer, -int64(c))
		if atomic.LoadInt64(&l.flush) > 0 {
			atomic.AddInt64(&l.flush, -int64(c))
		}
		l.metrics.BufferMessagesCount().Dec()
		l.metrics.BufferSize().Sub(float64(c))
	}
}

func (l *bufferLayer) write(msg *bytes.Buffer) {
	var t time.Time
	if l.metrics != nil {
		t = time.Now()
	}

	n, err := l.writer.Write(msg.Bytes())

	if l.metrics != nil {
		l.metrics.SocketWriteTime().Observe(float64(time.Since(t)) / float64(time.Second))
	}

	if n != msg.Len() || err != nil {
		l.onMsgWriteFailure(msg, err)
	} else {
		l.onMsgWriteSuccess(msg)
	}
}

func (l *bufferLayer) onMsgWriteSuccess(msg *bytes.Buffer) {
	l.putBufferBackToPool(msg)
	atomic.AddUint64(&l.sent, 1)
}

func (l *bufferLayer) onMsgWriteFailure(msg *bytes.Buffer, err error) {
	if l.stderr {
		writeError(msg.String(), err)
	}
	l.putBufferBackToPool(msg)
	atomic.AddUint64(&l.lost, 1)
	l.metrics.LostMsgs(err).Inc()
}

// bytes.Buffer pool
func (l *bufferLayer) getBufferFromPool() *bytes.Buffer {
	return l.pool.Get().(*bytes.Buffer)
}

func (l *bufferLayer) putBufferBackToPool(buf *bytes.Buffer) {
	atomic.AddUint64(&l.totalLoggedLength, uint64(buf.Len()))
	atomic.AddUint64(&l.totalLoggedCount, 1)

	c := uint64(buf.Cap())
	// reuse only buffer with capacity less then 3 * avg len and less then maxBufferCapacityToReuse
	if c > atomic.LoadUint64(&l.totalLoggedLength)/atomic.LoadUint64(&l.totalLoggedCount)*3 || c > maxBufferCapacityToReuse {
		atomic.AddUint64(&l.skippedBuf, 1)
	} else {
		buf.Reset()
		l.pool.Put(buf)
	}
}

func writeError(msg string, err error) {
	if len(msg) == 0 && err == nil {
		return
	}
	errString := "no err provided"
	if err != nil {
		errString = err.Error()
	}
	os.Stderr.WriteString("go-log: " + errString + "; original message: " + msg)
}
