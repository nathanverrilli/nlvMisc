package nlvMisc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// PrettyPrintJSON streams a JSON document from r to w, pretty-printing it
// token by token without ever holding the full document in memory.
// io should be buffered, but above this level
func PrettyPrintJSON(r io.Reader, w io.Writer) (err error) {
	dec := json.NewDecoder(r)
	if err := writeValue(w, dec, 0); err != nil {
		return err
	}
	writeDirect(w, []byte("\n"))
	return err
}

// writeValue reads and writes one complete JSON value at the current depth.
func writeValue(w io.Writer, dec *json.Decoder, depth int) error {
	tok, err := dec.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.EOF
		}
		return fmt.Errorf("reading token: %w", err)
	}

	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			return writeObject(w, dec, depth)
		case '[':
			return writeArray(w, dec, depth)
		default:
			return fmt.Errorf("unexpected delimiter: %v", v)
		}
	default:
		return writeScalar(w, tok)
	}
}

func writeObject(w io.Writer, dec *json.Decoder, depth int) error {
	if !dec.More() {
		// Empty object: consume closing '}' and output inline "{}"
		if _, err := dec.Token(); err != nil {
			return fmt.Errorf("reading object close: %w", err)
		}
		writeDirect(w, []byte("{}"))
		return nil
	}

	writeDirect(w, []byte("{\n"))
	first := true
	for dec.More() {
		if !first {
			writeDirect(w, []byte(",\n"))
		}
		first = false

		// key
		keyTok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("reading object key: %w", err)
		}
		writeIndent(w, depth+1)
		key, _ := keyTok.(string)
		writeDirect(w, []byte(strconv.Quote(key)))
		writeDirect(w, []byte(": "))

		// value
		if err := writeValue(w, dec, depth+1); err != nil {
			return err
		}
	}

	// consume closing '}'
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("reading object close: %w", err)
	}
	writeDirect(w, []byte("\n"))
	writeIndent(w, depth)
	writeDirect(w, []byte("}"))
	return nil
}

func writeArray(w io.Writer, dec *json.Decoder, depth int) error {
	if !dec.More() {
		// Empty array: consume closing ']' and output inline "[]"
		if _, err := dec.Token(); err != nil {
			return fmt.Errorf("reading array close: %w", err)
		}
		writeDirect(w, []byte("[]"))
		return nil
	}

	writeDirect(w, []byte("[\n"))
	first := true
	for dec.More() {
		if !first {
			writeDirect(w, []byte(",\n"))
		}
		first = false
		writeIndent(w, depth+1)
		if err := writeValue(w, dec, depth+1); err != nil {
			return err
		}
	}

	// consume closing ']'
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("reading array close: %w", err)
	}
	writeDirect(w, []byte("\n"))
	writeIndent(w, depth)
	writeDirect(w, []byte("]"))
	return nil
}

func writeScalar(w io.Writer, tok json.Token) error {
	switch v := tok.(type) {
	case nil:
		writeDirect(w, []byte("null"))
	case bool:
		if v {
			writeDirect(w, []byte("true"))
		} else {
			writeDirect(w, []byte("false"))
		}
	case float64:
		writeDirect(w, []byte(strconv.FormatFloat(v, 'f', -1, 64)))
	case json.Number:
		writeDirect(w, []byte(v.String()))
	case string:
		writeDirect(w, []byte(strconv.Quote(v)))
	default:
		return fmt.Errorf("unknown token type: %T", tok)
	}
	return nil
}

func writeIndent(w io.Writer, depth int) {
	for range depth {
		writeDirect(w, []byte(optIndent))
	}
}
