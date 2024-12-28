package main

import (
	"bytes"
	"io"
	"os"
)

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}

func NewIndenter(prefix string, other io.Writer) io.Writer {
	return &indenter{w: other, prefix: []byte(prefix), trailingNewline: true}
}

type indenter struct {
	w               io.Writer
	b               bytes.Buffer
	trailingNewline bool

	prefix []byte
}

func (i *indenter) Write(p []byte) (int, error) {
	i.b.Reset() // clear the buffer

	for _, b := range p {
		if i.trailingNewline {
			i.b.Write(i.prefix)
			i.trailingNewline = false
		}

		i.b.WriteByte(b)

		if b == '\n' {
			// do not print the prefix right after the newline character as this might
			// be the very last character of the stream and we want to avoid a trailing prefix.
			i.trailingNewline = true
		}
	}

	n, err := i.w.Write(i.b.Bytes())
	if err != nil {
		// never return more than original length to satisfy io.Writer interface
		if n > len(p) {
			n = len(p)
		}
		return n, err
	}

	// return original length to satisfy io.Writer interface
	return len(p), nil
}
