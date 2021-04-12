package webjars

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"os"
	"path/filepath"
	"text/template"
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

func generatePom(_ context.Context) error {
	t, e := template.New("templates").ParseFS(TmplFS, "*.tmpl")
	if e != nil {
		return e
	}

	// prepare temp pom.xml to write
	out := filepath.Join(cmdutils.GlobalArgs.TmpDir, "pom.xml")
	f, e := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	defer func() {_ = f.Close()}()

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
	return t.ExecuteTemplate(f, "pom.xml.tmpl", data)
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

