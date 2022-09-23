package codegen

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"embed"
	"github.com/spf13/cobra"
	"path"
)

const (
	CommandName = "codegen"
)

var (
	Cmd = &cobra.Command{
		Use:                CommandName,
		Short:              "Given openapi contract, generate controllers/structs",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}
	Args = Arguments{}
)

type Arguments struct {
}

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

//go:embed templates
var TmplFS embed.FS

func generateTestFile(ctx context.Context, fs embed.FS, file File, fsm FileSystemMapper) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         fs,
		TmplName:   fsm.GetPathToTemplate(file.template),
		Output:     fsm.GetOutputFilePath(file.template, file.filename, nil),
		OutputPerm: 0644,
		Overwrite:  true,
		Model:      file.model,
		CommonTmpl: path.Join(fsm.commonDir, "*"),
	})
}

type File struct {
	template string
	filename string
	model    interface{}
}

// GenerateTemplates will generate the files
// fs needs to have a "filesystem" folder to work
func GenerateTemplates(fsm FileSystemMapper, templates embed.FS) error {
	// Given a filename & desired template to use, output stuff!
	// TODO: Actual logic for files, this is just an example
	file := File{template: "test.tmpl", filename: "test.go", model: ""}
	err := generateTestFile(context.Background(), templates, file, fsm)
	if err != nil {
		return err
	}
	file2 := File{template: "inner.tmpl", filename: "inner.go", model: ""}
	err = generateTestFile(context.Background(), templates, file2, fsm)
	if err != nil {
		return err
	}
	return nil
}

func Run(cmd *cobra.Command, _ []string) error {
	fsm, err := NewFileSystemMapper(TmplFS, cmdutils.GlobalArgs.OutputDir)
	if err != nil {
		return err
	}
	if err := GenerateTemplates(fsm, TmplFS); err != nil {
		return err
	}
	return nil
}
