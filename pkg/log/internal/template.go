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
	text        string
	tmpl        *template.Template
	fixedFields utils.StringSet
	isTerm      bool
}

func NewTemplatedFormatter(tmpl string, fixedFields utils.StringSet, isTerm bool) *TemplatedFormatter {
	formatter := &TemplatedFormatter{
		text:        tmpl,
		fixedFields: fixedFields,
		isTerm:      isTerm,
	}
	formatter.init()
	return formatter
}

func (f *TemplatedFormatter) init() {
	if !strings.HasSuffix(f.text, "\n") {
		f.text = f.text + "\n"
	}

	funcMap := TmplFuncMapNonTerm
	colorFuncMap := TmplColorFuncMapNonTerm
	if f.isTerm {
		colorFuncMap = TmplFuncMap
		funcMap = TmplColorFuncMap
	}

	t, e := template.New(logTemplate).
		Option("missingkey=zero").
		Funcs(funcMap).
		Funcs(colorFuncMap).
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


