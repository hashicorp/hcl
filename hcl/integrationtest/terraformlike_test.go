package integrationtest

import (
	"reflect"
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/ext/dynblock"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hcl/json"
	"github.com/hashicorp/hcl2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// TestTerraformLike parses both a native syntax and a JSON representation
// of the same HashiCorp Terraform-like configuration structure and then makes
// assertions against the result of each.
//
// Terraform exercises a lot of different HCL codepaths, so this is not
// exhaustive but tries to cover a variety of different relevant scenarios.
func TestTerraformLike(t *testing.T) {
	tests := map[string]func() (*hcl.File, hcl.Diagnostics){
		"native syntax": func() (*hcl.File, hcl.Diagnostics) {
			return hclsyntax.ParseConfig(
				[]byte(terraformLikeNativeSyntax),
				"config.tf", hcl.Pos{Line: 1, Column: 1},
			)
		},
		"JSON": func() (*hcl.File, hcl.Diagnostics) {
			return json.Parse(
				[]byte(terraformLikeJSON),
				"config.tf.json",
			)
		},
	}

	type Variable struct {
		Name string `hcl:"name,label"`
	}
	type Resource struct {
		Type      string         `hcl:"type,label"`
		Name      string         `hcl:"name,label"`
		Config    hcl.Body       `hcl:",remain"`
		DependsOn hcl.Expression `hcl:"depends_on,attr"`
	}
	type Root struct {
		Variables []*Variable `hcl:"variable,block"`
		Resources []*Resource `hcl:"resource,block"`
	}
	instanceDecode := &hcldec.ObjectSpec{
		"image_id": &hcldec.AttrSpec{
			Name:     "image_id",
			Required: true,
			Type:     cty.String,
		},
		"instance_type": &hcldec.AttrSpec{
			Name:     "instance_type",
			Required: true,
			Type:     cty.String,
		},
	}
	securityGroupDecode := &hcldec.ObjectSpec{
		"ingress": &hcldec.BlockListSpec{
			TypeName: "ingress",
			Nested: &hcldec.ObjectSpec{
				"cidr_block": &hcldec.AttrSpec{
					Name:     "cidr_block",
					Required: true,
					Type:     cty.String,
				},
			},
		},
	}

	for name, loadFunc := range tests {
		t.Run(name, func(t *testing.T) {
			file, diags := loadFunc()
			if len(diags) != 0 {
				t.Errorf("unexpected diagnostics during parse")
				for _, diag := range diags {
					t.Logf("- %s", diag)
				}
				return
			}

			body := file.Body

			var root Root
			diags = gohcl.DecodeBody(body, nil, &root)
			if len(diags) != 0 {
				t.Errorf("unexpected diagnostics during root eval")
				for _, diag := range diags {
					t.Logf("- %s", diag)
				}
				return
			}

			wantVars := []*Variable{
				{
					Name: "image_id",
				},
			}
			if gotVars := root.Variables; !reflect.DeepEqual(gotVars, wantVars) {
				t.Errorf("wrong Variables\ngot:  %swant: %s", spew.Sdump(gotVars), spew.Sdump(wantVars))
			}

			if got, want := len(root.Resources), 3; got != want {
				t.Fatalf("wrong number of Resources %d; want %d", got, want)
			}

			sort.Slice(root.Resources, func(i, j int) bool {
				return root.Resources[i].Name < root.Resources[j].Name
			})

			t.Run("resource 0", func(t *testing.T) {
				r := root.Resources[0]
				if got, want := r.Type, "happycloud_security_group"; got != want {
					t.Errorf("wrong type %q; want %q", got, want)
				}
				if got, want := r.Name, "private"; got != want {
					t.Errorf("wrong type %q; want %q", got, want)
				}

				// For this one we're including support for the dynamic block
				// extension, since Terraform uses this to allow dynamic
				// generation of blocks within resource configuration.
				forEachCtx := &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.ObjectVal(map[string]cty.Value{
							"extra_private_cidr_blocks": cty.ListVal([]cty.Value{
								cty.StringVal("172.16.0.0/12"),
								cty.StringVal("169.254.0.0/16"),
							}),
						}),
					},
				}
				dynBody := dynblock.Expand(r.Config, forEachCtx)

				cfg, diags := hcldec.Decode(dynBody, securityGroupDecode, nil)
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics decoding Config")
					for _, diag := range diags {
						t.Logf("- %s", diag)
					}
					return
				}
				wantCfg := cty.ObjectVal(map[string]cty.Value{
					"ingress": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"cidr_block": cty.StringVal("10.0.0.0/8"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"cidr_block": cty.StringVal("192.168.0.0/16"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"cidr_block": cty.StringVal("172.16.0.0/12"),
						}),
						cty.ObjectVal(map[string]cty.Value{
							"cidr_block": cty.StringVal("169.254.0.0/16"),
						}),
					}),
				})
				if !cfg.RawEquals(wantCfg) {
					t.Errorf("wrong config\ngot:  %#v\nwant: %#v", cfg, wantCfg)
				}
			})

			t.Run("resource 1", func(t *testing.T) {
				r := root.Resources[1]
				if got, want := r.Type, "happycloud_security_group"; got != want {
					t.Errorf("wrong type %q; want %q", got, want)
				}
				if got, want := r.Name, "public"; got != want {
					t.Errorf("wrong type %q; want %q", got, want)
				}

				cfg, diags := hcldec.Decode(r.Config, securityGroupDecode, nil)
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics decoding Config")
					for _, diag := range diags {
						t.Logf("- %s", diag)
					}
					return
				}
				wantCfg := cty.ObjectVal(map[string]cty.Value{
					"ingress": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"cidr_block": cty.StringVal("0.0.0.0/0"),
						}),
					}),
				})
				if !cfg.RawEquals(wantCfg) {
					t.Errorf("wrong config\ngot:  %#v\nwant: %#v", cfg, wantCfg)
				}
			})

			t.Run("resource 2", func(t *testing.T) {
				r := root.Resources[2]
				if got, want := r.Type, "happycloud_instance"; got != want {
					t.Errorf("wrong type %q; want %q", got, want)
				}
				if got, want := r.Name, "test"; got != want {
					t.Errorf("wrong type %q; want %q", got, want)
				}

				vars := hcldec.Variables(r.Config, &hcldec.AttrSpec{
					Name: "image_id",
					Type: cty.String,
				})
				if got, want := len(vars), 1; got != want {
					t.Errorf("wrong number of variables in image_id %#v; want %#v", got, want)
				}
				if got, want := vars[0].RootName(), "var"; got != want {
					t.Errorf("wrong image_id variable RootName %#v; want %#v", got, want)
				}

				ctx := &hcl.EvalContext{
					Variables: map[string]cty.Value{
						"var": cty.ObjectVal(map[string]cty.Value{
							"image_id": cty.StringVal("image-1234"),
						}),
					},
				}
				cfg, diags := hcldec.Decode(r.Config, instanceDecode, ctx)
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics decoding Config")
					for _, diag := range diags {
						t.Logf("- %s", diag)
					}
					return
				}
				wantCfg := cty.ObjectVal(map[string]cty.Value{
					"instance_type": cty.StringVal("z3.weedy"),
					"image_id":      cty.StringVal("image-1234"),
				})
				if !cfg.RawEquals(wantCfg) {
					t.Errorf("wrong config\ngot:  %#v\nwant: %#v", cfg, wantCfg)
				}

				exprs, diags := hcl.ExprList(r.DependsOn)
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics extracting depends_on")
					for _, diag := range diags {
						t.Logf("- %s", diag)
					}
					return
				}
				if got, want := len(exprs), 1; got != want {
					t.Errorf("wrong number of depends_on exprs %#v; want %#v", got, want)
				}

				traversal, diags := hcl.AbsTraversalForExpr(exprs[0])
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics decoding depends_on[0]")
					for _, diag := range diags {
						t.Logf("- %s", diag)
					}
					return
				}
				if got, want := len(traversal), 2; got != want {
					t.Errorf("wrong number of depends_on traversal steps %#v; want %#v", got, want)
				}
				if got, want := traversal.RootName(), "happycloud_security_group"; got != want {
					t.Errorf("wrong depends_on traversal RootName %#v; want %#v", got, want)
				}
			})
		})
	}
}

