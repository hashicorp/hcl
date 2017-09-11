package include

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl2/zcl"
	"github.com/hashicorp/hcl2/zclparse"
)

// FileResolver creates and returns a Resolver that interprets include paths
// as filesystem paths relative to the calling configuration file.
//
// When an include is requested, the source filename of the calling config
// file is first interpreted relative to the given basePath, and then the
// path given in configuration is interpreted relative to the resulting
// absolute caller configuration directory.
//
// This resolver assumes that all calling bodies are loaded from local files
// and that the paths to these files were correctly provided to the parser,
// either absolute or relative to the given basePath.
//
// If the path given in configuration ends with ".json" then the referenced
// file is interpreted as JSON. Otherwise, it is interpreted as zcl native
// syntax.
func FileResolver(baseDir string, parser *zclparse.Parser) Resolver {
	return &fileResolver{
		BaseDir: baseDir,
		Parser:  parser,
	}
}

type fileResolver struct {
	BaseDir string
	Parser  *zclparse.Parser
}

func (r fileResolver) ResolveBodyPath(path string, refRange zcl.Range) (zcl.Body, zcl.Diagnostics) {
	callerFile := filepath.Join(r.BaseDir, refRange.Filename)
	callerDir := filepath.Dir(callerFile)
	targetFile := filepath.Join(callerDir, path)

	var f *zcl.File
	var diags zcl.Diagnostics
	if strings.HasSuffix(targetFile, ".json") {
		f, diags = r.Parser.ParseJSONFile(targetFile)
	} else {
		f, diags = r.Parser.ParseZCLFile(targetFile)
	}

	return f.Body, diags
}
