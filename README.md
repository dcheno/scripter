# Scripter

A testing library for Go command line tools. A `Script` is created from the perspective
of the command line user. The script is built of lines. Each line is either an `Expected`ed
output from the code under test, or a `Reply` which the script will provide to the code
when asked.

Example usage:

```go
// dogcli.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	DogCli(os.Stdin, os.Stdout)
}

func DogCli(in io.Reader, out io.Writer) {
	fmt.Fprint(out, "Are you a dog? (y/n)\n")

	scanner := bufio.NewScanner(in)
	scanner.Scan()

	if scanner.Text() == "y" {
		fmt.Fprint(out, "woof woof woof")
	} else {
		fmt.Fprint(out, "uh, meow?")
	}
}
```

The following tests will pass:
```go
// dogcli_test.go
package main

import (
	"github.com/dcheno/scripter"
	"testing"
)

func TestDogIsGreetedWithWoofing(t *testing.T) {
	script := scripter.NewScript(
		t,
		scripter.Expect("Are you a dog? (y/n)\n"),
		scripter.Reply("y\n"),
		scripter.Expect("woof woof woof"),
	)

	DogCli(script.In(), script.Out())
	script.AssertFinished()
}

func TestNonDogIsNotGreetedWithWoofing(t *testing.T) {
	script := scripter.NewScript(
		t,
		scripter.Expect("Are you a dog? (y/n)\n"),
		scripter.Reply("n\n"),
		scripter.Expect("uh, meow?"),
	)

	DogCli(script.In(), script.Out())
	script.AssertFinished()
}
```

Where the following test will fail:
```go
func TestDogIsGreetedWithBarking(t *testing.T) {
	script := scripter.NewScript(
		t,
		scripter.Expect("Are you a dog? (y/n)\n"),
		scripter.Reply("y\n"),
		scripter.Expect("bark bark bark"),
	)

	DogCli(script.In(), script.Out())
	script.AssertFinished()
}
```

with message:
```
--- FAIL: TestDogIsGreetedWithBarking (0.00s)
    scripter.go:119: Unexpected output: wrote [woof woof woof], expected [bark bark bark]
```
