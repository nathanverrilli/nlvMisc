# nlvMisc

This collection of tooling functions are a general collection of functions useful for error handling, often-needed string functions, setup/shutdown, and other utility functions.

## chan.go &mdash; MultiChan Implementation
This implementation provides a multi-channel structure that allows for efficient splitting of input &mdash; messages sent to a multi-channel are sent to all channels. This generic implementation has methods to create and add channels, and to close all channels. 

## error.go &mdash; Error Handling
When things go wrong, cleanup and reporting is critical. 
* atClose functionality: run functions (cleanup) at program exit, so not all cleanup needs to be deferred from main()
* DeferError &mdash; defer a function that returns an error, and report the error if it occurs in a one-line function rather than a clunky multiline anonymous closure. It&rsquo;s a personal preference.
  * **todo:** add test in debug mode to ensure DeferError was called with `defer`.
* HandleSignal &mdash; Catch kill signals, and exit the program cleanly (that is, run the registered cleanup functions.
  * **todo:** shut off the verbose output when not in verbose/debug.


## options.go
Tie the misc module to the main program, if desired. Not required for use, sane defaults are provided.
* Enable debug &amp; verbose mode with the module
* Register a fatal() function to terminate the program on disastrous error. Default is to run the cleanup functions and then exit.
* Register a default printf for logging
* Register a default output directory

## string.go
Handles the common cases of CSV output and string output to a file in some specific directory (default `/.output`)
* RecordCSV &mdash; write a CSV record to a file
* RecordString &mdash; write a string to a file. Adds a newline.

## misc.go
Miscellaneous functions
* ConcatanateErrors &mdash; wrap a set of Error objects into a single error object
* GetFunctionName &mdash; get the name of the calling function
* IsStringSet &mdash; check if a string is set (not empty and not `nil`)
* SafeString &mdash; returns a valid string object (empty string if nothing else works)
* UserHostInfo &mdash; get information about the current user and computer hosting the program



