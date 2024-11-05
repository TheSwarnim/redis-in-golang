package resp

import "time"

type Type byte

const (
	SimpleString Type = '+'
	Error        Type = '-'
	Integer      Type = ':'
	BulkString   Type = '$'
	Array        Type = '*'
)

// Value represents a RESP value.
type Value struct {
	Type       Type
	Str        string
	Num        int64
	Array      []Value
	IsNil      bool
	ExpiryTime *time.Time // expiry support
}

// IsExpired checks if the value has expired
func (v *Value) IsExpired() bool {
	return v.ExpiryTime != nil && time.Now().After(*v.ExpiryTime)
}

// SetExpiry sets the expiry time for the value
func (v *Value) SetExpiry(duration time.Duration) {
	t := time.Now().Add(duration)
	v.ExpiryTime = &t
}

// Helper functions to create RESP values.
func SimpleStringVal(s string) Value {
	return Value{Type: SimpleString, Str: s}
}

func ErrorVal(s string) Value {
	return Value{Type: Error, Str: s}
}

func IntegerVal(n int64) Value {
	return Value{Type: Integer, Num: n}
}

func BulkStringVal(s string) Value {
	return Value{Type: BulkString, Str: s}
}

// BulkStringValWithExpiry creates a new BulkString value with an expiry time.
func BulkStringValWithExpiry(s string, expiry time.Duration) Value {
	v := BulkStringVal(s)
	v.SetExpiry(expiry)
	return v
}

func NullBulkStringVal() Value {
	return Value{Type: BulkString, IsNil: true}
}

func ArrayVal(a []Value) Value {
	return Value{Type: Array, Array: a}
}

func NullArrayVal() Value {
	return Value{Type: Array, IsNil: true}
}
