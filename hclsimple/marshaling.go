package hclsimple

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Marshal is intended to provide an interface compatible with other serialization
// library such as json and yaml.
//
// It tries to render the provided data as hcl representation.
//
// It only works for struct source with proper hcl tags.
func Marshal(data interface{}) (result []byte, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%v", rec)
		}
	}()
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(data, f.Body())
	return f.Bytes(), nil
}

// Unmarshal is intended to provide an interface compatible with other serialization
// library such as json and yaml.
//
// It tries to read the provided data and fill the supplied target.
//
// It doesn't support evaluation context as Decode and DecodeFile do.
// It only works for target pointing to struct with proper hcl tags.
func Unmarshal(src []byte, target interface{}) error {
	return Decode(".hcl", src, nil, target)
}
