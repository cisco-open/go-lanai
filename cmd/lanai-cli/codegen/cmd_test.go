// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"
)

type TestPlan struct {
	name       string
	configPath string
	wantErr    bool
	outputDir  string
	goldenDir  string
	update     bool
}

func TestGenerateTemplates(t *testing.T) {
	plans := []TestPlan{
		{
			name:       "TestV2Configuration",
			configPath: "testdata/test-codegen-v2.yml",
			wantErr:    false,
			outputDir:  "testdata/output/v2",
			goldenDir:  "testdata/golden/v2",
			update:     false,
		},
		{
			name:       "TestV1Configuration",
			configPath: "testdata/test-codegen-v1.yml",
			wantErr:    false,
			outputDir:  "testdata/output/v1",
			goldenDir:  "testdata/golden/v1",
			update:     false,
		},
		{
			name:       "TestV2OPAConfiguration",
			configPath: "testdata/test-codegen-opa.yml",
			wantErr:    false,
			outputDir:  "testdata/output/opa",
			goldenDir:  "testdata/golden/opa",
			update:     false,
		},
		{
			name:       "TestV2NoSecConfiguration",
			configPath: "testdata/test-codegen-nosec.yml",
			wantErr:    false,
			outputDir:  "testdata/output/nosec",
			goldenDir:  "testdata/golden/nosec",
			update:     false,
		},
	}

	subTests := make([]test.Options, len(plans))
	for i := range plans {
		subTests[i] = test.GomegaSubTest(SubTestCodeGenerate(&plans[i]), plans[i].name)
	}

	test.RunTest(context.Background(), t,
		subTests...,
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestCodeGenerate(tt *TestPlan) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		prepareOutputDir(t, g, tt.outputDir)
		outputDirAbsolute, err := filepath.Abs(tt.outputDir)
		g.Expect(err).To(Succeed(), "Absolute Path for output dir should be available")

		cmdutils.GlobalArgs.OutputDir = outputDirAbsolute
		err = GenerateWithConfigPath(ctx, tt.configPath)
		if err != nil {
			t.Fatalf("Could not generate: %v", err)
		}

		//defer os.RemoveAll(tt.outputDir)
		var f fs.WalkDirFunc
		if tt.update {
			f = updateGoldenOutputFunc(t, tt)
		} else {
			f = compareToGoldenOutputFunc(t, tt)
		}
		err = fs.WalkDir(os.DirFS(tt.outputDir), ".", f)
		if err != nil {
			t.Fatalf("Could not compare output and golden dirs: %v", err)
		}

		if !tt.update {
			// Check if all dirs in Golden directory were created
			e := fs.WalkDir(os.DirFS(tt.goldenDir), ".",
				func(p string, d fs.DirEntry, err error) error {
					if d.Name() != ".ignore" && d.Name() != ".DS_Store" {
						outputPath := path.Join(tt.outputDir, p)
						if _, err := os.Stat(outputPath); os.IsNotExist(err) {
							return fmt.Errorf("%v expected, but does not exist", outputPath)
						}
					}
					return nil
				})
			g.Expect(e).To(Succeed(), "All directories should be created")
		}
	}
}

/*************************
	Helpers
 *************************/

func prepareOutputDir(_ *testing.T, g *gomega.WithT, dir string) {
	err := os.RemoveAll(dir)
	g.Expect(err).To(Succeed())
	err = os.MkdirAll(dir, 0755)
	g.Expect(err).To(Succeed())
	contents, e := os.ReadDir(dir)
	g.Expect(e).To(Succeed())
	g.Expect(contents).To(BeEmpty())
}

func updateGoldenOutputFunc(t *testing.T, tt *TestPlan) func(p string, d fs.DirEntry, err error) error {
	return func(p string, d fs.DirEntry, err error) error {
		absGoldenPath, err := filepath.Abs(tt.goldenDir)
		if err != nil {
			t.Fatalf("Absolute Path for golden dir not available")
		}
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
}

func compareToGoldenOutputFunc(t *testing.T, tt *TestPlan) func(p string, d fs.DirEntry, err error) error {
	return func(p string, d fs.DirEntry, err error) error {

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
}
