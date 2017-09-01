package ltsvlogger

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
)

func NewStructuredLogger(logger *log.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

type StructuredLogger struct {
	Logger *log.Logger
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{
		StructuredLogger: l,
		request:          r,
		buf:              &bytes.Buffer{},
	}

	reqID := middleware.GetReqID(r.Context())
	if reqID != "" {
		entry.buf.WriteString("request-id:" + reqID + "\t")
	}

	l.Logger.Print("method:" + r.Method + "\t")

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	// https://golang.org/pkg/net/http/#Request
	// https://www.w3.org/TR/WD-logfile.html
	entry.buf.WriteString("method:" + r.Method + "\t")
	// entry.buf.WriteString("host:" + r.Host + "\t")
	// entry.buf.WriteString("path:" + r.RequestURI + "\t")
	// entry.buf.WriteString("scheme:" + scheme + "\t")
	uri := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
	entry.buf.WriteString("uri:" + uri + "\t")
	entry.buf.WriteString("protocol:" + r.Proto + "\t")
	entry.buf.WriteString("remote-host:" + r.RemoteAddr + "\t")
	return entry
}

type StructuredLoggerEntry struct {
	StructuredLogger *StructuredLogger
	request          *http.Request
	buf              *bytes.Buffer
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.buf.WriteString("status:" + strconv.Itoa(status) + "\t")
	l.buf.WriteString("bytes:" + strconv.Itoa(bytes) + "\t")
	l.buf.WriteString("time-taken:" + fmt.Sprint(elapsed.Seconds()))
	l.StructuredLogger.Logger.Println(l.buf.String())
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.StructuredLogger.Logger.Println("level:ERROR\tmessage:" + escape(fmt.Sprintf("%+v", v)) + "\tstack:" + escape(string(stack)))
}

func escape(s string) string {
	s = strings.Replace(s, "\n", `\n`, -1)
	s = strings.Replace(s, "\r", `\r`, -1)
	s = strings.Replace(s, "\t", `\t`, -1)
	s = strings.Replace(s, "\\", `\\`, -1)
	return s
}
