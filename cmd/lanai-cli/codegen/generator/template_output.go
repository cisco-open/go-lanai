package generator

import (
    "bytes"
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
    "fmt"
    "path/filepath"
    "regexp"
    "strings"
    "text/template"
)

/*****************************
	Common Output Resolver
 *****************************/

const (
	outputRegexWithFilePrefix = `^(?:%s\.)(?P<filename>.+)(?:\.tmpl)`
)

// regexOutputResolver resolve output descriptor with following rules:
// 1. Apply generation data substitution on template's path
// 2. Use given regular expression resolve output file name.
//    The regular expression must contain a named capturing group "filename". e.g. (?:project\.)(?P<filename>.+)(?:\.tmpl)
// Note: when regex doesn't contains "filename" group, the 2nd step takes no effect
func regexOutputResolver(regex string) TemplateOutputResolver {
    const filenameGroup = `filename`
    compiled := regexp.MustCompile(regex)
    var filenameIdx int
    for i, n := range compiled.SubexpNames() {
        if n == filenameGroup {
            filenameIdx = i
        }
    }
    return TemplateOutputResolverFunc(func(ctx context.Context, tmplDesc TemplateDescriptor, data GenerationData) (TemplateOutputDescriptor, error) {
        resolvedTmplPath, e := resolvePathWithData(tmplDesc.Path, data)
        if e != nil {
            return TemplateOutputDescriptor{}, e
        }
        dir := resolveOutputDir(resolvedTmplPath)

        filename := filepath.Base(resolvedTmplPath)
        matches := compiled.FindStringSubmatch(filename)
        if filenameIdx != 0 && len(matches) > filenameIdx {
            filename = matches[filenameIdx]
        }

        return TemplateOutputDescriptor{
            Path: filepath.Join(dir, filename),
        }, nil
    })
}

// resolveOutputDir will take a template path and return an absolute path of the output dir.
// e.g. pkg/init/project.package.go.tmpl -> /path/to/output/folder/pkg/init
// Note: srcPath should be always relative to template root
func resolveOutputDir(tmplPath string) string {
    return filepath.Join(cmdutils.GlobalArgs.OutputDir, filepath.Dir(tmplPath))
}

var pathVarRegex = regexp.MustCompile(`@([^.].*)@`)
const pathVarReplacement = `@.${1}@`

// resolvePathWithData take an unresolved path and  apply any @...@ with values stored in given data.
// e.g. cmd/@.Project.Name@/project.main.go.tmpl would be resolved to cmd/testservice/project.main.go.tmpl
func resolvePathWithData(unresolved string, data GenerationData) (string, error) {
    toResolve := pathVarRegex.ReplaceAllString(unresolved, pathVarReplacement)
    toResolve = strings.ReplaceAll(toResolve, `@.@`, `.`)
    pathTmpl, e := template.New("filename").Delims("@", "@").Parse(toResolve)
    if e != nil {
        return "", fmt.Errorf("cannot resolve path: [%s]: %v", unresolved, e)
    }
    var buf bytes.Buffer
    if e := pathTmpl.Execute(&buf, data); e != nil {
        return "", fmt.Errorf("cannot resolve path: [%s]: %v", unresolved, e)
    }
    return buf.String(), nil
}
