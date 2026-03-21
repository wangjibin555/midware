package ErrorHandler

import "net/http"

type trackingWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *trackingWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *trackingWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func (w *trackingWriter) WroteHeader() bool {
	return w.wroteHeader
}
