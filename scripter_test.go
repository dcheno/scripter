package scripter_test

import (
	"bufio"
	"fmt"
	"testing"

	"github.com/dcheno/scripter"
)

func TestReportsFailureWithNormalTestingT(t *testing.T) {
	testT := &testing.T{}

	script := scripter.NewScript(
		testT,
		scripter.Expect("hello"),
	)

	fmt.Fprint(script, "goodbye")

	if !testT.Failed() {
		t.Error("expected this test to fail")
	}
}

func TestReportsCleanlyWithNormalTestingT(t *testing.T) {
	testT := &testing.T{}

	script := scripter.NewScript(
		testT,
		scripter.Expect("hello"),
	)

	fmt.Fprint(script, "hello")

	if testT.Failed() {
		t.Error("expected this test to pass")
	}
}

type mockT struct {
	errors []string
}

func (m *mockT) Error(args ...any) {
	error := fmt.Sprint(args...)
	m.errors = append(m.errors, error)
}

func (m *mockT) Errorf(message string, args ...any) {
	error := fmt.Sprintf(message, args...)
	m.errors = append(m.errors, error)
}

// TODO: add utility to separate out Script into In, Out so that
// one can make sure they go to the write file.
func TestScriptHandlesSeveralLines(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("Are you a dog?"),
		scripter.Reply("yes\n"),
		scripter.Expect("woof woof woof"),
	)

	fmt.Fprint(script, "Are you a dog?")

	scanner := bufio.NewScanner(script)

	scanner.Scan()
	if scanner.Text() == "yes" {
		fmt.Fprint(script, "woof woof woof")
	} else {
		t.Error("Got unexpected return:", scanner.Text())
	}

	script.AssertFinished()

	if len(testT.errors) != 0 {
		t.Error("unexpected test failure")
	}
}

func TestCatchesWrongWrite(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("the right thing to say"),
	)

	fmt.Fprint(script, "the wrong thing to say")

	if len(testT.errors) == 0 {
		t.Error("expected test failure")
		return
	}

	if testT.errors[0] != "Unexpected output: wrote [the wrong thing to say], expected [the right thing to say]" {
		t.Errorf("incorrect error: %s", testT.errors[0])
	}
}

func TestRespondsWithCorrectReply(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("hello"),
		scripter.Reply("hello to you too\n"),
	)

	fmt.Fprintf(script, "hello")

	scanner := bufio.NewScanner(script)
	scanner.Scan()
	if scanner.Text() != "hello to you too" {
		t.Error("did not reply with expected text")
	}
}

func TestCatchesMissingExpectation(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("hello"),
	)

	script.AssertFinished()

	if len(testT.errors) == 0 {
		t.Error("Should have reported script as unifinished but did not.")
	}

	if testT.errors[0] != "Script incomplete! Only reached: {Expect: \"hello\"}" {
		t.Errorf("incorrect error: %s", testT.errors[0])
	}
}

func TestCatchesMissingExpectationAtEnd(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("out1"),
		scripter.Expect("out2"),
	)

	fmt.Fprint(script, "out1")
	script.AssertFinished()

	if len(testT.errors) == 0 {
		t.Error("Should have reported script as unifinished but did not.")
	}

	if testT.errors[0] != "Script incomplete! Only reached: {Expect: \"out2\"}" {
		t.Errorf("incorrect error: %s", testT.errors[0])
	}
}

func TestCatchesMissingReply(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Reply("hello"),
	)

	script.AssertFinished()

	if len(testT.errors) == 0 {
		t.Error("Should have reported script as unifinished but did not.")
	}

	if testT.errors[0] != "Script incomplete! Only reached: {Reply with: \"hello\"}" {
		t.Errorf("incorrect error: %s", testT.errors[0])
	}
}

func TestCatchesMissingReplyAtEnd(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("out"),
		scripter.Reply("in"),
	)

	fmt.Fprint(script, "out")
	script.AssertFinished()

	if len(testT.errors) == 0 {
		t.Error("Should have reported script as unifinished but did not.")
	}

	if testT.errors[0] != "Script incomplete! Only reached: {Reply with: \"in\"}" {
		t.Errorf("incorrect error: %s", testT.errors[0])
	}
}

