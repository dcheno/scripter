package scripter

import (
	"fmt"
	"io"
	"strings"
)

type lineType int

const (
	reply       lineType = 0
	expectation lineType = 1
)

// Script keeps track of expected writes and reads to a pair of buffers
// which can be used as test doubles for Stdin and Stdout.
type Script struct {
	t          tlike
	lines      []line
	pos        int
	readBuffer *strings.Reader
}

// In returns an io.Reader which can be used as a test double for Stdin.
func (s *Script) In() io.Reader {
	return s
}

// Out returns an io.Writer which can be used as a test double for Stdout.
func (s *Script) Out() io.Writer {
	return s
}

// AssertFinished reports an error on the Script's t object if the script
// has not yet been completed.
func (s *Script) AssertFinished() {
	if !s.Finished() {
		s.t.Errorf("Script incomplete! Only reached: %v", s.lines[s.pos])
	}
}

// Finished returns true if the script has been completed.
func (s *Script) Finished() bool {
	return s.pos >= len(s.lines)
}

// NewScript creates a script which reports to the provided t object
// when writes and reads do not follow the lines provided.
func NewScript(t tlike, lines ...line) *Script {
	if len(lines) == 0 {
		panic("scripter must be created with lines")
	}

	script := &Script{
		t,
		lines,
		-1,
		strings.NewReader(""),
	}

	script.advance()
	return script
}

// Expect provides a line to a script which represents an expected write.
func Expect(message string) line {
	return line{
		message,
		expectation,
	}
}

// Reply sets the script up to reply with the specified message at a point
// in the script.
func Reply(message string) line {
	return line{
		message,
		reply,
	}
}

type line struct {
	message  string
	lineType lineType
}

type tlike interface {
	Error(args ...any)
	Errorf(message string, args ...any)
}

func (s *Script) advance() {
	s.pos++

	if s.Finished() {
		return
	}

	if s.lines[s.pos].lineType == reply {
		s.readBuffer.Reset(s.lines[s.pos].message)
	}
}

// Write implements the io.Writer interface for scripter (and for scripter.Out()) allowing
// it to be used as a test double for Stdout.
func (s *Script) Write(p []byte) (int, error) {
	written := string(p)
	if s.Finished() {
		s.t.Errorf("Tried to write after the end of the script! Wrote: \"%s\"", written)
		return 0, nil
	}

	currentLine := s.lines[s.pos]

	if currentLine.lineType == reply {
		s.t.Errorf("Tried to write off script! Wrote: \"%s\". Current Line: %s", written, currentLine)
	} else if written != currentLine.message {
		s.t.Errorf("Unexpected output: wrote [%s], expected [%s]", written, currentLine.message)
	} else {
		s.advance()
	}
	return len(p), nil
}

// Read implements the io.Reader interface for scripter (and for scripter.In()) allowing
// it to be used as a test double for Stdin
func (s *Script) Read(p []byte) (int, error) {
	if s.Finished() {
		s.t.Error("Tried to read after the end of the script!")
		return 0, nil
	}

	currentLine := s.lines[s.pos]
	if currentLine.lineType != reply {
		s.t.Errorf("Tried to read off script! Current Line: %s", currentLine)
		return 0, nil
	} else {
		read, error := s.readBuffer.Read(p)
		if s.readBuffer.Len() == 0 {
			s.advance()
		}
		return read, error
	}
}

func (l line) String() string {
	if l.lineType == reply {
		return fmt.Sprintf("{Reply with: \"%s\"}", l.message)
	} else {
		return fmt.Sprintf("{Expect: \"%s\"}", l.message)
	}
}
