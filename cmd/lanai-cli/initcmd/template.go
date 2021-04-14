package initcmd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"path/filepath"
)

type TemplateData struct {
	Package     string
	Executables map[string]Executable
	Generates   []Generate
	Resources   []Resource
}

func generateMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile-Build.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile-Build"),
		OutputPerm: 0644,
		Overwrite:  Args.Force,
		Model:      &Module,
	})
}