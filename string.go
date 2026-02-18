package nlvMisc

import (
	"bufio"
	"encoding/csv"
	"os"
	"path"
	"strings"
	"time"
)

// RecordString writes strings received from the `inTx` channel
// to a specified output file and calls `allDone` upon completion.
// Although generally meant for text files, JSON output is permitted.
// If the file extension is not `.json`, it is set to `.txt`.
// If writing fails, it logs the error and triggers a fatal exit.
// allDone() is intended to be a sync.WaitGroup.Done().
func RecordString(outFileName string, inTx <-chan string, allDone func()) {
	var now time.Time
	if flagVerbose {
		now = time.Now()
	}
	defer allDone()

	switch path.Ext(outFileName) {
	case ".json":
		break
	case ".txt":
		break
	default:
		outFileName += ".txt"
	}

	ffn := path.Join(defaultOutdir, outFileName)

	bout, err := os.OpenFile(ffn,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = defaultPrintf("Failed to open %s because %s\n",
			ffn, err.Error())
		defaultFatal()
	}
	defer DeferError(bout.Close)
	if flagDebug {
		_, _ = defaultPrintf("started output to file %s\n", ffn)
	}
	bw := bufio.NewWriterSize(bout, 1024*8) // disk block size usually multiple of 4K
	defer DeferError(bw.Flush)

	for val := range inTx {
		_, err = bw.Write([]byte(val))
		if nil != err {
			_, _ = defaultPrintf("failed to write string %s to file %s because %s\n",
				val, ffn, err.Error())
			defaultFatal()
		}
		_, err = bw.WriteRune('\n')
		if nil != err {
			_, _ = defaultPrintf("failed to write newline following string %s to file %s because %s\n",
				val, ffn, err.Error())
			defaultFatal()
		}

	}

	if flagVerbose {
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
	defer allDone()
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
	bout := bufio.NewWriterSize(out, 1024*4)
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
	{
		err = csvWriter.Error()
		if nil != err {
			_, _ = defaultPrintf("Failed to flush CSV filewriter %s because %s\n",
				ffn, err.Error())
			defaultFatal()
		}
	}

	// flush buffered writer
	err = bout.Flush()
	{
		if nil != err {
			_, _ = defaultPrintf("Failed to flush CSV filewriter %s because %s\n",
				ffn, err.Error())
			defaultFatal()
		}
	}

	// close writer file
	err = out.Close()
	{
		if nil != err {
			_, _ = defaultPrintf("Failed to close file %s because %s\n",
				ffn, err.Error())
			defaultFatal()
		}
	}

	if flagDebug {
		_, _ = defaultPrintf("Finished output to file %s || required %f seconds\n",
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
	bw := bufio.NewWriterSize(bout, 1024*8)
	defer DeferError(bw.Flush)

	if flagDebug {
		_, _ = defaultPrintf("started output to file %s\n", ffn)
	}

	for val := range inTx {
		_, err = bw.Write(val)
		if nil != err {
			_, _ = defaultPrintf("failed to write string %s to file %s because %s\n",
				val, ffn, err.Error())
			defaultFatal()
		}
	}

	if flagDebug {
		_, _ = defaultPrintf("Finished output to file %s || required %f seconds\n",
			ffn, time.Since(now).Seconds())
	}

}