func TestAllowsReadToHappenInMultipleStages(t *testing.T) {
	testT := new(mockT)

	expected := "a 16 byte string"
	script := scripter.NewScript(
		testT,
		scripter.Reply(expected),
	)

	p := make([]byte, 8)
	script.Read(p)

	first := string(p)

	script.Read(p)

	second := string(p)

	if first+second != expected {
		t.Errorf("handled multiple reads incorrectly. expected: %s got %s", expected, first+second)
	}
	script.AssertFinished()

	if len(testT.errors) != 0 {
		t.Error("unexpected failure")
	}
}

func TestHandlesMultipleReadsInARow(t *testing.T) {
	testT := new(mockT)

	expectedFirst := "hello"
	expectedSecond := "hello to you too"
	script := scripter.NewScript(
		testT,
		scripter.Reply(expectedFirst+"\n"),
		scripter.Reply(expectedSecond+"\n"),
	)

	scanner := bufio.NewScanner(script)

	scanner.Scan()
	first := scanner.Text()

	scanner.Scan()
	second := scanner.Text()

	if first != expectedFirst || second != expectedSecond {
		t.Errorf(
			"didn't do both reads properly. expected %s, %s, got %s, %s",
			expectedFirst,
			expectedSecond,
			first,
			second,
		)
	}

	script.AssertFinished()

	if len(testT.errors) != 0 {
		t.Error("unexpected failure")
	}
}

func TestHandlesMutipleWritesInARow(t *testing.T) {
	testT := new(mockT)

	first := "hello"
	second := "hello to you too"
	script := scripter.NewScript(
		testT,
		scripter.Expect(first),
		scripter.Expect(second),
	)

	fmt.Fprint(script, first)
	fmt.Fprint(script, second)

	script.AssertFinished()

	if len(testT.errors) != 0 {
		t.Error("unexpected failure")
	}
}

func TestCatchesReadWhenWriteExpected(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("hello?"),
	)

	scanner := bufio.NewScanner(script)
	scanner.Scan()

	// even if correct write comes afterward
	fmt.Fprint(script, "hello?")

	if len(testT.errors) == 0 {
		t.Error("script should have reported incorrect read")
	}

	if testT.errors[0] != "Tried to read off script! Current Line: {Expect: \"hello?\"}" {
		t.Errorf("unexpected error: %s", testT.errors[0])
	}
}

func TestCatchesWriteWhenReadExpected(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Reply("hello?"),
	)

	fmt.Fprint(script, "hello?")

	// even if correct read comes afterward
	scanner := bufio.NewScanner(script)
	scanner.Scan()

	if len(testT.errors) == 0 {
		t.Error("script should have reported incorrect write")
	}

	if testT.errors[0] != "Tried to write off script! Wrote: \"hello?\". Current Line: {Reply with: \"hello?\"}" {
		t.Errorf("unexpected error: %s", testT.errors[0])
	}
}

func TestCatchesWriteAfterEndOfScript(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("hello?"),
	)

	// should end script
	fmt.Fprint(script, "hello?")

	// so this is extra
	fmt.Fprint(script, "Are you there?")

	if len(testT.errors) == 0 {
		t.Error("script should have reported extra write")
	}

	if testT.errors[0] != "Tried to write after the end of the script! Wrote: \"Are you there?\"" {
		t.Errorf("unexpected error: %s", testT.errors[0])
	}
}

func TestCatchesReadAfterEndOfScript(t *testing.T) {
	testT := new(mockT)

	script := scripter.NewScript(
		testT,
		scripter.Expect("hello?"),
	)

	// should end script
	fmt.Fprint(script, "hello?")

	// so this is extra
	scanner := bufio.NewScanner(script)
	scanner.Scan()

	if len(testT.errors) == 0 {
		t.Error("script should have reported extra read")
	}

	if testT.errors[0] != "Tried to read after the end of the script!" {
		t.Errorf("unexpected error: %s", testT.errors[0])
	}
}
