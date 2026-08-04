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
	cnt, err := out.Write(data)
	if nil != err {
		_, _ = miscPrintf("failed to write string %s because %s\n",
			string(data), err.Error())
		miscFatal()
		return
	}
	if cnt != len(data) {
		_, _ = miscPrintf("mismatched byte write, wrote %d but wanted to write %d\n",
			cnt, len(data))
		miscFatal()
		return
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
	if IsDebug() {
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

	ffn := path.Join(getOutputDir(), outFileName)

	out, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = miscPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		miscFatal()
		return
	}
	defer DeferError(out.Close)
	defer DeferError(out.Sync)

	if IsDebug() {
		_, _ = miscPrintf("started output to file %s\n", ffn)
	}
	bw := bufio.NewWriterSize(out, BIGBUFFSIZE) // disk block size usually multiple of 4K
	defer DeferError(bw.Flush)

	for val := range inTx {
		writeDirect(bw, []byte(val))
		writeDirect(bw, []byte("\n"))
	}

	if IsDebug() {
		_, _ = miscPrintf("Finished output to file %s || required %f seconds\n",
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
	defer allDone()
	extension := path.Ext(outFileName)
	if !strings.EqualFold(extension, ".csv") {
		outFileName += ".csv"
	}

	ffn := path.Join(getOutputDir(), outFileName)
	out, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = miscPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		miscFatal()
		return
	}
	if IsDebug() {
		_, _ = miscPrintf("started output to file %s\n", ffn)
	}
	bout := bufio.NewWriterSize(out, BIGBUFFSIZE)
	csvWriter := csv.NewWriter(bout)
	csvWriter.Comma = getCsvSep()
	csvWriter.UseCRLF = true

	for valSet := range inTx {
		err := csvWriter.Write(valSet)
		if nil != err {
			_, _ = miscPrintf("Failed to write CSV record %v to file %s because %s\n",
				valSet, ffn, err.Error())
			miscFatal()
			return
		}
	}

	// flush writer
	csvWriter.Flush()
	err = csvWriter.Error()
	if nil != err {
		_, _ = miscPrintf("Failed to flush CSV filewriter %s because %s\n",
			ffn, err.Error())
		miscFatal()
	}

	// flush buffered writer
	err = bout.Flush()
	if nil != err {
		_, _ = miscPrintf("Failed to flush buffered io for csvwriter %s because %s\n",
			ffn, err.Error())
		miscFatal()
	}

	// flush to storage
	err = out.Sync()
	if nil != err {
		_, _ = miscPrintf("Failed to sync file %s because %s\n",
			ffn, err.Error())
		miscFatal()
	}

	// close writer file
	err = out.Close()
	if nil != err {
		_, _ = miscPrintf("Failed to close file %s because %s\n",
			ffn, err.Error())
		miscFatal()
	}

	if IsDebug() {
		_, _ = miscPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}
}

func RecordBytes(outFileName string, inTx <-chan []byte, allDone func()) {
	now := time.Now()
	defer allDone()
	extension := path.Ext(outFileName)
	if !IsStringSet(&extension) {
		outFileName += ".log"
	}
	ffn := path.Join(getOutputDir(), outFileName)

	bout, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = miscPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		miscFatal()
		return
	}
	defer DeferError(bout.Close)
	defer DeferError(bout.Sync)
	bw := bufio.NewWriterSize(bout, BIGBUFFSIZE)
	defer DeferError(bw.Flush)

	if IsDebug() {
		_, _ = miscPrintf("started output to file %s\n", ffn)
	}

	for val := range inTx {
		writeDirect(bw, val)
	}

	if IsDebug() {
		_, _ = miscPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}
}
