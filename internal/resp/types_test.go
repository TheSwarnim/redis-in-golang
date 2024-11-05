package resp_test

import (
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func TestValue_Expiry(t *testing.T) {
	tests := []struct {
		name        string
		value       resp.Value
		sleep       time.Duration
		shouldExist bool
	}{
		{
			name:        "no expiry",
			value:       resp.BulkStringVal("hello"),
			sleep:       time.Millisecond,
			shouldExist: true,
		},
		{
			name:        "with expiry - not expired",
			value:       resp.BulkStringValWithExpiry("test", time.Second),
			sleep:       time.Millisecond,
			shouldExist: true,
		},
		{
			name:        "with expiry - expired",
			value:       resp.BulkStringValWithExpiry("test", time.Millisecond),
			sleep:       time.Millisecond * 10,
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			time.Sleep(tt.sleep)
			isExpired := tt.value.IsExpired()
			if isExpired == tt.shouldExist {
				t.Errorf("IsExpired() = %v, want %v", isExpired, !tt.shouldExist)
			}
		})
	}
}

func TestBulkStringValWithExpiry(t *testing.T) {
	val := resp.BulkStringValWithExpiry("test", time.Millisecond)

	// Should not be expired immediately
	if val.IsExpired() {
		t.Error("expected value to not be expired immediately")
	}

	// should be expired after waiting
	time.Sleep(time.Millisecond * 2)
	if !val.IsExpired() {
		t.Error("expected value to be expired after waiting")
	}
}
