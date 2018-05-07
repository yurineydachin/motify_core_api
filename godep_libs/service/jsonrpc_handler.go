package service

import (
	"bytes"
	"context"
	"encoding/json"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/opentracing/opentracing-go"
	gotrace "motify_core_api/godep_libs/go-trace"
)

type jsonRPCHandler struct {
	Name            string
	hostname        string
	httpJSONHandler http.Handler
}

func (h *jsonRPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	span, ctx := gotrace.StartSpanFromRequest(r.Context(), h.Name, opentracing.HTTPHeadersCarrier(r.Header))
	defer span.Finish()

	if r.Method != "POST" {
		h.writeError(ctx, w, -32600, "Invalid Request", nil)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.writeError(ctx, w, -32700, "Parse error", nil)
		return
	}
	r.Body.Close()

	var req jsonRPCRequest
	if err := json.Unmarshal(b, &req); err != nil {
		h.writeError(ctx, w, -32700, "Parse error", nil)
		return
	}

	if req.Protocol != "2.0" {
		h.writeError(ctx, w, -32600, "Invalid Request", req.ID)
		return
	}
	if req.Method == "" {
		h.writeError(ctx, w, -32601, "Method not found", req.ID)
		return
	}

	if len(req.Params) == 0 {
		req.Params = json.RawMessage([]byte("{}"))
	}
	proxyReq, err := http.NewRequest("POST", req.Method, bytes.NewBuffer([]byte(req.Params)))
	if err != nil {
		h.writeError(ctx, w, -32602, "Invalid params", req.ID)
		return
	}

	copyHeader(proxyReq.Header, r.Header)

	proxyWriter := httptest.NewRecorder()
	h.httpJSONHandler.ServeHTTP(proxyWriter, proxyReq)

	switch proxyWriter.Code {
	//	case http.StatusOK:
	case http.StatusBadRequest:
		h.writeError(ctx, w, -32600, "Invalid Request", req.ID)
		return
	case http.StatusNotFound:
		h.writeError(ctx, w, -32601, "Method not found", req.ID)
		return
	}

	var sessionResp httpSessionResponse
	if err := json.Unmarshal(proxyWriter.Body.Bytes(), &sessionResp); err != nil {
		h.writeError(ctx, w, -32603, "Internal error", req.ID)
		return
	}

	copyHeader(w.Header(), proxyWriter.Header())
	w.Header().Set("Content-Type", "application/json-rpc")

	switch sessionResp.Result {
	case "OK":
		resp := jsonRPCResponse{
			Protocol: "2.0",
			ID:       req.ID,
			Result:   sessionResp.Data,
		}
		b, err = json.Marshal(&resp)
		if err != nil {
			h.writeError(ctx, w, -32603, "Internal error", req.ID)
		} else {
			gotrace.InjectSpanToResponseFromContext(ctx, opentracing.HTTPHeadersCarrier(w.Header()))
			w.Write(b)
		}
	case "ERROR":
		h.writeError(ctx, w, int(hash(sessionResp.Error)), sessionResp.Error, req.ID)
	default:
		h.writeError(ctx, w, -32603, "Internal error", req.ID)
	}
}

func (*jsonRPCHandler) writeError(ctx context.Context, w http.ResponseWriter,
	code int, message string, id *uint64) {

	gotrace.InjectSpanToResponseFromContext(ctx, opentracing.HTTPHeadersCarrier(w.Header()))
	resp := jsonRPCResponse{
		Protocol: "2.0",
		ID:       id,
		Error: &jsonRPCError{
			Code:    code,
			Message: message,
		},
	}
	b, err := json.Marshal(&resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

type jsonRPCRequest struct {
	Protocol string          `json:"jsonrpc"`
	Method   string          `json:"method"`
	Params   json.RawMessage `json:"params"`
	ID       *uint64         `json:"id,omitempty"`
}

type jsonRPCResponse struct {
	Protocol string          `json:"jsonrpc"`
	Result   json.RawMessage `json:"result,omitempty"`
	ID       *uint64         `json:"id,omitempty"`
	Error    *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TODO: Almost copied from gorpc http_json.go
type httpSessionResponse struct {
	Result string          `json:"result"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}
