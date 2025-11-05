package nlvMisc

import (
	"fmt"
	"os"
	"sync"
)

var optMutex sync.Mutex // thread-safe these variables
var flagDebug = false
var flagVerbose = false
var miscPrintf = defaultPrintf
var miscFatal = defaultFatal
var defaultOutdir = ".output"
var defaultCvsSep = '\t'

func XPrintf(format string, a ...interface{}) (n int, err error) {
	return miscPrintf(format, a...)
}

func IsDebug() bool   { return flagDebug }
func IsVerbose() bool { return flagVerbose }

// OptionOutputDir sets the output directory and returns the old value of the output directory.
// Caller is responsible for ensuring this directory exists, even if the caller is using the
// default value -- this module will not create directory.
func OptionOutputDir(outdir string) (old string) {
	optMutex.Lock()
	defer optMutex.Unlock()
	old, defaultOutdir = defaultOutdir, outdir
	return old
}

// OptionDebug sets the debug flag to the specified value and returns the old value of the debug flag.
func OptionDebug(debug bool) (old bool) {
	optMutex.Lock()
	defer optMutex.Unlock()
	old, flagDebug = flagDebug, debug
	return old
}

// OptionVerbose sets the verbose flag to the specified value, and returns the old value of the verbose flag.
func OptionVerbose(verbose bool) (old bool) {
	optMutex.Lock()
	defer optMutex.Unlock()
	old, flagVerbose = flagVerbose, verbose
	return old
}

// OptionPrintf allows you to set a custom printf function for logging.
// It takes a function `f` with the same signature as the default `printf`
// function and returns the old `printf` function.
func OptionPrintf(f func(format string, a ...interface{}) (n int, err error)) (old func(format string, a ...interface{}) (n int, err error)) {
	optMutex.Lock()
	defer optMutex.Unlock()
	old, miscPrintf = miscPrintf, f
	return old
}

// defaultPrintf writes a formatted string to stderr. It returns the number of bytes
// written and any write error encountered. It is a default print function, probably
// overwritten by SafeLogPrint or xlog.Printf
func defaultPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, format, a...)
}

// OptionFatal sets a custom fatal handler function and returns the previous handler function.
// The provided function is used to handle fatal errors, optionally with exit codes.
func OptionFatal(f func(retcode ...int)) (old func(retcode ...int)) {
	optMutex.Lock()
	defer optMutex.Unlock()
	old, miscFatal = miscFatal, f
	return old
}

// defaultFatal terminates the program after executing cleanup functions in FinishClose,
// then exits with the provided code, if provided. Ideally the larger function provides
// a custom fatal method to close everything cleanly, but if not, there's always
// defaultFatal as a fallback
func defaultFatal(retcode ...int) {
	rc := 0
	if len(retcode) > 0 {
		rc = retcode[0]
	}
	FinishClose()
	os.Exit(rc)
}
