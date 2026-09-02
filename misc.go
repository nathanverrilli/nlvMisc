package nlvMisc

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"os/user"
	"reflect"
	"runtime"
	"slices"
	"strings"
)

// DATE_OCPI time format for DateTime 2015-06-29T20:39:09
// Jan 2 15:04:05 2006 MST
// const DATE_OCPI = "2006-01-02T15:04:05"
const DATE_YYMMDD string = "060102"

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

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

// GetFunctionName returns the full name of the
// given function as a string by using reflection
// and runtime package.
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// GetCallerFunctionName returns the name of the
// calling function's caller as a string.
// An optional skipcount argument determines the
// stack frame to inspect.
func GetCallerFunctionName(skipcount ...int) string {
	var skip int = 2
	if len(skipcount) > 0 {
		skip = skipcount[0]
	}
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	if nil == fn {
		return ""
	}
	return fn.Name()
}

// MapToKeySlice extracts and returns all keys from a map as a slice.
func MapToKeySlice[T comparable, S any](key map[T]S) (keyList []T) {
	keyList = make([]T, 0, len(key))
	for k := range key {
		keyList = append(keyList, k)
	}
	return keyList
}

// MapSortKeys returns the sorted keys of a map as a slice. Handles case-insensitive sorting for string keys.
func MapSortKeys[S cmp.Ordered, T any](mk map[S]T) []S {
	if len(mk) == 0 {
		return []S{}
	}

	keys := MapToKeySlice(mk)

	// Check if S is string or a type derived from string
	if reflect.TypeOf(*new(S)).Kind() == reflect.String {
		slices.SortFunc(keys, CompareStringWithoutCase)
	} else {
		slices.Sort(keys)
	}
	return keys
}

// func StringSliceSortInsensitive compares two ordered
// elements a and b case-insensitively.
// It returns -1 if a < b, 1 if a > b, and 0 if they are equal,
// based on normalized lowercase comparison.
func CompareStringWithoutCase[S cmp.Ordered](a, b S) int {
	sa := reflect.ValueOf(a).String()
	sb := reflect.ValueOf(b).String()

	la := strings.ToLower(sa)
	lb := strings.ToLower(sb)
	if la != lb {
		if la < lb {
			return -1
		}
		return 1
	}
	// Tie-break with original case for deterministic output
	if sa < sb {
		return -1
	}
	if sa > sb {
		return 1
	}
	return 0
}
