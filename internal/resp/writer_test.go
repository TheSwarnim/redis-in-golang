package resp_test

import (
	"bytes"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

// Helper function to test parsing and writing roundtrip
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		value resp.Value
	}{
		{
			name:  "simple string",
			value: resp.SimpleStringVal("OK"),
		},
		{
			name:  "error",
			value: resp.ErrorVal("Error message"),
		},
		{
			name:  "integer",
			value: resp.IntegerVal(1234),
		},
		{
			name:  "bulk string",
			value: resp.BulkStringVal("hello"),
		},
		{
			name:  "null bulk string",
			value: resp.NullBulkStringVal(),
		},
		{
			name:  "empty bulk string",
			value: resp.BulkStringVal(""),
		},
		{
			name: "array",
			value: resp.ArrayVal([]resp.Value{
				resp.BulkStringVal("hello"),
				resp.BulkStringVal("world"),
			}),
		},
		{
			name:  "null array",
			value: resp.NullArrayVal(),
		},
		{
			name:  "empty array",
			value: resp.ArrayVal([]resp.Value{}),
		},
		{
			name: "nested array",
			value: resp.ArrayVal([]resp.Value{
				resp.ArrayVal([]resp.Value{
					resp.SimpleStringVal("hello"),
					resp.SimpleStringVal("world"),
				}),
				resp.BulkStringVal("hello"),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write the value
			buf := &bytes.Buffer{}
			writer := resp.NewWriter(buf)
			if err := writer.Write(tt.value); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			// Parse it back
			parser := resp.NewParser(buf)
			got, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Compare
			if !bytes.Equal([]byte(got.Str), []byte(tt.value.Str)) {
				t.Errorf("Roundtrip failed: got %v, want %v", got, tt.value)
			}
		})
	}
}
