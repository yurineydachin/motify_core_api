package sessionlogger

import (
	"sync/atomic"
	"time"
)

type rootSession struct {
	traceID   string
	idGen     uint32
	logWriter *FileLogWriter
}

const (
	eventTypeRequest  = "REQ"
	eventTypeResponse = "RESP"
	eventTypeError    = "ERR"
)

// nextID generates child ids starting from 0
func (s *rootSession) nextID() uint32 {
	return atomic.AddUint32(&s.idGen, 1) - 1
}

// Session session
type Session struct {
	root     *rootSession
	id       uint32
	parentID uint32
	caption  string
}

// NewSession newsession
func (s *Session) NewSession(caption string, request interface{}) *Session {
	return startSession(s.root, s.id, caption, request)
}

// NewBlackholeSession creates dummy session that does not write any real data anywhere
func NewBlackholeSession() *Session {
	return &Session{}
}

func startSession(rootSession *rootSession, parentID uint32, caption string, request interface{}) *Session {
	var id uint32 = 0
	if rootSession != nil {
		id = rootSession.nextID()
	}
	childSession := &Session{
		root:     rootSession,
		parentID: parentID,
		id:       id,
		caption:  caption,
	}

	childSession.log(eventTypeRequest, request)

	return childSession
}

// Finish finish
func (s *Session) Finish(response interface{}) {
	s.log(eventTypeResponse, response)
}

// Error error
func (s *Session) Error(err interface{}) {
	s.log(eventTypeError, err)
}

// TraceID session ID
func (s *Session) TraceID() string {
	if s.root == nil {
		return ""
	}
	return s.root.traceID
}

func (s *Session) log(eventType string, data interface{}) {
	isLoggingEnabled := s.root != nil

	if s.root != nil && s.root.logWriter != nil && isLoggingEnabled {
		l := Log{
			traceID:   s.root.traceID,
			id:        s.id,
			parentID:  s.parentID,
			caption:   s.caption,
			eventType: eventType,
			timestamp: time.Now().UTC(),
			data:      data,
		}
		s.root.logWriter.Write(&l)
	}
}

// Log contains information to be logged.
type Log struct {
	traceID   string
	id        uint32
	parentID  uint32
	caption   string
	eventType string
	timestamp time.Time
	data      interface{}
}
