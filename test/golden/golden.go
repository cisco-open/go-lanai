// Package golden will contain some utility functions for golden file testing
package golden

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/sergi/go-diff/diffmatchpatch"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// PopulateGoldenFiles will write golden files to the according to the GetGoldenFilePath function
// data should be of a type struct and not []byte or string. The function will
// marshal the data into JSON.
func PopulateGoldenFiles(t *testing.T, data interface{}) {
	t.Errorf("Running PopulateGoldenFiles will result in a failed test.")
	goldenFilePath := GetGoldenFilePath(t)
	b, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		t.Fatalf("unable to marshal to json: %v", err)
	}

	if _, err := os.Stat(goldenFilePath); err == nil {
		t.Fatalf("cannot use this function to update golden files")
	}

	err = os.MkdirAll(filepath.Dir(goldenFilePath), 0755)
	if err != nil {
		t.Fatalf("unable to mkdir to golden file path")
	}
	err = os.WriteFile(goldenFilePath, b, 0644)
	if err != nil {
		t.Fatalf("unable to write golden file: %v", err)
	}
}

// GetGoldenFilePath will typically return the path in the form ./testdata/golden/<sub-test-name>/<table-driven-test-name>.json
// However, if the test is not run in a subtest or table driven test, the path may differ. However, the last portion
// of the path will always become the golden json name.
func GetGoldenFilePath(t *testing.T) string {
	fullName := t.Name()
	splitName := strings.Split(fullName, "/")
	// we expect 3 parts. TestName, SubTest, TableDrivenTest
	goldenFilePath := filepath.Join("testdata", "golden")
	for i, part := range splitName {
		if i == len(splitName)-1 {
			// if this is the last part, use it as the .json
			goldenFilePath = filepath.Join(goldenFilePath, part+".json")
			break
		}
		goldenFilePath = filepath.Join(goldenFilePath, part)
	}
	return goldenFilePath
}

// Assert will assert that the data matches what is in the golden file.
// data should be of a type struct and not []byte or string. The function will
// marshal the data into JSON.
// The diff will be represented in a colored diff
func Assert(t *testing.T, data interface{}) {
	goldenData, err := os.ReadFile(GetGoldenFilePath(t))
	if err != nil {
		t.Fatalf("unable to read golden file: %v", err)
	}
	dataJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		t.Fatalf("unable to marshal to json: %v", err)
	}

	if !cmp.Equal(goldenData, dataJSON) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(goldenData), string(dataJSON), false)
		dmp.PatchMake()
		t.Errorf("[red] missing, [green] extra:\n%v", dmp.DiffPrettyText(diffs))
	}
}
