package nlvMisc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const CLOSE_BUFFER_SIZE = 4

var atCloseMutex sync.Mutex
var atClose []func() error
var atCloseName []string

// init initializes the atClose and atCloseName
// slices with a predefined capacity of CLOSE_BUFFER_SIZE.
func init() {
	atClose = make([]func() error, 0, CLOSE_BUFFER_SIZE)
	atCloseName = make([]string, 0, CLOSE_BUFFER_SIZE)
}

// AtCloseErr registers a function that returns an error
// to be called when the application is closing.
func AtCloseErr(f func() error) {
	atCloseMutex.Lock()
	defer atCloseMutex.Unlock()
	atClose = append(atClose, f)
	atCloseName = append(atCloseName, GetFunctionName(f))
}

// AtClose registers a function to be called upon program termination.
// Functions are run in the reverse order they are registered.
func AtClose(f func()) {
	atCloseMutex.Lock()
	defer atCloseMutex.Unlock()
	atClose = append(atClose, func() error { f(); return nil })
	atCloseName = append(atCloseName, GetFunctionName(f))
}

// FinishClose runs all functions in the atClose slice
// __in reverse order__. Logs function names and errors.
// It continues to run as long as new functions are registered
// during the closing process.
func FinishClose() {
	var err error

	for {
		// snapshot under lock, then run the functions unlocked so that a
		// close function is free to register more cleanup (AtClose/AtCloseErr)
		// or trigger a fatal exit without deadlocking on a non-reentrant mutex
		atCloseMutex.Lock()
		if len(atClose) == 0 {
			atCloseMutex.Unlock()
			break
		}
		closeFns := atClose
		closeNames := atCloseName
		atClose = make([]func() error, 0, CLOSE_BUFFER_SIZE)
		atCloseName = make([]string, 0, CLOSE_BUFFER_SIZE)
		atCloseMutex.Unlock()

		if isDebug() {
			_, _ = miscPrintf("Running %d AtClose/AtCloseErr functions (started with capacity %d)\n",
				len(closeFns), CLOSE_BUFFER_SIZE)
		}
		for ix, fn := range slices.Backward(closeFns) {
			err = fn()
			if nil != err {
				_, _ = miscPrintf("AtClose function %s failed because %s\n",
					closeNames[ix], err.Error())
			}
		}
	}
}

// HandleSignal waits for an OS signal from the signalChan channel and
// acts upon receiving the signal, exiting the program. This allows for
// registered at-close routines to execute when the program is killed.
func HandleSignal(signalChan <-chan os.Signal) {
	sig := <-signalChan
	_, _ = miscPrintf("Got signal %v, exiting immediately\n", sig)
	miscFatal(-2)
}

// DeferError accounts for an at-close function that
// returns an error at its close. It handles nil functions gracefully.
func DeferError(f func() error) {
	if f == nil {
		return
	}
	err := f()
	if nil != err {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
			line = 0
		} else {
			file = filepath.Base(file)
		}
		_, _ = miscPrintf("[%s] error in DeferError from file: %s line %d\n"+
			" error: %s\n\t(may be harmless!)",
			time.Now().UTC().Format(time.RFC822),
			file, line, err.Error())
	}
}

func SafeFatal(val ...int) {
	var x = 0
	if len(val) > 0 {
		x = val[0]
	}
	miscFatal(x)
}

// ConcatenateErrors combines a list of errors into a single error, where each
// non-nil error is formatted and included in order.
// Returns nil if all errors in the list are nil.
func ConcatenateErrors(errList ...error) error {
	if nil == errList {
		return nil
	}
	var sb strings.Builder

	if 0 == len(errList) {
		return nil
	}

	width := len(strconv.Itoa(len(errList)))
	fmtString := "\n%" + strconv.Itoa(width) + "d.\t%s"
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

// CheckFatalError handles an error by logging the function
// name and error message, then terminates the program
// execution.
func CheckFatalError(err error) {
	if nil == err {
		return
	}
	fn := GetCallerFunctionName()
	_, _ = miscPrintf("( %s ) Unexpected Fatal Error: %s\n", fn, err.Error())
	miscFatal(-1)
}

// CheckWarningErr logs a warning with the caller function name if
// the passed error is not nil and returns true in that case.
func CheckWarningErr(err error) (isErr bool) {
	isErr = false
	if nil != err {
		fn := GetCallerFunctionName()
		_, _ = miscPrintf("( %s ) Warning: %s\n", fn, err.Error())
		isErr = true
	}
	return isErr
}
