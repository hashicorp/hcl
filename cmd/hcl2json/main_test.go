package main

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func TestCLINoParamsPrintsUsage(t *testing.T) {
	fakeReader := strings.NewReader("")
	fakeOut := new(bytes.Buffer)
	fakeErr := new(bytes.Buffer)

	cli := &CLI{
		Args:   []string{"hcl2json"},
		Stdin:  fakeReader,
		Stdout: fakeOut,
		Stderr: fakeErr,
	}

	cli.Run()

	actual := fakeErr.String()
	expected := "usage: hcl2json <file>\n"

	if actual != expected {
		t.Fatalf("bad:\n%s", actual)
	}

	outStr := fakeOut.String()
	if outStr != "" {
		t.Fatalf("err:\n%s", outStr)
	}
}

func TestCLIDashHPrintsUsage(t *testing.T) {
	fakeReader := strings.NewReader("")
	fakeOut := new(bytes.Buffer)
	fakeErr := new(bytes.Buffer)

	cli := &CLI{
		Args:   []string{"hcl2json", "-h"},
		Stdin:  fakeReader,
		Stdout: fakeOut,
		Stderr: fakeErr,
	}

	cli.Run()

	actual := fakeErr.String()
	expected := "usage: hcl2json <file>\n"

	if actual != expected {
		t.Fatalf("bad:\n%s", actual)
	}

	outStr := fakeOut.String()
	if outStr != "" {
		t.Fatalf("err:\n%s", outStr)
	}
}

func TestCLIFilenameReadsFile(t *testing.T) {
	fakeReader := strings.NewReader("")
	fakeOut := new(bytes.Buffer)
	fakeErr := new(bytes.Buffer)

	fakeFile, _ := ioutil.TempFile("", "hcl2json")

	cli := &CLI{
		Args:   []string{"hcl2json", fakeFile.Name()},
		Stdin:  fakeReader,
		Stdout: fakeOut,
		Stderr: fakeErr,
	}

	cli.Run()

	errStr := fakeErr.String()
	if errStr != "" {
		t.Fatalf("err:\n%s", errStr)
	}

	actual := fakeOut.String()
	expected := "{}\n"
	if actual != expected {
		t.Fatalf("out:\n%s", actual)
	}
}

func TestCLIDashReadsFile(t *testing.T) {
	fakeReader := strings.NewReader("\"key\"=\"val\"\n")
	fakeOut := new(bytes.Buffer)
	fakeErr := new(bytes.Buffer)

	cli := &CLI{
		Args:   []string{"hcl2json", "-"},
		Stdin:  fakeReader,
		Stdout: fakeOut,
		Stderr: fakeErr,
	}

	cli.Run()

	errStr := fakeErr.String()
	if errStr != "" {
		t.Fatalf("err:\n%s", errStr)
	}

	actual := fakeOut.String()
	expected := "{\n  \"key\": \"val\"\n}\n"
	if actual != expected {
		t.Fatalf("bad:\n%s", actual)
	}
}
