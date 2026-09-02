package nlvMisc

import (
	"fmt"
	"os"
	"sync"
)

var optMutex sync.Mutex // thread-safe these variables
var flagDebug = false
var flagVerbose = false
var optPrintf = defaultPrintf
var optFatal = defaultFatal
var defaultOutdir = ".output"
var defaultCsvSep = '\t'
var optIndent = "\t"

func OptionIndent(newIndent string) (oldIndent string) {
	optMutex.Lock()
	defer optMutex.Unlock()
	oldIndent = optIndent
	optIndent = newIndent
	return oldIndent
}

// miscPrintf formats and writes output using a format string, protected by a mutex to ensure thread safety.
func miscPrintf(format string, a ...interface{}) (n int, err error) {
	optMutex.Lock()
	f := optPrintf
	optMutex.Unlock()
	return f(format, a...)
}

// miscFatal invokes the configured fatal error function, passing optional return codes, and ensures thread safety.
func miscFatal(retcode ...int) {
	optMutex.Lock()
	fatalReport := optFatal
	defer optMutex.Unlock()
	fatalReport(retcode...)
}

// XPrintf formats and writes log output using a format string and arguments,
// ensuring thread safety via a mutex. It may use a log format from the calling
// program, but defaults to writing stderr.
func XPrintf(format string, a ...interface{}) (n int, err error) {
	return miscPrintf(format, a...)
}

// isDebug checks if the debug mode is enabled by evaluating the `flagDebug` variable with thread safety.
func isDebug() bool {
	optMutex.Lock()
	defer optMutex.Unlock()
	return flagDebug
}

// isVerbose checks if the verbose mode is enabled by evaluating the globally synchronized flagVerbose variable.
func isVerbose() bool {
	optMutex.Lock()
	defer optMutex.Unlock()
	return flagVerbose
}

// OptionOutputDir sets the output directory and returns the old value of the output directory.
// Caller is responsible for ensuring this directory exists, even if the caller is using the
// default value -- this module will not create directory.
func OptionOutputDir(outdir string) (old string) {
	optMutex.Lock()
	defer optMutex.Unlock()
	old, defaultOutdir = defaultOutdir, outdir
	return old
}

// getOutputDir returns the current output directory.
func getOutputDir() string {
	optMutex.Lock()
	defer optMutex.Unlock()
	return defaultOutdir
}

// getCsvSep returns the current CSV field separator.
func getCsvSep() rune {
	optMutex.Lock()
	defer optMutex.Unlock()
	return defaultCsvSep
}

// OptionCsvSep sets the CSV field separator and returns the old value.
func OptionCsvSep(sep rune) (old rune) {
	optMutex.Lock()
	defer optMutex.Unlock()
	old, defaultCsvSep = defaultCsvSep, sep
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
	old = optPrintf
	if f != nil {
		optPrintf = f
	}
	return old
}

// defaultPrintf writes the formatted output to os.Stderr
// and returns the number of bytes written and any write error.
// Can be overridden by OptionPrintf, this is just a reasonable
// default implementation if the larger program fails to provide one.
func defaultPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, format, a...)
}

// OptionFatal sets a custom fatal handler function and returns the previous handler function.
// The provided function is used to handle fatal errors, optionally with exit codes.
func OptionFatal(f func(retcode ...int)) (old func(retcode ...int)) {
	optMutex.Lock()
	defer optMutex.Unlock()
	old = optFatal
	if f != nil {
		optFatal = f
	}
	return old
}

var fatalMutex sync.Mutex

// defaultFatal terminates the program after executing cleanup functions in FinishClose,
// then exits with the provided code, if provided. Ideally the larger function provides
// a custom fatal method to close everything cleanly, but if not, there's always
// defaultFatal as a fallback
// do not permit multiple calls to defaultFatal!
func defaultFatal(retcode ...int) {
	fatalMutex.Lock()
	defer fatalMutex.Unlock()
	rc := 0
	if len(retcode) > 0 {
		rc = retcode[0]
	}
	FinishClose()
	os.Exit(rc)
}
