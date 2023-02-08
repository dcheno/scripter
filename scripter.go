package scripter

import (
	"fmt"
	"strings"
)

type lineType int

const (
	reply       lineType = 0
	expectation lineType = 1
)

type line struct {
	message  string
	lineType lineType
}

type tlike interface {
	Error(args ...any)
	Errorf(message string, args ...any)
}

type script struct {
	t          tlike
	lines      []line
	pos        int
	readBuffer *strings.Reader
}

func (s *script) AssertFinished() {
	if s.pos < len(s.lines) {
		s.t.Errorf("Script incomplete! Only reached: %v", s.lines[s.pos])
	}
}

func NewScript(t tlike, lines ...line) *script {
	if len(lines) == 0 {
		panic("scripter must be created with lines")
	}

	script := &script{
		t,
		lines,
		0,
		strings.NewReader(""),
	}

	script.moveToLine(0)
	return script
}

func Expect(message string) line {
	return line{
		message,
		expectation,
	}
}

func Reply(message string) line {
	return line{
		message,
		reply,
	}
}

func (s *script) moveToLine(n int) { // TODO: change to increment, start at -1
	if n == len(s.lines) {
		s.pos = n
		return
	}

	if n > len(s.lines) {
		panic("unexpected error in scripter package! please report")
	}

	if s.lines[n].lineType == reply {
		s.readBuffer.Reset(s.lines[n].message)
	}

	s.pos = n
}

func (s *script) Write(p []byte) (int, error) {
	written := string(p)
	if s.pos >= len(s.lines) {
		s.t.Errorf("Tried to write after the end of the script! Wrote: \"%s\"", written)
		return 0, nil
	}

	currentLine := s.lines[s.pos]

	if currentLine.lineType == reply {
		s.t.Errorf("Tried to write off script! Wrote: \"%s\". Current Line: %s", written, currentLine)
	} else if written != currentLine.message {
		s.t.Errorf("Unexpected output: wrote [%s], expected [%s]", written, currentLine.message)
	} else {
		s.moveToLine(s.pos + 1)
	}
	return len(p), nil
}

func (s *script) Read(p []byte) (int, error) {
	if s.pos >= len(s.lines) {
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
			s.moveToLine(s.pos + 1)
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
