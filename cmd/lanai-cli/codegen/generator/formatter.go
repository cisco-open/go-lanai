package generator

import (
	"bytes"
	"fmt"
	regoformat "github.com/open-policy-agent/opa/format"
	"go/format"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	FileTypeUnknown FileType = ""
	FileTypeGo      FileType = "go"
	FileTypeYaml    FileType = "yaml"
	FileTypeRego    FileType = "rego"
)

type FileType string

type TextFormatter interface {
	Format(in []byte) ([]byte, error)
}

var errFormatterUnsupportedFileType = fmt.Errorf("unsupported file type to format")

var formatters = map[FileType]TextFormatter{
	FileTypeGo:   GoSourceFormatter{},
	FileTypeYaml: YamlFormatter{},
	FileTypeRego: RegoFormatter{},
}

// FormatFile clean up given file with file path, replace the content with prettier format.
// When type hint is unknown, this function uses file's extension to deduct file type
func FormatFile(path string, typeHint FileType) error {
	if typeHint == FileTypeUnknown {
		switch filepath.Ext(path) {
		case ".go", ".goref":
			typeHint = FileTypeGo
		case ".yml", ".yaml":
			typeHint = FileTypeYaml
		case ".rego":
			typeHint = FileTypeRego
		default:
			return errFormatterUnsupportedFileType
		}
	}
	formatter, ok := formatters[typeHint]
	if !ok {
		return errFormatterUnsupportedFileType
	}
	return formatFile(path, formatter)
}

func formatFile(path string, formatter TextFormatter) error {
	content, fi, e := readFile(path)
	formatted, e := formatter.Format(content)
	if e != nil {
		return e
	}
	return writeFile(path, formatted, fi)
}

func readFile(path string) ([]byte, fs.FileInfo, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, nil, e
	}
	defer func() { _ = f.Close() }()

	fi, e := f.Stat()
	if e != nil {
		return nil, nil, e
	}
	content, e := io.ReadAll(f)
	if e != nil {
		return nil, nil, e
	}
	return content, fi, nil
}

func writeFile(path string, data []byte, fi fs.FileInfo) error {
	perm := fi.Mode().Perm()
	if perm == 0000 {
		perm = 0644
	}
	return os.WriteFile(path, data, perm)
}

/*********************
   Text Formatters
*********************/

type GoSourceFormatter struct{}

func (f GoSourceFormatter) Format(in []byte) ([]byte, error) {
	return format.Source(in)
}

// YamlFormatter format single yaml document text (without "---").
// Add empty line between each top level component if root document is an object or list
type YamlFormatter struct{}

func (f YamlFormatter) Format(in []byte) ([]byte, error) {
	// decode
	decoder := yaml.NewDecoder(bytes.NewReader(in))
	root := yaml.Node{}
	if e := decoder.Decode(&root); e != nil {
		return nil, e
	}

	// encode
	var buf bytes.Buffer
	for _, node := range f.topLevelNodes(&root) {
		// for each top level component, add a empty line after it
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if e := encoder.Encode(node); e != nil {
			return nil, e
		}
		_, _ = fmt.Fprintln(&buf, "")
		_ = encoder.Close()
	}
	return buf.Bytes(), nil
}

func (f YamlFormatter) topLevelNodes(root *yaml.Node) (nodes []*yaml.Node) {
	if root.Kind != yaml.DocumentNode {
		return
	}
	nodes = make([]*yaml.Node, 0, len(root.Content)*5)
	for _, node := range root.Content {
		switch node.Kind {
		case yaml.MappingNode:
			for i := 1; i < len(node.Content); i += 2 {
				nodes = append(nodes, f.copyNode(node, node.Content[i-1], node.Content[i]))
			}
		case yaml.SequenceNode:
			for i := range node.Content {
				nodes = append(nodes, f.copyNode(node, node.Content[i]))
			}
		default:
			nodes = append(nodes, node)
		}
	}
	return
}

func (f YamlFormatter) copyNode(src *yaml.Node, contents ...*yaml.Node) *yaml.Node {
	node := *src
	node.Content = contents
	return &node
}

type RegoFormatter struct{}

func (f RegoFormatter) Format(in []byte) ([]byte, error) {
	return regoformat.Source("generated.rego", in)
}
