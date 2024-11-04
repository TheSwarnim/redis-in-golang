package resp_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    resp.Value
		wantErr bool
	}{
		{
			name:  "simple string",
			input: "+OK\r\n",
			want:  resp.SimpleStringVal("OK"),
		},
		{
			name:  "error",
			input: "-Error message\r\n",
			want:  resp.ErrorVal("Error message"),
		},
		{
			name:  "integer",
			input: ":1234\r\n",
			want:  resp.IntegerVal(1234),
		},
		{
			name:  "bulk string",
			input: "$5\r\nhello\r\n",
			want:  resp.BulkStringVal("hello"),
		},
		{
			name:  "null bulk string",
			input: "$-1\r\n",
			want:  resp.NullBulkStringVal(),
		},
		{
			name:  "empty bulk string",
			input: "$0\r\n\r\n",
			want:  resp.BulkStringVal(""),
		},
		{
			name:  "array",
			input: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			want: resp.ArrayVal([]resp.Value{
				resp.BulkStringVal("hello"),
				resp.BulkStringVal("world"),
			}),
		},
		{
			name:  "null array",
			input: "*-1\r\n",
			want:  resp.NullArrayVal(),
		},
		{
			name:  "empty array",
			input: "*0\r\n",
			want:  resp.ArrayVal([]resp.Value{}),
		},
		{
			name:  "nested array",
			input: "*2\r\n*2\r\n+hello\r\n+world\r\n$5\r\nhello\r\n",
			want: resp.ArrayVal([]resp.Value{
				resp.ArrayVal([]resp.Value{
					resp.SimpleStringVal("hello"),
					resp.SimpleStringVal("world"),
				}),
				resp.BulkStringVal("hello"),
			}),
		},
		{
			name:    "invalid type",
			input:   "invalid\r\n",
			wantErr: true,
		},
		{
			name:    "incomplete input",
			input:   "+OK",
			wantErr: true,
		},
		{
			name:    "invalid integer",
			input:   ":abc\r\n",
			wantErr: true,
		},
		{
			name:    "invalid bulk string length",
			input:   "$abc\r\n",
			wantErr: true,
		},
		{
			name:    "invalid array length",
			input:   "*abc\r\n",
			wantErr: true,
		},
		{
			name:    "bulk string too short",
			input:   "$5\r\nhi\r\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := resp.NewParser(bytes.NewBufferString(tt.input))
			got, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser__EOF(t *testing.T) {
	parser := resp.NewParser(bytes.NewBuffer(nil))
	_, err := parser.Parse()
	if err != io.EOF {
		t.Errorf("Expected EOF error, got %v", err)
	}
}
