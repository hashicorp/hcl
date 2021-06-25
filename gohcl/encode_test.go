package gohcl_test

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func ExampleEncodeIntoBody() {
	type DynamicExample struct {
		Foo string  `hcl:"foo"`
		Bar float64 `hcl:"bar"`
		Baz bool    `hcl:"baz"`
	}
	type HTTPOptions struct {
		Listener string `hcl:"listener,label"`
		Address  string `hcl:"address"`
		Secure   bool   `hcl:"secure"`
	}
	type GRPCOptions struct {
		Listener string `hcl:"listener,label"`
		Address  string `hcl:"address"`
	}
	type MQTTOptions struct {
		Listener string   `hcl:"listener,label"`
		Address  string   `hcl:"address"`
		Topics   []string `hcl:"topics"`
	}

	type Service struct {
		Name string   `hcl:"name,label"`
		Exe  []string `hcl:"executable"`
	}
	type Constraints struct {
		OS   string `hcl:"os"`
		Arch string `hcl:"arch"`
	}
	type App struct {
		Name        string         `hcl:"name"`
		Desc        string         `hcl:"description"`
		Constraints *Constraints   `hcl:"constraints,block"`
		Services    []Service      `hcl:"service,block"`
		Options     [3]interface{} `hcl:"option,block"`
		Dynamic     interface{}    `hcl:"whatever,block"`
	}

	app := App{
		Name: "awesome-app",
		Desc: "Such an awesome application",
		Constraints: &Constraints{
			OS:   "linux",
			Arch: "amd64",
		},
		Services: []Service{
			{
				Name: "web",
				Exe:  []string{"./web", "--listen=:8080"},
			},
			{
				Name: "worker",
				Exe:  []string{"./worker"},
			},
		},
		Options: [3]interface{}{
			HTTPOptions{
				Listener: "http",
				Address:  ":8080",
			},
			GRPCOptions{
				Listener: "grpc",
				Address:  ":5051",
			},
			MQTTOptions{
				Listener: "mqtt",
				Address:  ":1883",
				Topics:   []string{"foo", "bar"},
			},
		},
		Dynamic: DynamicExample{
			Foo: "foo",
			Bar: 42,
			Baz: true,
		},
	}

	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&app, f.Body())
	fmt.Printf("%s", f.Bytes())

	// Output:
	// name        = "awesome-app"
	// description = "Such an awesome application"
	//
	// constraints {
	//   os   = "linux"
	//   arch = "amd64"
	// }
	//
	// service "web" {
	//   executable = ["./web", "--listen=:8080"]
	// }
	// service "worker" {
	//   executable = ["./worker"]
	// }
	//
	// option "http" {
	//   address = ":8080"
	//   secure  = false
	// }
	// option "grpc" {
	//   address = ":5051"
	// }
	// option "mqtt" {
	//   address = ":1883"
	//   topics  = ["foo", "bar"]
	// }
	//
	// whatever {
	//   foo = "foo"
	//   bar = 42
	//   baz = true
	// }

}
