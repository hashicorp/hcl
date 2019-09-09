package hclpack_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclpack"
)

func Example_marshalJSON() {
	src := `
	service "example" {
	  priority = 2
	  platform {
		os   = "linux"
		arch = "amd64"
	  }
	  process "web" {
	    exec = ["./webapp"]
	  }
	  process "worker" {
	    exec = ["./worker"]
	  }
	}
	`

	body, diags := hclpack.PackNativeFile([]byte(src), "example.svc", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		fmt.Fprintf(os.Stderr, "Failed to parse: %s", diags.Error())
		return
	}

	jb, err := body.MarshalJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal: %s", err)
		return
	}

	// Normally the compact form is best, but we'll indent just for the sake
	// of this example so the result is readable.
	var buf bytes.Buffer
	json.Indent(&buf, jb, "", " ")
	os.Stdout.Write(buf.Bytes())

	// Output:
	// {
	//  "r": {
	//   "b": [
	//    {
	//     "h": [
	//      "service",
	//      "example"
	//     ],
	//     "b": {
	//      "a": {
	//       "priority": {
	//        "s": "2",
	//        "r": "ChAKDA4QDhA"
	//       }
	//      },
	//      "b": [
	//       {
	//        "h": [
	//         "platform"
	//        ],
	//        "b": {
	//         "a": {
	//          "arch": {
	//           "s": "\"amd64\"",
	//           "r": "IiwiJCYsKCo"
	//          },
	//          "os": {
	//           "s": "\"linux\"",
	//           "r": "FiAWGBogHB4"
	//          }
	//         },
	//         "r": "Li4"
	//        },
	//        "r": "EhQSFA"
	//       },
	//       {
	//        "h": [
	//         "process",
	//         "web"
	//        ],
	//        "b": {
	//         "a": {
	//          "exec": {
	//           "s": "[\"./webapp\"]",
	//           "r": "OEA4OjxAPD4"
	//          }
	//         },
	//         "r": "QkI"
	//        },
	//        "r": "MDYwMjQ2"
	//       },
	//       {
	//        "h": [
	//         "process",
	//         "worker"
	//        ],
	//        "b": {
	//         "a": {
	//          "exec": {
	//           "s": "[\"./worker\"]",
	//           "r": "TFRMTlBUUFI"
	//          }
	//         },
	//         "r": "VlY"
	//        },
	//        "r": "REpERkhK"
	//       }
	//      ],
	//      "r": "WFg"
	//     },
	//     "r": "AggCBAYI"
	//    }
	//   ],
	//   "r": "Wlo"
	//  },
	//  "s": [
	//   "example.svc"
	//  ],
	//  "p": "BAQEAA4OAAICABISAggMABAQAAYGAAICAggIABAQAgYKAAQEAAoKAAICAAoKAAICAgYGAAgIAAYGAAICAAoKAAICAgoKAggIAA4OAAICAAoKAgwQAAgIAAYGAAICABYWAgoKAggIAA4OAAICABAQAgwQAAgIAAYGAAICABYWAgoKAgYGAgQE"
	// }
}
