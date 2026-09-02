package nlvMisc

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestConcatenateErrors tests ConcatenateErrors for various error list inputs,
// verifying empty lists, all-nil errors, single errors, multiple errors formatting,
// and number alignment across double-digit error counts.
func TestConcatenateErrors(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		if err := ConcatenateErrors(); err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})

	t.Run("AllNil", func(t *testing.T) {
		if err := ConcatenateErrors(nil, nil); err != nil {
			t.Errorf("Expected nil, got %v", err)
		}
	})

	t.Run("SingleError", func(t *testing.T) {
		err1 := errors.New("error 1")
		err := ConcatenateErrors(err1)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		expected := "\n1.\terror 1"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("MultipleErrors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		err := ConcatenateErrors(err1, nil, err2)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		// len(errList) is 3. strconv.Itoa(3) is "3", len is 1.
		// fmtString is "\n% 1d.\t%s"
		expected := "\n1.\terror 1\n2.\terror 2"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Alignment", func(t *testing.T) {
		var errs []error
		for i := 0; i < 10; i++ {
			errs = append(errs, errors.New(fmt.Sprintf("e%d", i+1)))
		}
		err := ConcatenateErrors(errs...)
		// len(errList) is 10. width is 2.
		// fmtString is "\n%2d.\t%s"
		s := err.Error()
		if !strings.Contains(s, "\n 1.\te1") {
			t.Errorf("Expected alignment for 1, got %q", s)
		}
		if !strings.Contains(s, "\n10.\te10") {
			t.Errorf("Expected alignment for 10, got %q", s)
		}
	})
}

// TestAtClose verifies that functions registered with AtClose are executed
// upon calling FinishClose and cleared so they do not execute again.
func TestAtClose(t *testing.T) {
	executed := false
	AtClose(func() {
		executed = true
	})

	FinishClose()

	if !executed {
		t.Error("AtClose function was not executed")
	}

	// Verify it's cleared
	executed = false
	FinishClose()
	if executed {
		t.Error("AtClose function was executed again")
	}
}

// TestAtCloseRecursive verifies that AtClose callbacks registered dynamically
// during FinishClose execution are also invoked properly.
func TestAtCloseRecursive(t *testing.T) {
	innerExecuted := false
	AtClose(func() {
		AtClose(func() {
			innerExecuted = true
		})
	})

	FinishClose()

	if !innerExecuted {
		t.Error("Inner AtClose function was not executed")
	}
}

// TestAtCloseOrder tests that functions registered with AtClose are executed
// in reverse order of registration (LIFO).
func TestAtCloseOrder(t *testing.T) {
	var order []int
	AtClose(func() { order = append(order, 1) })
	AtClose(func() { order = append(order, 2) })

	FinishClose()

	// They should run in reverse order: 2 then 1
	if len(order) < 2 || order[len(order)-2] != 2 || order[len(order)-1] != 1 {
		t.Errorf("Expected order [2, 1], got %v", order)
	}
}

// TestDeferError tests DeferError handling of nil functions, functions that return
// nil error, and functions returning a non-nil error without panicking.
func TestDeferError(t *testing.T) {
	t.Run("NilFunc", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic for DeferError(nil): %v", r)
			}
		}()
		DeferError(nil)
	})

	t.Run("NoError", func(t *testing.T) {
		executed := false
		DeferError(func() error {
			executed = true
			return nil
		})
		if !executed {
			t.Error("Function was not executed")
		}
	})

	t.Run("WithError", func(t *testing.T) {
		// This will log to stderr, we just want to make sure it doesn't panic
		DeferError(func() error {
			return errors.New("test error")
		})
	})
}
