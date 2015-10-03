package hclprinter

import (
	"os"
	"testing"

	"github.com/hashicorp/hcl/hcl"
)

var complexHcl = `// This comes from Terraform, as a test
variable "foo" {
    default = "bar"
    description = "bar"
}

provider "aws" {
  access_key = "foo"
  secret_key = "bar"
}

provider "do" {
  api_key = "${var.foo}"
}

resource "aws_security_group" "firewall" {
    count = 5
}

resource aws_instance "web" {
    ami = "${var.foo}"
    security_groups = [
        "foo",
        "${aws_security_group.firewall.foo}"
    ]

    network_interface {
        device_index = 0
        description = "Main network interface"
    }
}

resource "aws_instance" "db" {
    security_groups = "${aws_security_group.firewall.*.id}"
    VPC = "foo"

    depends_on = ["aws_instance.web"]
}

output "web_ip" {
    value = "${aws_instance.web.private_ip}"
}
`

func TestPrint(t *testing.T) {
	obj, err := hcl.Parse(complexHcl)
	if err != nil {
		t.Fatal(err)
	}

	if err := Fprint(os.Stdout, obj); err != nil {
		t.Error(err)
	}
}
