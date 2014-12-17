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
	d, err := ioutil.ReadAll(os.Stdin)
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
