package bootstrap

import (
	"bytes"
	"fmt"
	"github.com/felixge/httpsnoop"
	"io"
	"io/ioutil"
	"log/slog"
	"net/http"
	"openai/internal/util"
	"strings"
	"time"
)

// HTTPReqInfo LogReqInfo describes info about HTTP request
type HTTPReqInfo struct {
	// GET etc.
	method  string
	uri     string
	referer string
	ipaddr  string
	// response code, like 200, 404
	code int
	// number of bytes of the response sent
	size int64
	// how long did it take to
	duration  time.Duration
	userAgent string
	user      string
	msgId     int64
	content   string
	body      string
}

type Msg struct {
	FromUserName string `xml:"FromUserName"`
	Content      string `xml:"Content"`
	MsgId        int64  `xml:"MsgId,omitempty"`
}

type Image struct {
	MediaId string `xml:"MediaId"`
}

func LogRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		ri := &HTTPReqInfo{
			method:    r.Method,
			uri:       r.URL.Path,
			referer:   r.Header.Get("Referer"),
			userAgent: r.Header.Get("User-Agent"),
			body:      string(bodyBytes),
		}

		ri.ipaddr = requestGetRemoteAddress(r)

		// this runs handler h and captures information about
		// HTTP request
		m := httpsnoop.CaptureMetrics(h, w, r)

		ri.code = m.Code
		ri.size = m.Written
		ri.duration = m.Duration
		logHTTPReq(ri)
	}
	return http.HandlerFunc(fn)
}

// Request.RemoteAddress contains port, which we want to remove i.e.:
// "[::1]:58292" => "[::1]"
func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// requestGetRemoteAddress returns ip address of the client making the request,
// taking into account http proxies
func requestGetRemoteAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIP := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address
		return parts[0]
	}
	return hdrRealIP
}

func logHTTPReq(ri *HTTPReqInfo) {
	slog.Info(fmt.Sprintf("[HTTP] %s %s %d %dms %s %dB 「%s」",
		ri.method,
		ri.uri,
		ri.code,
		ri.duration.Milliseconds(),
		ri.ipaddr,
		ri.size,
		util.EscapeNewline(ri.body),
	))
	return
}
