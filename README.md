# Scripter

A testing library for Go command line tools.

Example usage:

```go
// dogcli.go
// TODO: test this code
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
  fmt.Fprint("Are you a dog? (y/n)")

  scanner := bufio.NewScanner(in)
  scanner.Scan()

  if scanner.Text() == "y" {
    fmt.Fprint("woof woof woof")
  } else {
    fmt.Fprint("uh, meow?")
  }
}
```

```go
// dogcli_test.go
package dogcli_test

import (
  "github.com/dcheno/scripter"
  "testing"
)

func TestDogIsGreetedWithWoofing(t *testing.T) {
  // TODO
}

func TestNonDogIsNotGreetedWithWoofing(t *testing.T) {
  // TODO
}
```
