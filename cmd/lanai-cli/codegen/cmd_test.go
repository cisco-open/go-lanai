package codegen

import (
	"embed"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

const testDir = "testdata/"

//go:embed testdata/validFileSystem
var ValidFilesystem embed.FS

type ExpectedFiles struct {
	Files []string `yaml:"files"`
}

func TestGenerateTemplates(t *testing.T) {
	tests := []struct {
		name                 string
		wantErr              bool
		filesystem           embed.FS
		outputDir            string
		expectedFileListPath string
	}{
		{
			name:                 "Should create files in destination directory in same format as input dir",
			wantErr:              false,
			filesystem:           ValidFilesystem,
			outputDir:            path.Join(testDir, "output"),
			expectedFileListPath: "testdata/validFileSystem/expectedFiles.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			outputDirAbsolute, err := filepath.Abs(tt.outputDir)
			if err != nil {
				t.Fatalf("Absolute Path not available")
			}
			fsm, err := NewFileSystemMapper(tt.filesystem, outputDirAbsolute)
			if tt.wantErr {
				g.Expect(err).NotTo(gomega.Succeed())
			} else {
				g.Expect(err).To(gomega.Succeed())

				err = GenerateTemplates(fsm, tt.filesystem)
				defer os.RemoveAll(tt.outputDir)
				g.Expect(err).To(gomega.Succeed(), "error = %v, expected nil", err)

				yamlFile, _ := ioutil.ReadFile(tt.expectedFileListPath)
				if err != nil {
					t.Fatalf("Unable to read the yaml")
				}

				f := ExpectedFiles{}
				err = yaml.Unmarshal(yamlFile, &f)
				if err != nil {
					t.Fatalf("Could not unmarshal yaml")
				}

				for _, y := range f.Files {
					_, err = os.Stat(y)
					g.Expect(err).To(gomega.Succeed(), "file %v does not exist", y)
				}
			}

		})
	}
}
