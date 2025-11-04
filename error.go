package misc

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"
)

const CLOSE_BUFFER_SIZE = 4

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
	atClose = append(atClose, f)
	atCloseName = append(atCloseName, GetFunctionName(f))
}

// AtClose registers a function to be called upon program termination.
// Functions are run in the reverse order they are registered.
func AtClose(f func()) {
	atClose = append(atClose, func() error { f(); return nil })
	atCloseName = append(atCloseName, GetFunctionName(f))
}

// FinishClose runs all functions in the atClose slice
// __in reverse order__. Logs function names and errors if
// flagDebug or flagVerbose is set.
func FinishClose() {
	var err error
	if flagDebug {
		_, _ = miscPrintf("Number of AtClose/AtCloseErr functions is %d (started with capacity %d)\n",
			len(atClose), CLOSE_BUFFER_SIZE)
	}
	for ix, fn := range slices.Backward(atClose) {
		err = fn()
		/* if flagDebug || flagVerbose {
			_, _ = printf("AtClose running function %s\n",
				atCloseName[ix])
		} */
		if nil != err {
			_, _ = miscPrintf("AtClose function %s failed because %s\n",
				atCloseName[ix], err.Error())
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
// returns an error at its close
func DeferError(f func() error) {
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
