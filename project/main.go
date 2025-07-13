package main

import (
	"fmt"
	"os"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
