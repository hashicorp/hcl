package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/hcl"
)

func main() {
	args := os.Args

	if 2 != len(args) {
		usage()
	}

	file := args[1]

	var d []byte
	var err error

	if "-h" == file {
		usage()
	} else if "-" == file {
		d, err = ioutil.ReadAll(os.Stdin)
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

	out, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	fmt.Println(string(out))
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: hcl2json <file>")
	os.Exit(1)
}
