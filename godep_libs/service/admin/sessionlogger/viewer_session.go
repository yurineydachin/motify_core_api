package sessionlogger

import (
	"time"
)

type ViewerSession struct {
	ParentID       uint64         `json:"parent_id"`
	Caption        string         `json:"caption"`
	RequestTime    time.Time      `json:"request_time"`
	RawRequestDump string         `json:"raw_request_dump"`
	Responses      []*response    `json:"responses"`
	Errors         []*response    `json:"errors"`
	Children       ViewerSessions `json:"children"`
	ErrorMessage   string         `json:"error_message"`
}

type response struct {
	Time         time.Time `json:"time"`
	RawDump      string    `json:"raw_dump"`
	ErrorMessage string    `json:"error_message"`
}

type viewSessionIndexEntry struct {
	parentID  uint64
	request   int64
	responses []int64
	errors    []int64
	children  []*viewSessionIndexEntry
	isDumped  bool
}

type logLine struct {
	requestTime  time.Time
	traceID      uint64
	id           uint64
	parentID     uint64
	lineType     string
	caption      string
	rawDump      string
	errorMessage string
}

type ViewerSessions []*ViewerSession

func (l ViewerSessions) Len() int {
	return len(l)
}

func (l ViewerSessions) Less(i, j int) bool {
	return l[i].RequestTime.Before(l[j].RequestTime)
}

func (l ViewerSessions) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
