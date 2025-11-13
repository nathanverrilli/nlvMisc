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
// The output file's extension is forced to be `.txt`.
// If writing fails, it logs the error and triggers a fatal exit.
// allDone() is intended to be a sync.WaitGroup.Done().
func RecordString(outFileName string, inTx <-chan string, allDone func()) {
	now := time.Now()
	defer allDone()
	extension := path.Ext(outFileName)
	if !strings.EqualFold(extension, ".txt") {
		outFileName += ".txt"
	}
	fullFileName := path.Join(defaultOutdir, outFileName)

	bout, err := os.OpenFile(fullFileName,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = defaultPrintf("Failed to open %s because %s\n",
			fullFileName, err.Error())
		defaultFatal()
	}
	defer DeferError(bout.Close)
	if flagDebug {
		_, _ = defaultPrintf("started output to file %s\n", outFileName)
	}
	bw := bufio.NewWriterSize(bout, 1024*8)
	defer DeferError(bw.Flush)

	for val := range inTx {
		_, err = bw.Write([]byte(val))
		if nil != err {
			_, _ = defaultPrintf("failed to write string %s to file %s because %s\n",
				val, outFileName, err.Error())
			defaultFatal()
		}
		_, err = bw.WriteRune('\n')
		if nil != err {
			_, _ = defaultPrintf("failed to write newline following string %s to file %s because %s\n",
				val, outFileName, err.Error())
			defaultFatal()
		}

	}

	if flagDebug {
		_, _ = defaultPrintf("Finished output to file %s || required %f seconds\n",
			fullFileName, time.Since(now).Seconds())
	}
	return
}

// RecordCsv writes CSV records to a file, ensuring proper formatting
// and handling errors during writing and flushing.
// The output file's extension is forced to be `.csv`.
// If writing fails, it logs the error and triggers a fatal exit.
// allDone() is intended to be a sync.WaitGroup.Done().
func RecordCsv(outFileName string, inTx <-chan []string, allDone func()) {
	now := time.Now()
	defer allDone()
	extension := path.Ext(outFileName)
	if !strings.EqualFold(extension, ".csv") {
		outFileName += ".csv"
	}

	fullFileName := path.Join(defaultOutdir, outFileName)

	out, err := os.OpenFile(fullFileName,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = defaultPrintf("Failed to open %s because %s\n",
			fullFileName, err.Error())
		defaultFatal()
	}
	if flagDebug {
		_, _ = defaultPrintf("started output to file %s\n", outFileName)
	}
	// defer DeferError(out.Close)
	bout := bufio.NewWriterSize(out, 1024*4)
	// defer DeferError(bout.Flush)
	csvWriter := csv.NewWriter(bout)
	csvWriter.Comma = defaultCvsSep
	// defer csvWriter.Flush()

	for valSet := range inTx {
		err := csvWriter.Write(valSet)
		if nil != err {
			_, _ = defaultPrintf("Failed to write CSV record %v to file %s because %s\n",
				valSet, fullFileName, err.Error())
			defaultFatal()
		}
	}

	// flush writer
	csvWriter.Flush()
	{
		err = csvWriter.Error()
		if nil != err {
			_, _ = defaultPrintf("Failed to flush CSV filewriter %s because %s\n",
				fullFileName, err.Error())
			defaultFatal()
		}
	}

	// flush buffered writer
	err = bout.Flush()
	{
		if nil != err {
			_, _ = defaultPrintf("Failed to flush CSV filewriter %s because %s\n",
				fullFileName, err.Error())
			defaultFatal()
		}
	}

	// close writer file
	err = out.Close()
	{
		if nil != err {
			_, _ = defaultPrintf("Failed to close file %s because %s\n",
				fullFileName, err.Error())
			defaultFatal()
		}
	}

	if flagDebug {
		_, _ = defaultPrintf("Finished output to file %s || required %f seconds\n",
			fullFileName, time.Since(now).Seconds())
	}
}

func RecordBytes(outFileName string, inTx <-chan []byte, allDone func()) {
	now := time.Now()
	defer allDone()
	extension := path.Ext(outFileName)
	if !IsStringSet(&extension) {
		outFileName += ".log"
	}
	fullFileName := path.Join(defaultOutdir, outFileName)

	bout, err := os.OpenFile(fullFileName,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if nil != err {
		_, _ = defaultPrintf("Failed to open %s because %s\n",
			fullFileName, err.Error())
		defaultFatal()
		return
	}
	defer DeferError(bout.Close)
	bw := bufio.NewWriterSize(bout, 1024*8)
	defer DeferError(bw.Flush)

	if flagDebug {
		_, _ = defaultPrintf("started output to file %s\n", outFileName)
	}

	for val := range inTx {
		_, err = bw.Write(val)
		if nil != err {
			_, _ = defaultPrintf("failed to write string %s to file %s because %s\n",
				val, outFileName, err.Error())
			defaultFatal()
		}
	}

	if flagDebug {
		_, _ = defaultPrintf("Finished output to file %s || required %f seconds\n",
			fullFileName, time.Since(now).Seconds())
	}

}
