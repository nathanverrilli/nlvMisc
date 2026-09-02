package nlvMisc

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
)

// TestRecordString tests RecordString writing string channels to files,
// verifying file content, completion callback execution, and filename extension handling.
func TestRecordString(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := OptionOutputDir(tmpDir)
	defer OptionOutputDir(oldDir)

	t.Run("BasicWrite", func(t *testing.T) {
		inTx := make(chan string, 2)
		allDone := make(chan struct{})
		doneFunc := func() { close(allDone) }

		inTx <- "line 1"
		inTx <- "line 2"
		close(inTx)

		RecordString("test.txt", inTx, doneFunc)
		<-allDone

		content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
		if err != nil {
			t.Fatal(err)
		}
		expected := "line 1\r\nline 2\r\n"
		if string(content) != expected {
			t.Errorf("Expected %q, got %q", expected, string(content))
		}
	})

	t.Run("ExtensionHandling", func(t *testing.T) {
		inTx := make(chan string)
		close(inTx)
		RecordString("my.log", inTx, func() {})
		// New behavior: stays my.log if extension is present
		expectedFile := filepath.Join(tmpDir, "my.log")
		if _, err := os.Stat(expectedFile); err != nil {
			t.Errorf("Expected %s to exist, got error: %v", expectedFile, err)
		}

		RecordString("noname", inTx, func() {})
		// New behavior: appends .txt if no extension
		expectedFile2 := filepath.Join(tmpDir, "noname.txt")
		if _, err := os.Stat(expectedFile2); err != nil {
			t.Errorf("Expected %s to exist, got error: %v", expectedFile2, err)
		}
	})
}

// TestRecordCsv tests RecordCsv writing CSV records from a channel to a file,
// verifying separator configuration, record formatting, and completion callback execution.
func TestRecordCsv(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := OptionOutputDir(tmpDir)
	defer OptionOutputDir(oldDir)

	inTx := make(chan []string, 2)
	allDone := make(chan struct{})
	doneFunc := func() { close(allDone) }

	inTx <- []string{"a", "b"}
	inTx <- []string{"c", "d"}
	close(inTx)

	RecordCsv("test.csv", inTx, doneFunc)
	<-allDone

	f, err := os.Open(filepath.Join(tmpDir, "test.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = getCsvSep()
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}
	if records[0][0] != "a" || records[1][1] != "d" {
		t.Errorf("Unexpected CSV content: %v", records)
	}
}

// TestRecordBytes tests RecordBytes writing byte slices from a channel to a file,
// verifying binary content output and completion callback execution.
func TestRecordBytes(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := OptionOutputDir(tmpDir)
	defer OptionOutputDir(oldDir)

	inTx := make(chan []byte, 2)
	allDone := make(chan struct{})
	doneFunc := func() { close(allDone) }

	inTx <- []byte{0x01, 0x02}
	inTx <- []byte{0x03, 0x04}
	close(inTx)

	RecordBytes("data.bin", inTx, doneFunc)
	<-allDone

	content, err := os.ReadFile(filepath.Join(tmpDir, "data.bin"))
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x01, 0x02, 0x03, 0x04}
	if string(content) != string(expected) {
		t.Errorf("Expected %v, got %v", expected, content)
	}
}
