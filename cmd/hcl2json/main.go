package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/hcl"
)

func main() {
	cli := &CLI{
		Args:   os.Args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	os.Exit(cli.Run())
}

type CLI struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (c *CLI) Run() int {
	if 2 != len(c.Args) {
		return c.Usage()
	}

	file := c.Args[1]

	var d []byte
	var err error

	if "-h" == file {
		return c.Usage()
	} else if "-" == file {
		d, err = ioutil.ReadAll(c.Stdin)
	} else {
		d, err = ioutil.ReadFile(file)
	}

	if err != nil {
		log.Fatalf("err: %s", err)
	}

	var obj interface{}
	err = hcl.Decode(&obj, string(d))
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	res, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	fmt.Fprintln(c.Stdout, string(res))

	return 0
}

func (c *CLI) Usage() int {
	fmt.Fprintln(c.Stderr, "usage: hcl2json <file>")
	return 1
}
