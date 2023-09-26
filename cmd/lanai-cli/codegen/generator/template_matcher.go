package generator

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
    "strings"
)

/*****************************
	Common Template Matchers
 *****************************/

const patternWithFilePrefix = `**/%s.*.tmpl`

// matchPatterns match template path with wildcard patterns. e.g. **/project.*.tmpl
func matchPatterns(patterns ...string) TemplateMatcher {
    return &cmdutils.GenericMatcher[TemplateDescriptor]{
        Description: strings.Join(patterns, ", "),
        MatchFunc: func(ctx context.Context, tmplDesc TemplateDescriptor) (bool, error) {
            for _, pattern := range patterns {
                if match, e := cmdutils.MatchPathPattern(pattern, tmplDesc.Path); e == nil && match {
                    return true, nil
                }
            }
            return false, nil
        },
    }
}

func isDir() TemplateMatcher {
    return &cmdutils.GenericMatcher[TemplateDescriptor]{
        Description: "directory",
        MatchFunc: func(ctx context.Context, tmplDesc TemplateDescriptor) (bool, error) {
            return tmplDesc.FileInfo.IsDir(), nil
        },
    }
}

func isTmplFile() TemplateMatcher {
    return &cmdutils.GenericMatcher[TemplateDescriptor]{
        Description: "file",
        MatchFunc: func(ctx context.Context, tmplDesc TemplateDescriptor) (bool, error) {
            return !tmplDesc.FileInfo.IsDir(), nil
        },
    }
}
