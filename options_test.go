package nlvMisc

import (
	"sync"
	"testing"
)

// TestOptions tests setting and retrieving various global options including
// Debug, Verbose, OutputDir, CsvSep, and Printf handlers.
func TestOptions(t *testing.T) {
	t.Run("Debug", func(t *testing.T) {
		old := OptionDebug(true)
		defer OptionDebug(old)
		if !isDebug() {
			t.Error("Expected Debug to be true")
		}
		OptionDebug(false)
		if isDebug() {
			t.Error("Expected Debug to be false")
		}
	})

	t.Run("Verbose", func(t *testing.T) {
		old := OptionVerbose(true)
		defer OptionVerbose(old)
		if !isVerbose() {
			t.Error("Expected Verbose to be true")
		}
		OptionVerbose(false)
		if isVerbose() {
			t.Error("Expected Verbose to be false")
		}
	})

	t.Run("OutputDir", func(t *testing.T) {
		old := OptionOutputDir("newdir")
		defer OptionOutputDir(old)
		if getOutputDir() != "newdir" {
			t.Errorf("Expected output dir to be newdir, got %s", getOutputDir())
		}
	})

	t.Run("CsvSep", func(t *testing.T) {
		old := OptionCsvSep(',')
		defer OptionCsvSep(old)
		if getCsvSep() != ',' {
			t.Errorf("Expected CSV separator to be ',', got %c", getCsvSep())
		}
	})

	t.Run("Printf", func(t *testing.T) {
		var called bool
		customPrintf := func(format string, a ...interface{}) (int, error) {
			called = true
			return 0, nil
		}
		old := OptionPrintf(customPrintf)
		defer OptionPrintf(old)

		_, _ = XPrintf("test")
		if !called {
			t.Error("Expected custom printf to be called")
		}
	})
}

// TestConcurrentOptions tests concurrent reading and writing of options
// across multiple goroutines to ensure thread safety.
func TestConcurrentOptions(t *testing.T) {
	var wg sync.WaitGroup
	iterations := 1000

	wg.Add(4)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			OptionDebug(i%2 == 0)
			isDebug()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			OptionVerbose(i%2 == 0)
			isVerbose()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			OptionPrintf(defaultPrintf)
			_, _ = XPrintf("test")
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			OptionFatal(defaultFatal)
			// Don't call fatal here!
		}
	}()

	wg.Wait()
}
