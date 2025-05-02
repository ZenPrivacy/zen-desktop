// Package httprewrite provides utilities for rewriting HTTP responses.
package httprewrite

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// StreamRewrite decodes the HTTP response body and passes it to the processor
// for transformation in a streaming fashion.
//
// It automatically handles decompression and character set conversion to UTF-8.
// The processor receives the unpacked (decompressed and decoded) response body as an io.ReadCloser and
// writes the transformed output to the provided io.PipeWriter. The processor
// is responsible for closing both streams.
func StreamRewrite(res *http.Response, processor func(src io.ReadCloser, dst *io.PipeWriter)) error {
	rawBodyReader, mimeType, err := getRawBodyReader(res)
	if err != nil {
		return fmt.Errorf("get raw body reader: %v", err)
	}

	reader, writer := io.Pipe()

	go processor(rawBodyReader, writer)

	res.Body = reader
	// Content-Length cannot be determined ahead of time.
	// To get around this, the response is chunked to allow for HTTP connection reuse without having to TCP FIN terminate each connection.
	res.ContentLength = -1
	res.Header.Del("Content-Length")
	res.Header.Del("Content-Encoding")
	res.TransferEncoding = []string{"chunked"}
	res.Header.Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", mimeType))
	return nil
}

// BufferRewrite reads and decodes the HTTP response body, applies a transformation
// to it using the provided processor function, and replaces the original body
// with the transformed version.
//
// It automatically handles decompression and character set conversion to UTF-8.
// The processor receives the fully buffered, unpacked (decompressed and decoded) body as input and returns
// a modified byte slice.
func BufferRewrite(res *http.Response, processor func(src []byte) ([]byte, error)) error {
	rawBodyReader, mimeType, err := getRawBodyReader(res)
	if err != nil {
		return fmt.Errorf("get raw body reader: %v", err)
	}

	rawBody, err := io.ReadAll(rawBodyReader)
	if err != nil {
		return fmt.Errorf("read body: %v", err)
	}

	processedBody, err := processor(rawBody)
	if err != nil {
		return fmt.Errorf("process body: %v", err)
	}

	res.Body = io.NopCloser(bytes.NewReader(processedBody))
	res.ContentLength = int64(len(processedBody))
	res.Header.Set("Content-Length", fmt.Sprint(len(processedBody)))
	res.TransferEncoding = nil
	res.Header.Del("Content-Encoding")
	res.Header.Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", mimeType))
	return nil
}
