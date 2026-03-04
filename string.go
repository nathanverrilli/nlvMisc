package nlvMisc

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

// set to prevent excessive disk block fragmentation
const BIGBUFFSIZE = 1024 * 32

// writeDirect writes the provided byte slice
// `data` directly to the specified `io.Writer`.
// Logs an error and terminates the program if
// the write operation fails.
func writeDirect(out io.Writer, data []byte) {
	_, err := out.Write(data)
	if nil != err {
		_, _ = defaultPrintf("failed to write string %s because %s\n",
			string(data), err.Error())
		defaultFatal()
	}
}

// RecordString writes strings received from the `inTx` channel
// to a specified output file and calls `allDone` upon completion.
// Although generally meant for text files, JSON output is permitted.
// If the file extension is not `.json`, it is set to `.txt`.
// If writing fails, it logs the error and triggers a fatal exit.
// allDone() is intended to be a sync.WaitGroup.Done().
func RecordString(outFileName string, inTx <-chan string, allDone func()) {
	var now time.Time
	if flagDebug {
		now = time.Now()
	}
	defer allDone()

	switch strings.ToLower(path.Ext(outFileName)) {
	case ".json":
		break
	case ".txt":
		break
	default:
		outFileName += ".txt"
	}

	ffn := path.Join(defaultOutdir, outFileName)

	out, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = defaultPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		defaultFatal()
	}
	defer DeferError(out.Close)
	defer DeferError(out.Sync)
	if flagDebug {
		_, _ = defaultPrintf("started output to file %s\n", ffn)
	}
	bw := bufio.NewWriter(out) // disk block size usually multiple of 4K
	defer DeferError(bw.Flush)

	for val := range inTx {
		writeDirect(bw, []byte(val))
		writeDirect(bw, []byte("\n"))
	}

	if flagDebug {
		_, _ = defaultPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}
	return
}

// RecordCsv writes CSV records to a file, ensuring proper formatting
// and handling errors during writing and flushing.
// The output file's extension is forced to be `.csv`.
// If writing fails, it logs the error and triggers a fatal exit.
// Fields are all STRING, should be passed in as an array of string
// allDone() is intended to be a sync.WaitGroup.Done().
func RecordCsv(outFileName string, inTx <-chan []string, allDone func()) {
	now := time.Now()

	extension := path.Ext(outFileName)
	if !strings.EqualFold(extension, ".csv") {
		outFileName += ".csv"
	}

	ffn := path.Join(defaultOutdir, outFileName)
	out, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = defaultPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		defaultFatal()
	}
	if flagDebug {
		_, _ = defaultPrintf("started output to file %s\n", ffn)
	}
	// defer DeferError(out.Close)
	bout := bufio.NewWriterSize(out, BIGBUFFSIZE)
	csvWriter := csv.NewWriter(bout)
	csvWriter.Comma = defaultCvsSep
	csvWriter.UseCRLF = true

	for valSet := range inTx {
		err := csvWriter.Write(valSet)
		if nil != err {
			_, _ = defaultPrintf("Failed to write CSV record %v to file %s because %s\n",
				valSet, ffn, err.Error())
			defaultFatal()
		}
	}

	// flush writer
	csvWriter.Flush()
	err = csvWriter.Error()
	if nil != err {
		_, _ = defaultPrintf("Failed to flush CSV filewriter %s because %s\n",
			ffn, err.Error())
		defaultFatal()
	}

	// flush buffered writer
	err = bout.Flush()
	if nil != err {
		_, _ = defaultPrintf("Failed to flush buffered io for csvwriter %s because %s\n",
			ffn, err.Error())
		defaultFatal()
	}

	// flush to storage
	err = out.Sync()
	if nil != err {
		_, _ = defaultPrintf("Failed to sync file %s because %s\n",
			ffn, err.Error())
		defaultFatal()
	}

	// close writer file
	err = out.Close()
	if nil != err {
		_, _ = defaultPrintf("Failed to close file %s because %s\n",
			ffn, err.Error())
		defaultFatal()
	}

	if flagDebug {
		_, _ = defaultPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}

	// release waiter
	allDone()
}

func RecordBytes(outFileName string, inTx <-chan []byte, allDone func()) {
	now := time.Now()
	defer allDone()
	extension := path.Ext(outFileName)
	if !IsStringSet(&extension) {
		outFileName += ".log"
	}
	ffn := path.Join(defaultOutdir, outFileName)

	bout, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = defaultPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		defaultFatal()
		return
	}
	defer DeferError(bout.Close)
	defer DeferError(bout.Sync)
	bw := bufio.NewWriterSize(bout, BIGBUFFSIZE)
	defer DeferError(bw.Flush)

	if flagDebug {
		_, _ = defaultPrintf("started output to file %s\n", ffn)
	}

	for val := range inTx {
		writeDirect(bw, val)
	}

	if flagDebug {
		_, _ = defaultPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}

}
