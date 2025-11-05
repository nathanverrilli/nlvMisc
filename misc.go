package nlvMisc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/user"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

// DATE_OCPI time format for DateTime 2015-06-29T20:39:09
// Jan 2 15:04:05 2006 MST
// const DATE_OCPI = "2006-01-02T15:04:05"
const DATE_YYMMDD string = "060102"

// SafeString returns either the pointer to the string,
// or a pointer to the empty string if the string is
// unset
func SafeString(test *string) (safe *string) {
	if IsStringSet(test) {
		return test
	}
	return &emptyString
}

var emptyString = ""

// IsStringSet checks if the provided string pointer is non-nil and points
// to a non-empty string, returning true or false accordingly.
func IsStringSet(s *string) (isSet bool) {
	switch {
	case nil == s:
		fallthrough
	case "" == *s:
		return false
	default:
		break
	}
	return true
}

// UserHostInfo returns the current username, current hostname and an error, as appropriate
func UserHostInfo() (userName string, hostName string, err error) {
	var ui *user.User
	ui, err = user.Current()
	if nil != err {
		return "",
			"",
			errors.New(fmt.Sprintf("UserHostInfo failed to get user.Current() because %s",
				err.Error()))
	}
	hostName, err = os.Hostname()
	if nil != err {
		return ui.Name, "",
			errors.New(fmt.Sprintf("UserHostInfo failed to get os.Hostname() because %s",
				err.Error()))
	}
	return ui.Name, hostName, nil
}

// ConcatenateErrors combines a list of errors into a single error, where each
// non-nil error is formatted and included in order.
// Returns nil if all errors in the list are nil.
func ConcatenateErrors(errList ...error) error {
	if nil == errList {
		return nil
	}
	var sb strings.Builder

	fmtString := "\n% " + strconv.Itoa(int(math.Ceil(math.Log10(float64(len(errList)))))) + "d.\t%s"
	ix := 1
	for _, err := range errList {
		if err == nil {
			continue
		}
		sb.WriteString(fmt.Sprintf(fmtString, ix, err.Error()))
		ix++
	}
	if sb.Len() > 0 {
		return errors.New(sb.String())
	}
	return nil
}

// GetFunctionName returns the full name of the
// given function as a string by using reflection
// and runtime package.
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// PrettifyJson reads JSON data from an input stream, reformats it with
// the specified indentation, and writes it to an output stream. It
// builds readable JSON by adjusting indentation based on braces,
// brackets, and formatting guidelines.
func PrettifyJson(fin io.Reader, fout io.Writer, indent string) (err error) {
	var ci int = 0
	var r rune
	var sz int
	var ind string

	in := bufio.NewReader(fin)
	out := bufio.NewWriter(fout)

	err = consumeWhiteSpace(in)
	if err != nil {
		return err
	}
	for r, sz, err = in.ReadRune(); sz > 0; r, sz, err = in.ReadRune() {
		switch r {
		case '{', '[':
			ci++
			ind = strings.Repeat(indent, ci)
			outRune(out, r)
			outRune(out, '\n')
			outString(out, ind)
			break
		case '}', ']':
			ci--
			ind = strings.Repeat(indent, ci)
			outRune(out, '\n')
			outString(out, ind)
			outRune(out, r)
			outRune(out, '\n')
			outString(out, ind)
			break
		case ',':
			outRune(out, r)
			outRune(out, '\n')
			outString(out, ind)
			break
		case '"':
			str, err := getTheString(in)
			if nil != err {
				return err
			}
			outRune(out, '"')
			outString(out, str)
			outRune(out, '"')
			break
		case ':':
			outRune(out, r)
			outRune(out, ' ')
			break
		default:
			outRune(out, r)
		}
		err := consumeWhiteSpace(in)
		if nil != err && err.Error() != "EOF" {
			return err
		}
	}
	if nil != err && err.Error() != "EOF" {
		return err
	}
	return nil
}

func outString(fout io.Writer, str string) {
	var err error
	_, err = fout.Write([]byte(str))
	if nil != err {
		miscPrintf("error writing string %s because %s", str, err.Error())
	}
}

func outRune(fout io.Writer, r rune) {
	var err error
	_, err = fout.Write([]byte{byte(r)})
	if nil != err {
		miscPrintf("error writing rune %c because %s", r, err.Error())
	}
}

// getTheString reads a string from the provided reader, handling escaped
// characters and stopping at a closing double quote. It returns the
// extracted string and any error encountered during processing. Please
// note that it expects the initial opening double quote to have been
// consumed by the caller, and it consumes the ending double quote.
func getTheString(in *bufio.Reader) (str string, err error) {
	var sb strings.Builder
	for r, sz, err := in.ReadRune(); sz > 0; r, sz, err = in.ReadRune() {
		switch r {
		case '"':
			goto done
		case '\\':
			r, sz, err = in.ReadRune()
			if err != nil {
				return "", err
			} else if sz == 0 {
				return "", errors.New("unexpected EOF")
			}
		default:
			sb.WriteRune(r)
		}
	}
done:
	return sb.String(), nil
}

// consumeWhiteSpace consumes and discards all leading whitespace characters
// from the provided bufio.Reader.
func consumeWhiteSpace(in *bufio.Reader) (err error) {
	var r rune
	var s int

	for r, s, err = in.ReadRune(); s > 0; r, s, err = in.ReadRune() {
		if unicode.IsSpace(r) {
			continue
		}
		in.UnreadRune()
		break
	}
	// this is clumsy, but checking s==0 first would
	// turn any underlying error into 'Unexpected EOF'
	// which is NOT the desired behavior. Blame io.Reader
	// for returning nil error and s==0 for the first
	// requested character that doesn't exist instead of
	// a nice, standard EOF.
	if err != nil {
		return err
	}
	if 0 == s {
		return errors.New("Unexpected EOF")
	}
	return nil
}

// MapToKeys extracts and returns all keys from a map as a slice.
// It works with maps having keys of any comparable type T.
func MapToKeys[T comparable](key map[T]any) (keys []T) {
	keys = make([]T, 0, len(key))
	for k := range key {
		keys = append(keys, k)
	}
	return keys
}
