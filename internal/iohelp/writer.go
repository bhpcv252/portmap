package iohelp

import (
	"fmt"
	"io"
)

type ErrWriter struct {
	W   io.Writer
	Err error
}

func (ew *ErrWriter) Printf(format string, args ...any) {
	if ew.Err == nil {
		_, ew.Err = fmt.Fprintf(ew.W, format, args...)
	}
}

func (ew *ErrWriter) Println(s string) {
	if ew.Err == nil {
		_, ew.Err = fmt.Fprintln(ew.W, s)
	}
}
