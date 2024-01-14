package caching

import (
	"io"
)

type progressReader struct {
	reader           io.Reader
	readBytes        int64
	progressCallback func(bytesSoFar int64, bytesSinceLast int64)
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	bytesRead := int64(n)
	r.readBytes += bytesRead
	if r.progressCallback != nil && bytesRead > 0 {
		r.progressCallback(r.readBytes, bytesRead)
	}
	return n, err
}