const terraformLikeNativeSyntax = `

variable "image_id" {
}

resource "happycloud_instance" "test" {
  instance_type = "z3.weedy"
  image_id      = var.image_id

  depends_on = [
    happycloud_security_group.public,
  ]
}

resource "happycloud_security_group" "public" {
  ingress {
    cidr_block = "0.0.0.0/0"
  }
}

resource "happycloud_security_group" "private" {
  ingress {
    cidr_block = "10.0.0.0/8"
  }
  ingress {
    cidr_block = "192.168.0.0/16"
  }
  dynamic "ingress" {
    for_each = var.extra_private_cidr_blocks
    content {
      cidr_block = ingress.value
    }
  }
}

`

const terraformLikeJSON = `
{
  "variable": {
    "image_id": {}
  },
  "resource": {
    "happycloud_instance": {
      "test": {
        "instance_type": "z3.weedy",
        "image_id": "${var.image_id}",
        "depends_on": [
          "happycloud_security_group.public"
        ]
      }
    },
    "happycloud_security_group": {
      "public": {
        "ingress": {
          "cidr_block": "0.0.0.0/0"
        }
      },
      "private": {
        "ingress": [
          {
            "cidr_block": "10.0.0.0/8"
          },
          {
            "cidr_block": "192.168.0.0/16"
          }
        ],
        "dynamic": {
          "ingress": {
            "for_each": "${var.extra_private_cidr_blocks}",
            "iterator": "block",
            "content": {
              "cidr_block": "${block.value}"
            }
          }
        }
      }
    }
  }
}
`
