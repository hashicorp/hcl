// This is a helper to transform the HCL.yaml-tmLanguage file (the source of
// record) into both HCL.json-tmLanguage and HCL.tmLanguage (in plist XML
// format).
//
// Run this after making updates to HCL.yaml-tmLanguage to generate the other
// formats.
//
// This file is intended to be run with "go run":
//
//     go run ./build.go
//
// This file is also set up to run itself under "go generate":
//
//     go generate .

package main

//go:generate go run ./build.go

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
	plist "howett.net/plist"

	multierror "github.com/hashicorp/go-multierror"
)

func main() {
	err := realMain()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func realMain() error {
	var err error
	buildErr := build("HCL")
	if buildErr != nil {
		err = multierror.Append(err, fmt.Errorf("in HCL: %s", buildErr))
	}
	buildErr = build("HCLTemplate")
	if buildErr != nil {
		err = multierror.Append(err, fmt.Errorf("in HCLTemplate: %s", buildErr))
	}
	buildErr = build("HCLExpression")
	if buildErr != nil {
		err = multierror.Append(err, fmt.Errorf("in HCLExpression: %s", buildErr))
	}
	return err
}

func build(basename string) error {
	yamlSrc, err := ioutil.ReadFile(basename + ".yaml-tmLanguage")
	if err != nil {
		return err
	}

	var content interface{}
	err = yaml.Unmarshal(yamlSrc, &content)
	if err != nil {
		return err
	}

	// Normalize the value so it's both JSON- and plist-friendly.
	content = prepare(content)

	jsonSrc, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}

	plistSrc, err := plist.MarshalIndent(content, plist.XMLFormat, "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(basename+".json-tmLanguage", jsonSrc, os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(basename+".tmLanguage", plistSrc, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func prepare(v interface{}) interface{} {
	switch tv := v.(type) {

	case map[interface{}]interface{}:
		var ret map[string]interface{}
		if len(tv) == 0 {
			return ret
		}
		ret = make(map[string]interface{}, len(tv))
		for k, v := range tv {
			ret[k.(string)] = prepare(v)
		}
		return ret

	case []interface{}:
		for i := range tv {
			tv[i] = prepare(tv[i])
		}
		return tv

	default:
		return v
	}
}
