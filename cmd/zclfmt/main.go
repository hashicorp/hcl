package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/zclconf/go-zcl/zclwrite"
)

const versionStr = "0.0.1-dev"

var (
	overwrite   = flag.Bool("w", false, "overwrite source files instead of writing to stdout")
	showVersion = flag.Bool("version", false, "show the version number and immediately exit")
)

func main() {
	if err := realmain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realmain() error {
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(versionStr)
		return nil
	}

	if flag.NArg() == 0 {
		if *overwrite {
			return errors.New("error: cannot use -w without source filenames")
		}

		return processFile("<stdin>", os.Stdin)
	}

	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			return err
		case dir.IsDir():
			// This tool can't walk a whole directory because it doesn't
			// know what file naming schemes will be used by different
			// zcl-embedding applications, so it'll leave that sort of
			// functionality for apps themselves to implement.
			return fmt.Errorf("can't format directory %s", path)
		default:
			if err := processFile(path, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func processFile(fn string, in *os.File) error {
	var err error
	if in == nil {
		in, err = os.Open(fn)
		if err != nil {
			return fmt.Errorf("failed to open %s: %s", fn, err)
		}
	}

	inSrc, err := ioutil.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read %s: %s", fn, err)
	}

	outSrc := zclwrite.Format(inSrc)

	if *overwrite {
		return ioutil.WriteFile(fn, outSrc, 0644)
	}

	_, err = os.Stdout.Write(outSrc)
	return err
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: zclfmt [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
