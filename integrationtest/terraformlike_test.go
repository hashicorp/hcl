package integrationtest

import (
	"reflect"
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/dynblock"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
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
	type Module struct {
		Name      string         `hcl:"name,label"`
		Providers hcl.Expression `hcl:"providers"`
	}
	type Root struct {
		Variables []*Variable `hcl:"variable,block"`
		Resources []*Resource `hcl:"resource,block"`
		Modules   []*Module   `hcl:"module,block"`
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
		"tags": &hcldec.AttrSpec{
			Name:     "tags",
			Required: false,
			Type:     cty.Map(cty.String),
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
					"tags": cty.MapVal(map[string]cty.Value{
						"Name":        cty.StringVal("foo"),
						"Environment": cty.StringVal("prod"),
					}),
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

			t.Run("module", func(t *testing.T) {
				if got, want := len(root.Modules), 1; got != want {
					t.Fatalf("wrong number of Modules %d; want %d", got, want)
				}
				mod := root.Modules[0]
				if got, want := mod.Name, "foo"; got != want {
					t.Errorf("wrong module name %q; want %q", got, want)
				}

				pExpr := mod.Providers
				pairs, diags := hcl.ExprMap(pExpr)
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics extracting providers")
					for _, diag := range diags {
						t.Logf("- %s", diag)
					}
				}
				if got, want := len(pairs), 1; got != want {
					t.Fatalf("wrong number of key/value pairs in providers %d; want %d", got, want)
				}

				pair := pairs[0]
				kt, diags := hcl.AbsTraversalForExpr(pair.Key)
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics extracting providers key %#v", pair.Key)
					for _, diag := range diags {
						t.Logf("- %s", diag)
					}
				}
				vt, diags := hcl.AbsTraversalForExpr(pair.Value)
				if len(diags) != 0 {
					t.Errorf("unexpected diagnostics extracting providers value  %#v", pair.Value)
					for _, diag := range diags {
						t.Logf("- %s", diag)
					}
				}

				if got, want := len(kt), 1; got != want {
					t.Fatalf("wrong number of key traversal steps %d; want %d", got, want)
				}
				if got, want := len(vt), 2; got != want {
					t.Fatalf("wrong number of value traversal steps %d; want %d", got, want)
				}

				if got, want := kt.RootName(), "null"; got != want {
					t.Errorf("wrong number key traversal root %s; want %s", got, want)
				}
				if got, want := vt.RootName(), "null"; got != want {
					t.Errorf("wrong number value traversal root %s; want %s", got, want)
				}
				if at, ok := vt[1].(hcl.TraverseAttr); ok {
					if got, want := at.Name, "foo"; got != want {
						t.Errorf("wrong number value traversal attribute name %s; want %s", got, want)
					}
				} else {
					t.Errorf("wrong value traversal [1] type %T; want hcl.TraverseAttr", vt[1])
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

  tags = {
  "Name" = "foo"
  "${"Environment"}" = "prod"
  }

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

module "foo" {
  providers = {
    null = null.foo
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
        "tags": {
            "Name": "foo",
            "${\"Environment\"}": "prod"
        },
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
  },
  "module": {
    "foo": {
      "providers": {
        "null": "null.foo"
      }
    }
  }
}
`
