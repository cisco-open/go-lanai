package generator

import (
    "context"
    "path/filepath"
    "regexp"
)

/*****************************
	Common Output Resolver
 *****************************/

const outputRegexWithFilePrefix = `^(?:%s\.)(?P<filename>.+)(?:\.tmpl)`

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
        path, e := ConvertSrcRootToTargetDir(tmplDesc.Path, data)
        if e != nil {
            return TemplateOutputDescriptor{}, e
        }

        dir := filepath.Dir(path)
        filename := filepath.Base(path)
        matches := compiled.FindStringSubmatch(filename)
        if filenameIdx != 0 && len(matches) > filenameIdx {
            filename = matches[filenameIdx]
        }

        return TemplateOutputDescriptor{
            Path: filepath.Join(dir, filename),
        }, nil
    })
}
