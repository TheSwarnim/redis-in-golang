package resp

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
	Type  Type
	Str   string
	Num   int64
	Array []Value
	IsNil bool
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

func NullBulkStringVal() Value {
	return Value{Type: BulkString, IsNil: true}
}

func ArrayVal(a []Value) Value {
	return Value{Type: Array, Array: a}
}

func NullArrayVal() Value {
	return Value{Type: Array, IsNil: true}
}
