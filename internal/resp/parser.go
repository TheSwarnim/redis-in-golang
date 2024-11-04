package resp

/*
Parser reads RESP values from an io.Reader.

Example usage:

	parser := resp.NewParser(strings.NewReader("+OK\r\n"))
	value, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value.Str) // Output: OK

	parser = resp.NewParser(strings.NewReader(":1234\r\n"))
	value, err = parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value.Num) // Output: 1234

	parser = resp.NewParser(strings.NewReader("$5\r\nhello\r\n"))
	value, err = parser.Parse()
	if err != nil {
		log.Fatal(err)
	}
*/

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(reader io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(reader)}
}

func (p *Parser) Parse() (Value, error) {
	// read the type byte
	respType, erro := p.reader.ReadByte()

	if erro != nil {
		return Value{}, erro
	}

	switch Type(respType) {
	case SimpleString:
		return p.parseSimpleString()
	case Error:
		return p.parseError()
	case Integer:
		return p.parseInteger()
	case BulkString:
		return p.parseBulkString()
	case Array:
		return p.parseArray()
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %c", respType)
	}
}

func (p *Parser) parseSimpleString() (Value, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return Value{}, err
	}

	return SimpleStringVal(strings.Trim(line, "\r\n")), nil
}

func (p *Parser) parseError() (Value, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return Value{}, err
	}

	return ErrorVal(strings.Trim(line, "\r\n")), nil
}

func (p *Parser) parseInteger() (Value, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return Value{}, err
	}

	num, err := strconv.ParseInt(strings.TrimRight(line, "\r\n"), 10, 64)
	if err != nil {
		return Value{}, err
	}

	return IntegerVal(num), nil
}

func (p *Parser) parseBulkString() (Value, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(strings.TrimRight(line, "\r\n"))
	if err != nil {
		return Value{}, err
	}

	if length == -1 {
		return NullBulkStringVal(), nil
	}

	buf := make([]byte, length+2) // +2 for \r\n
	_, err = io.ReadFull(p.reader, buf)
	if err != nil {
		return Value{}, err
	}

	return BulkStringVal(string(buf[:length])), nil
}

func (p *Parser) parseArray() (Value, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(strings.TrimRight(line, "\r\n"))
	if err != nil {
		return Value{}, err
	}

	if length == -1 {
		return NullArrayVal(), nil
	}

	array := make([]Value, length)
	for i := 0; i < length; i++ {
		value, err := p.Parse()
		if err != nil {
			return Value{}, err
		}
		array[i] = value
	}

	return ArrayVal(array), nil
}
