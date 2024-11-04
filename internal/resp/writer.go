package resp

import (
	"fmt"
	"io"
)

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func (w *Writer) Write(value Value) error {
	switch value.Type {
	case SimpleString:
		return w.writeString('+', value.Str)
	case Error:
		return w.writeString('-', value.Str)
	case Integer:
		return w.writeInteger(value.Num)
	case BulkString:
		return w.writeBulkString(value)
	case Array:
		return w.writeArray(value)
	default:
		return fmt.Errorf("unknown RESP type: %v", value.Type)
	}
}

func (w *Writer) writeString(prefix byte, s string) error {
	_, err := fmt.Fprintf(w.writer, "%c%s\r\n", prefix, s)
	return err
}

func (w *Writer) writeInteger(n int64) error {
	_, err := fmt.Fprintf(w.writer, ":%d\r\n", n)
	return err
}

func (w *Writer) writeBulkString(value Value) error {
	if value.IsNil {
		_, err := fmt.Fprintf(w.writer, "$-1\r\n")
		return err
	}
	_, err := fmt.Fprintf(w.writer, "$%d\r\n%s\r\n", len(value.Str), value.Str)
	return err
}

func (w *Writer) writeArray(value Value) error {
	if value.IsNil {
		_, err := fmt.Fprintf(w.writer, "*-1\r\n")
		return err
	}
	_, err := fmt.Fprintf(w.writer, "*%d\r\n", len(value.Array))
	if err != nil {
		return err
	}
	for _, item := range value.Array {
		if err := w.Write(item); err != nil {
			return err
		}
	}
	return nil
}
