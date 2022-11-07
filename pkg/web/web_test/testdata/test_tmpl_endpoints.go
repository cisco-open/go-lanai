package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"fmt"
	"net/http"
	"strings"
)

func IndexPage(_ context.Context, _ *http.Request) (template.ModelView, error) {
	return template.ModelView{
		View:  "index.html.tmpl",
		Model: template.Model{
			"Title": "TemplateMVCTest",
		},
	}, nil
}

func RedirectPage(_ context.Context, _ *http.Request) (*template.ModelView, error) {
	return template.RedirectView("/index", http.StatusFound, false), nil
}


const ModelPrintTmpl = `%s=%v`

// PrintKV is a template function
func PrintKV(model map[string]any) string {
	lines := flattenMap(model, "")
	return strings.Join(lines, "\n")
}

func flattenMap[T any](m map[string]T, prefix string) []string {
	lines := make([]string, 0, len(m))
	for k, val := range m {
		var unknown interface{} = val
		switch v := unknown.(type) {
		case template.RequestContext:
			lines = append(lines, flattenMap(v, prefix + "." + k)...)
		case map[string]any:
			lines = append(lines, flattenMap(v, prefix + "." + k)...)
		case map[string]string:
			lines = append(lines, flattenMap(v, prefix + "." + k)...)
		case []any:
			for i := range v {
				k = fmt.Sprintf(`%s.%d`, prefix, i)
				lines = append(lines, fmt.Sprintf(ModelPrintTmpl, k, v[i]))
			}
		case []string:
			for i := range v {
				k = fmt.Sprintf(`%s.%d`, prefix, i)
				lines = append(lines, fmt.Sprintf(ModelPrintTmpl, k, v[i]))
			}
		default:
			k = fmt.Sprintf(`%s.%s`, prefix, k)
			lines = append(lines, fmt.Sprintf(ModelPrintTmpl, k, v))
		}
	}
	return lines
}