package nlvMisc

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const BIGBUFFSIZE = 1024 * 128

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
	now := time.Now()
	defer allDone()

	extension := filepath.Ext(outFileName)

	if !IsStringSet(&extension) {
		outFileName += ".txt"
	}

	ffn := filepath.Join(getOutputDir(), outFileName)

	f, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = miscPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		miscFatal()
		return
	}
	defer DeferError(f.Close)
	defer DeferError(f.Sync)

	if isDebug() {
		_, _ = miscPrintf("started output to file %s\n", ffn)
	}
	bw := bufio.NewWriterSize(f, BIGBUFFSIZE) // disk block size usually multiple of 4K
	defer DeferError(bw.Flush)

	if ".json" == strings.ToLower(extension) {
		var wg sync.WaitGroup
		wg.Add(1)
		go recordJson(inTx, bw, wg.Done)
		wg.Wait()
	} else {
		for val := range inTx {
			writeDirect(bw, []byte(val))
			writeDirect(bw, []byte("\r\n"))
		}
	}

	if isVerbose() || isDebug() {
		_, _ = miscPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}
}

func recordJson(inTx <-chan string, bw *bufio.Writer, allDone func()) {

	var wg sync.WaitGroup
	defer allDone()
	pipeIn, pipeOut := io.Pipe()

	wg.Go(func() {
		err := PrettyPrintJSON(pipeIn, bw)
		if nil != err {
			_, _ = miscPrintf("pipeIn: Failed to read data to json formatting pipe because %s\n",
				err.Error())
			miscFatal()
		}
	})
	for val := range inTx {
		_, err := pipeOut.Write([]byte(val))
		if nil != err {
			_, _ = miscPrintf("pipeOut: Failed to write to json formatting pipe because %s\n",
				err.Error())
			miscFatal()
		}
	}

	_ = pipeOut.Close()
	wg.Wait()
	_ = pipeIn.Close()
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
	extension := filepath.Ext(outFileName)
	if !strings.EqualFold(extension, ".csv") {
		outFileName += ".csv"
	}

	ffn := filepath.Join(getOutputDir(), outFileName)
	f, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = miscPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		miscFatal()
		return
	}
	defer DeferError(f.Close)
	defer DeferError(f.Sync)

	if isDebug() {
		_, _ = miscPrintf("started output to file %s\n", ffn)
	}
	bw := bufio.NewWriterSize(f, BIGBUFFSIZE)
	defer DeferError(bw.Flush)

	csvWriter := csv.NewWriter(bw)
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
	if err := csvWriter.Error(); nil != err {
		_, _ = miscPrintf("Failed to flush CSV filewriter %s because %s\n",
			ffn, err.Error())
		miscFatal()
	}

	if isDebug() {
		_, _ = miscPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}
}

func RecordBytes(outFileName string, inTx <-chan []byte, allDone func()) {
	now := time.Now()
	defer allDone()
	extension := filepath.Ext(outFileName)
	if !IsStringSet(&extension) {
		outFileName += ".log"
	}
	ffn := filepath.Join(getOutputDir(), outFileName)

	f, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = miscPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		miscFatal()
		return
	}
	defer DeferError(f.Close)
	defer DeferError(f.Sync)

	if isDebug() {
		_, _ = miscPrintf("started output to file %s\n", ffn)
	}
	bw := bufio.NewWriterSize(f, BIGBUFFSIZE)
	defer DeferError(bw.Flush)

	for val := range inTx {
		writeDirect(bw, val)
	}

	if isDebug() {
		_, _ = miscPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}
}
