package internal

import (
	"bytes"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"go.uber.org/zap/buffer"
	"io"
	"strings"
	"text/template"
)

const (
	LogKeyMessage    = "msg"
	LogKeyName       = "logger"
	LogKeyTimestamp  = "time"
	LogKeyCaller     = "caller"
	LogKeyLevel      = "level"
	LogKeyContext    = "ctx"
	LogKeyStacktrace = "stacktrace"
)

const (
	logTemplate = "lanai-log-template"
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

func (f *TemplatedFormatter) Format(kvs Fields, w io.Writer) error {
	switch w.(type) {
	case *buffer.Buffer:
		return f.tmpl.Execute(w, kvs)
	default:
		// from documents of template.Template.Execute:
		// 		A template may be executed safely in parallel, although if parallel
		// 		executions share a Writer the output may be interleaved.
		// to prevent this from happening, we use an in-memory buffer. Hopefully this is faster than mutex locking
		var buf bytes.Buffer
		if e := f.tmpl.Execute(&buf, kvs); e != nil {
			return e
		}
		if _, e := w.Write(buf.Bytes()); e != nil {
			return e
		}
		return nil
	}
}
