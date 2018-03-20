package sessionlogger

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"godep.lzd.co/mobapi_lib/logger"
	"godep.lzd.co/mobapi_lib/sessionlogger"
	"godep.lzd.co/mobapi_lib/sessionlogger/libs/uniq_dumper"
)

type Viewer struct {
	logsDir string
}

func NewViewer(logsDir string) *Viewer {
	return &Viewer{
		logsDir: logsDir,
	}
}

func (v *Viewer) GetLogsNames() ([]string, error) {
	dir, err := os.Open(v.logsDir)
	if err != nil {
		return nil, err
	}

	fileInfos, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	logNames := make([]string, 0, len(fileInfos))
	for _, fi := range fileInfos {
		if !fi.IsDir() && strings.HasSuffix(fi.Name(), sessionlogger.LogFileSuffix) {
			logNames = append(logNames, fi.Name())
		}
	}

	return logNames, nil
}

func (v *Viewer) GetSession(file string, traceID string) (*ViewerSession, error) {
	if strings.Contains(file, "/") || strings.Contains(file, "\\") || !strings.HasSuffix(file, sessionlogger.LogFileSuffix) {
		return nil, fmt.Errorf("Invalid filename %s", file)
	}

	fileHandler, err := os.Open(path.Join(v.logsDir, file))
	if err != nil {
		return nil, err
	}
	defer fileHandler.Close()

	sessionsInd := make(map[uint64]*viewSessionIndexEntry)
	reader := bufio.NewReader(fileHandler)
	offset := int64(0)
	for {
		curOffset := offset
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		offset += int64(len(line))

		parts := strings.SplitN(line, "\t", 7)

		if len(parts) == 7 {
			if parts[1] != traceID {
				continue
			}

			id, err := strconv.ParseUint(parts[2], 0, 32)
			if err != nil {
				continue
			}

			parentID, err := strconv.ParseUint(parts[3], 0, 32)
			if err != nil {
				continue
			}

			indEntry, exists := sessionsInd[id]
			if !exists {
				indEntry = &viewSessionIndexEntry{parentID: parentID}
				sessionsInd[id] = indEntry
			}

			switch parts[4] {
			case "REQ":
				indEntry.request = curOffset
			case "RESP":
				indEntry.responses = append(indEntry.responses, curOffset)
			case "ERR":
				indEntry.errors = append(indEntry.errors, curOffset)
			}
		}
	}

	for id, entry := range sessionsInd {
		if id == 0 && entry.parentID == 0 {
			continue
		}
		parentEntry, exists := sessionsInd[entry.parentID]
		if exists {
			parentEntry.children = append(parentEntry.children, entry)
		}
	}

	indEntry, exists := sessionsInd[0]
	if !exists {
		return nil, fmt.Errorf("Session was not found %s", traceID)
	}
	sessionsInd = nil

	logDir := strings.TrimSuffix(fileHandler.Name(), sessionlogger.LogFileSuffix) + sessionlogger.DumpDirSuffix
	logDumper := uniq_dumper.New(logDir)
	return indEntryToSession(indEntry, fileHandler, logDumper)
}

func indEntryToSession(entry *viewSessionIndexEntry, file *os.File, logDumper *uniq_dumper.Dumper) (*ViewerSession, error) {
	session := &ViewerSession{
		ParentID:  entry.parentID,
		Responses: make([]*response, len(entry.responses)),
		Errors:    make([]*response, len(entry.errors)),
		Children:  make(ViewerSessions, len(entry.children)),
	}

	if line, err := readLogLine(file, logDumper, entry.request); err == nil {
		session.RequestTime = line.requestTime
		session.Caption = line.caption
		session.RawRequestDump = line.rawDump
		session.ErrorMessage = line.errorMessage
	} else {
		return nil, err
	}

	for i, respEntry := range entry.responses {
		if line, err := readLogLine(file, logDumper, respEntry); err == nil {
			session.Responses[i] = &response{
				Time:         line.requestTime,
				RawDump:      line.rawDump,
				ErrorMessage: line.errorMessage,
			}
		} else {
			return nil, err
		}
	}

	for i, errEntry := range entry.errors {
		if line, err := readLogLine(file, logDumper, errEntry); err == nil {
			session.Errors[i] = &response{
				Time:         line.requestTime,
				RawDump:      line.rawDump,
				ErrorMessage: line.errorMessage,
			}
		} else {
			return nil, err
		}
	}

	for i, child := range entry.children {
		var err error
		session.Children[i], err = indEntryToSession(child, file, logDumper)
		if err != nil {
			return nil, err
		}
	}
	sort.Sort(session.Children)

	return session, nil
}

func readLogLine(file *os.File, logDumper *uniq_dumper.Dumper, offset int64) (*logLine, error) {
	reader := bufio.NewReader(file)
	file.Seek(offset, os.SEEK_SET)

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	parts := strings.Split(line, "\t")
	if len(parts) < 6 {
		return nil, fmt.Errorf("could not read line %q", line)
	}

	result := &logLine{}
	result.requestTime, _ = time.Parse(time.RFC3339Nano, parts[0])

	var dumpData string
	lineID := parts[1]
	if len(parts) == 7 {
		result.traceID, _ = strconv.ParseUint(lineID, 0, 64)
		result.id, _ = strconv.ParseUint(parts[2], 0, 32)
		result.parentID, _ = strconv.ParseUint(parts[3], 0, 32)
		result.lineType = parts[4]
		result.caption = parts[5]
		dumpData = parts[6]
	}

	dumpData = strings.TrimSpace(dumpData)
	if dumpData != "" {
		if dump, err := logDumper.Read(dumpData); err == nil {
			dumpData = string(dump)
		} else {
			logger.Error(nil, "Can't get dump for '%s': %s", dumpData, err)
		}

	}

	result.rawDump = dumpData

	return result, nil
}
