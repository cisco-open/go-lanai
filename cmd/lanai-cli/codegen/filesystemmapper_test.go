package codegen

import (
	"embed"
	"github.com/onsi/gomega"
	"path"
	"testing"
)

//go:embed testdata/invalidFileSystem/MissingFileSystem
var InvalidMissingFileSystem embed.FS

//go:embed testdata/invalidFileSystem/MissingCommon
var InvalidMissingCommon embed.FS

func TestFileSystemMapper_GetOutputFileDir(t *testing.T) {
	type args struct {
		tmpl      string
		filename  string
		modifiers map[string]string
	}
	tests := []struct {
		name       string
		filesystem embed.FS
		args       args
		want       string
	}{
		{
			name:       "GetOutputFilePath should return the correct output path",
			filesystem: ValidFilesystem,
			args: args{
				tmpl:     "inner.tmpl",
				filename: "in.go",
			},
			want: path.Join("testdata", "output", "inner", "in.go"),
		},
		{
			name:       "GetOutputFilePath should return the path to the template with special names resolved",
			filesystem: ValidFilesystem,
			args: args{
				tmpl: "version.tmpl",
				modifiers: map[string]string{
					"VERSION": "v2",
				},
			},
			want: path.Join("testdata", "output", "v2"),
		},
		{
			name:       "GetOutputFilePath should return the path to the template with modifiers removed if not available",
			filesystem: ValidFilesystem,
			args: args{
				tmpl:      "version.tmpl",
				modifiers: map[string]string{},
			},
			want: path.Join("testdata", "output"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			f, err := NewFileSystemMapper(tt.filesystem, path.Join(testDir, "output"))
			g.Expect(err).To(gomega.Succeed())
			got := f.GetOutputFilePath(tt.args.tmpl, tt.args.filename, tt.args.modifiers)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFileSystemMapper_GetPathToTemplate(t *testing.T) {
	type args struct {
		tmpl string
	}
	tests := []struct {
		name       string
		filesystem embed.FS
		args       args
		want       string
	}{
		{
			name:       "GetPathToTemplate should return the correct path to the template",
			filesystem: ValidFilesystem,
			args: args{
				tmpl: "inner.tmpl",
			},
			want: path.Join("testdata", "validFileSystem", "filesystem", "inner", "inner.tmpl"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			f, err := NewFileSystemMapper(tt.filesystem, path.Join(testDir, "output"))
			g.Expect(err).To(gomega.Succeed())
			got := f.GetPathToTemplate(tt.args.tmpl)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestNewFileSystemMapper_Errors(t *testing.T) {
	type args struct {
		templates embed.FS
		outputDir string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Should fail if filesystem directory is not available",
			args: args{
				templates: InvalidMissingFileSystem,
				outputDir: "",
			},
		},
		{
			name: "Should fail if common directory is not available",
			args: args{
				templates: InvalidMissingCommon,
				outputDir: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			_, err := NewFileSystemMapper(tt.args.templates, tt.args.outputDir)
			g.Expect(err).NotTo(gomega.Succeed())
		})
	}
}
