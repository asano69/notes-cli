package main

import (
	"fmt"
	"github.com/asano69/notes-cli"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"os"
)

func exit(err error) {
	if err != nil {
		fmt.Fprintln(colorable.NewColorableStderr(), color.RedString("notes: error:"), err.Error())
		os.Exit(110)
	}
	os.Exit(0)
}

func main() {
	c, err := notes.ParseCmd(os.Args[1:])
	if err != nil {
		exit(err)
	}
	exit(c.Do())
}
