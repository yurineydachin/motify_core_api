package sessionmocker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"motify_core_api/godep_libs/go-config"
	"motify_core_api/godep_libs/mobapi_lib/logger"
	"motify_core_api/godep_libs/mobapi_lib/watcher"
)

type Mocker struct {
	isEnabled   bool
	traceID     string
	serviceName string
	mockParams  string
	filename    string
	mutex       sync.Mutex
}

var mockEnabled = false

var (
	mockPath  = ""
	storeDays = uint(1)

	day       = 24 * time.Hour
	dayFormat = "2006-01-02"

	watchPeriod = 12 * time.Hour
)

func init() {
	config.RegisterBool("force-sessions-mock-enabled", "Path to sessions mock", false)
	config.RegisterString("sessions-mock-path", "Path to sessions mock", "")
	config.RegisterUint("sessions-mock-store-days", "How long store mock in days", 1)
}

func InitSessionMockerFromFlags() error {
	mockEnabled, _ = config.GetBool("force-sessions-mock-enabled")
	mockPath, _ = config.GetString("sessions-mock-path")
	storeDays, _ = config.GetUint("sessions-mock-store-days")

	watcher.WatchForever(clean, watchPeriod)

	return nil
}

func NewMocker(isEnabled bool, traceID string) *Mocker {
	if (isEnabled || mockEnabled) && mockPath != "" {
		return &Mocker{
			isEnabled: true,
			traceID:   traceID,
		}
	}

	return nil
}

func (m *Mocker) NewMocker(serviceName string) *Mocker {
	if !m.Enabled() {
		return m
	}

	return &Mocker{
		isEnabled:   m.isEnabled,
		serviceName: serviceName,
		traceID:     m.traceID,
		filename:    m.filename,
	}
}

func NewMockerWithFilename(serviceName, filename string) *Mocker {
	return &Mocker{
		isEnabled:   true,
		serviceName: serviceName,
		filename:    filename,
	}
}

func (m *Mocker) Enabled() bool {
	return m != nil && (m.isEnabled || mockEnabled)
}

func (m *Mocker) Request(req *http.Request) {
	if !m.Enabled() {
		return
	}

	err := m.createFilePath(req)
	if err != nil {
		logger.Error(nil, "Can't create file path: %s", err)
		return
	}

	record := ""
	record += fmt.Sprintf("--method %s\n", req.Method)
	record += fmt.Sprintf("--request-uri %s\n", req.RequestURI)

	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err == nil {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			if len(body) > 0 {
				record += fmt.Sprintf("--body %s\n", indentJSON(body))
			}
		}
	}

	_ = m.appendToFile(record)
}

func (m *Mocker) SetFields(fields map[string]string) {
	if !m.Enabled() {
		return
	}

	record := ""
	for key, value := range fields {
		record += fmt.Sprintf("--%s %s\n", key, value)
	}

	_ = m.appendToFile(record)
}

func (m *Mocker) ExternalRequestParams(data interface{}) {
	if !m.Enabled() {
		return
	}

	if dataStr, ok := data.(string); ok {
		m.mockParams = dataStr

	} else if j, err := json.Marshal(data); err == nil {
		m.mockParams = string(j)
	}
}

func (m *Mocker) ExternalRequest(req *http.Request, statusCode int, result []byte) error {
	if !m.Enabled() {
		return nil
	}

	record := fmt.Sprintf("--mock %s\t%d\t%s\n", m.serviceName, statusCode, req.URL.RequestURI())
	if m.mockParams != "" {
		record += fmt.Sprintf("--mock-request-body %s\n", indentJSON([]byte(m.mockParams)))
	}
	record += fmt.Sprintf("--mock-response %s\n", indentJSON(result))

	err := m.appendToFile(record)
	if err != nil {
		logger.Error(nil, "Can't append record '%s' to file: %s", record, err)
	}

	return err
}

func (m *Mocker) createFilePath(req *http.Request) error {
	if mockPath == "" {
		return nil
	}

	urlPath := strings.Trim(strings.Replace(req.URL.Path, "/", "_", -1), "_")
	date := today().Format(dayFormat)
	dir := mockPath + "/" + date + "/" + urlPath
	m.filename = dir + "/" + m.traceID

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		logger.Error(nil, "Can't create dir: %s", err)
	}
	return err
}

func (m *Mocker) appendToFile(record string) error {
	record += "\n"

	m.mutex.Lock()
	defer m.mutex.Unlock()

	f, err := os.OpenFile(m.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.Error(nil, "Can't open file '%s': %s", m.filename, err)
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(record); err != nil {
		logger.Error(nil, "Can't write record '%s' to file '%s': %s", record, m.filename, err)
		return err
	}

	return nil
}

func clean() {
	if mockPath == "" {
		return
	}

	dirnames, err := filepath.Glob(mockPath + "/*")
	if err != nil {
		logger.Error(nil, "Can't list '%s': %s", mockPath, err)
	}

	allowedDirnames := map[string]bool{}
	for d := uint(0); d <= storeDays; d++ {
		date := today().Add(-time.Duration(d) * day).Format(dayFormat)
		allowedDirnames[mockPath+"/"+date] = true
	}

	for _, dirname := range dirnames {
		if !allowedDirnames[dirname] {
			os.RemoveAll(dirname)
		}
	}
}

func today() time.Time {
	return time.Now().UTC().Truncate(day)
}

func indentJSON(body []byte) string {
	dst := new(bytes.Buffer)
	err := json.Indent(dst, body, "", "    ")
	if err == nil {
		return dst.String()
	}

	return string(body)
}
