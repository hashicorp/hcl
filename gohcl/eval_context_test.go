package gohcl

import (
	"bytes"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

var (
	valueComparer = cmp.Comparer(cty.Value.RawEquals)
)

func TestEvalContext(t *testing.T) {

	type ServiceConfig struct {
		Type       string `hcl:"type,label"`
		Name       string `hcl:"name,label"`
		ListenAddr string `hcl:"listen_addr"`
	}
	type Config struct {
		IOMode   string          `hcl:"io_mode"`
		Services []ServiceConfig `hcl:"service,block"`
	}

	type Context struct {
		Pid string
	}

	tests := []struct {
		Input  interface{}
		Output hcl.EvalContext
	}{
		{
			Input: &Context{
				Pid: "fake-pid",
			},
			Output: hcl.EvalContext{
				Variables: map[string]cty.Value{
					"pid": cty.StringVal("fake-pid"),
				},
			},
		},
		{
			Input: &Config{
				IOMode: "fake-mode",
				Services: []ServiceConfig{
					{
						Type:       "t",
						Name:       "n",
						ListenAddr: "addr",
					},
				},
			},
			Output: hcl.EvalContext{
				Variables: map[string]cty.Value{
					"i_o_mode": cty.StringVal("fake-mode"),
					"services": cty.ListVal([]cty.Value{
						cty.MapVal(map[string]cty.Value{
							"type":        cty.StringVal("t"),
							"name":        cty.StringVal("n"),
							"listen_addr": cty.StringVal("addr"),
						}),
					}),
				},
			},
		},
	}

	for index, test := range tests {
		t.Run(fmt.Sprintf("test-%d", index), func(t *testing.T) {
			realOutput := EvalContext(test.Input)

			gotVal := realOutput.Variables
			wantVal := test.Output.Variables

			if !cmp.Equal(gotVal, wantVal, valueComparer) {
				diff := cmp.Diff(gotVal, wantVal, cmp.Comparer(func(a, b []byte) bool {
					return bytes.Equal(a, b)
				}))
				t.Errorf(
					"wrong result\nvalue: %#v\ngot:   %#v\nwant:  %#v\ndiff:  %s",
					test.Input, gotVal, wantVal, diff,
				)
			}

		})
	}
}
