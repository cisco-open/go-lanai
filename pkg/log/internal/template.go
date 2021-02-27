package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"io"
	"strings"
	"text/template"
)

const (
	logTemplate = "lanai-log-template"
)

var (

)

type TextFormatter interface {
	Format(kvs Fields, w io.Writer) error
}

type TemplatedFormatter struct {
	text string
	tmpl *template.Template
	fixedFields utils.StringSet
}

func NewTemplatedFormatter(tmpl string, fixedFields utils.StringSet) *TemplatedFormatter {
	formatter := &TemplatedFormatter{
		text: tmpl,
		fixedFields: fixedFields,
	}
	formatter.init()
	return formatter
}

func (f *TemplatedFormatter) init() {
	if !strings.HasSuffix(f.text, "\n") {
		f.text = f.text + "\n"
	}

	t, e := template.New(logTemplate).
		Option("missingkey=zero").
		Funcs(TmplFuncMap).
		Funcs(template.FuncMap{
			"kv": MakeKVFunc(f.fixedFields),
		}).
		Parse(f.text)
	if e != nil {
		panic(e)
	}
	f.tmpl = t
}

func (f TemplatedFormatter) Format(kvs Fields, w io.Writer) error {
	return f.tmpl.Execute(w, kvs)
}


