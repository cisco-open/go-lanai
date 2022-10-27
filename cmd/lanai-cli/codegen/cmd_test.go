package codegen

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator"
	"embed"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-cmp/cmp"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"
)

const testDir = "testdata/"

//go:embed all:template/src
var ActualFilesystem embed.FS

const serviceName = "testservice"

func TestGenerateTemplates(t *testing.T) {
	tests := []struct {
		name       string
		contract   string
		wantErr    bool
		filesystem fs.FS
		outputDir  string
		goldenDir  string
		update     bool
	}{
		{
			name:       "Should generate the correct files based on an input yaml",
			contract:   path.Join(testDir, "test.yaml"),
			wantErr:    false,
			filesystem: ActualFilesystem,
			outputDir:  path.Join(testDir, "output"),
			goldenDir:  path.Join("testdata", "golden"),
			update:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.RemoveAll(tt.outputDir)
			outputDirAbsolute, err := filepath.Abs(tt.outputDir)
			if err != nil {
				t.Fatalf("Absolute Path for outputdir not available")
			}
			absGoldenPath, err := filepath.Abs(tt.goldenDir)
			if err != nil {
				t.Fatalf("Absolute Path for golden dir not available")
			}
			openAPIData, err := openapi3.NewLoader().LoadFromFile(tt.contract)
			if err != nil {
				t.Fatalf("error parsing OpenAPI file: %v", err)
			}

			data := map[string]interface{}{
				generator.ProjectName: serviceName,
				generator.OpenAPIData: openAPIData,
				generator.Repository:  "cto-github.cisco.com/NFV-BU/test-service",
			}
			templates, err := generator.LoadTemplates(tt.filesystem)
			if err != nil {
				t.Fatalf("Could not load templates: %v", err)
			}

			cmdutils.GlobalArgs.OutputDir = outputDirAbsolute

			err = generator.GenerateFiles(tt.filesystem,
				generator.WithData(data),
				generator.WithFS(tt.filesystem),
				generator.WithTemplate(templates))
			if err != nil {
				t.Fatalf("Could not generate: %v", err)
			}
			defer os.RemoveAll(tt.outputDir)
			var f fs.WalkDirFunc
			updateGoldenOutput := func(p string, d fs.DirEntry, err error) error {
				outputPath := path.Join(tt.outputDir, p)
				goldenPath := path.Join(tt.goldenDir, p)
				if d.IsDir() {
					if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
						dirToCreate := path.Join(absGoldenPath, p)
						err := os.Mkdir(dirToCreate, 0750)
						if err != nil && !os.IsExist(err) {
							t.Fatalf("Could not create directory:%v", err)
						}
					}
				} else {
					r, err := os.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("Could not open output dir file: %v", err)
					}

					fileToUpdate := path.Join(absGoldenPath, p)
					err = os.WriteFile(fileToUpdate, r, 0666)
					if err != nil {
						t.Fatalf("could not update golden directory: %v", err)
					}
				}
				return nil
			}
			compareToGoldenOutput := func(p string, d fs.DirEntry, err error) error {
				outputPath := path.Join(tt.outputDir, p)
				goldenPath := path.Join(tt.goldenDir, p)
				if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
					t.Fatalf("%v does not exist", goldenPath)
				}

				if !d.IsDir() {
					expected := path.Join(tt.goldenDir, p)
					e, err := os.ReadFile(expected)
					if err != nil {
						t.Fatalf("Could not open expected file: %v", err)
					}

					r, err := os.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("Could not open result file: %v", err)
					}

					if diff := cmp.Diff(string(e), string(r)); diff != "" {
						t.Errorf("output does not match golden file %v: %s", expected, diff)
					}
				}
				return nil
			}

			if tt.update {
				f = updateGoldenOutput
			} else {
				f = compareToGoldenOutput
			}
			err = fs.WalkDir(os.DirFS(tt.outputDir), ".", f)
			if err != nil {
				t.Fatalf("Could not compare output and golden dirs: %v", err)
			}

			if !tt.update {
				// Check if all dirs in Golden directory were created
				fs.WalkDir(os.DirFS(tt.goldenDir), ".",
					func(p string, d fs.DirEntry, err error) error {
						if d.Name() != ".ignore" {
							outputPath := path.Join(tt.outputDir, p)
							if _, err := os.Stat(outputPath); os.IsNotExist(err) {
								t.Fatalf("%v expected, but does not exist", outputPath)
							}
						}
						return nil
					})
			}
		})
	}
}
