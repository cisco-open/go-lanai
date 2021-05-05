package initcmd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"path/filepath"
	"text/template"
)

type TemplateData struct {
	Package     string
	Executables map[string]Executable
	Generates   []Generate
	Resources   []Resource
}

func forceUpdateServiceMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile"),
		OutputPerm: 0644,
		Overwrite:  Args.Force,
		Model:      &Module,
		Customizer: func(t *template.Template) {
			// custom delimiter, we want this file's templates shows as-is
			t.Delims("{{{", "}}}")
		},
	})
}


func generateServiceCICDMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile-CICD.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile-Generated"),
		OutputPerm: 0644,
		Overwrite:  true,
		Model:      &Module,
	})
}

func generateServiceBuildMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile-Build.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile-Build"),
		OutputPerm: 0644,
		Overwrite:  Args.Force,
		Model:      &Module,
	})
}

func generateDockerfile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Dockerfile.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "build/package/Dockerfile"),
		OutputPerm: 0644,
		Overwrite:  Args.Force,
		Model:      &Module,
	})
}

func generateLibsCICDMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile-Libs.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile-Generated"),
		OutputPerm: 0644,
		Overwrite:  true,
		Model:      &Module,
	})
}