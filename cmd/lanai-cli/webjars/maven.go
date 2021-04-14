package webjars

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"path/filepath"
)

type resource struct {
	Directory string
	Includes []string
	Excludes []string
}
type pomTmplData struct {
	*cmdutils.Global
	*Arguments
	Resources []resource
}


func generatePom(ctx context.Context) error {
	resources := []resource{
		{
			Directory: defaultWebjarContentPath,
			Excludes: []string{"vendor/**"},
		},
	}
	for _, res := range Args.Resources {
		resources = append(resources, resource{
			Directory: res,
		})
	}
	data := &pomTmplData{
		Global: &cmdutils.GlobalArgs,
		Arguments: &Args,
		Resources: resources,
	}

	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "pom.xml.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.TmpDir, "pom.xml"),
		OutputPerm: 0644,
		Overwrite:  true,
		Model:      data,
	})
}

func executeMaven(ctx context.Context) error {
	_, e := cmdutils.RunShellCommands(ctx,
		cmdutils.ShellShowCmd(true),
		cmdutils.ShellUseTmpDir(),
		cmdutils.ShellCmd("mvn versions:resolve-ranges -q"),
		cmdutils.ShellCmd("mvn dependency:unpack-dependencies -q"),
		cmdutils.ShellCmd("mvn resources:copy-resources -q"),
	)
	return e
}

