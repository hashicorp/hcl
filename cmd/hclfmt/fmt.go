package main

import (
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/fatih/hcl/printer"
)

var (
	write = flag.Bool("w", false, "write result to (source) file instead of stdout")

	// debugging
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to this file")
)

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain() error {

	flag.Usage = usage
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			return fmt.Errorf("creating cpu profile: %s\n", err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flag.NArg() == 0 {
		if *write {
			return errors.New("error: cannot use -w with standard input")
		}

		return processFile("<standard input>", os.Stdin, os.Stdout, true)
	}

	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			report(err)
		case dir.IsDir():
			walkDir(path)
		default:
			if err := processFile(path, nil, os.Stdout, false); err != nil {
				report(err)
			}
		}
	}

	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: hclfmt [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func report(err error) {
	scanner.PrintError(os.Stderr, err)
}

func isHclFile(f os.FileInfo) bool {
	// ignore non-hcl files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".hcl")
}

func walkDir(path string) {
	filepath.Walk(path, visitFile)
}

func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil && isHclFile(f) {
		err = processFile(path, nil, os.Stdout, false)
	}
	if err != nil {
		report(err)
	}
	return nil
}

// If in == nil, the source is the contents of the file with the given filename.
func processFile(filename string, in io.Reader, out io.Writer, stdin bool) error {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	res, err := printer.Format(src)
	if err != nil {
		return err
	}

	if *write {
		err = ioutil.WriteFile(filename, res, 0644)
		if err != nil {
			return err
		}
	}

	_, err = out.Write(res)
	return err
}
