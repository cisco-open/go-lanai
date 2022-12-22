package golden

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	"path/filepath"
	"testing"
	"time"
)

func TestGetGoldenFilePath(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test a standard non subtest test",
			want: filepath.Join("testdata", "golden", "TestGetGoldenFilePath", "test_a_standard_non_subtest_test.json"),
		},
		{
			name: "test_a_standard_non_subtest_test_with_snake_case_already",
			want: filepath.Join("testdata", "golden", "TestGetGoldenFilePath", "test_a_standard_non_subtest_test_with_snake_case_already.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGoldenFilePath(t); got != tt.want {
				t.Errorf("GetGoldenFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithAppTest(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithTimeout(300*time.Second),
		test.GomegaSubTest(SubTestWithoutTableDriven(), "SubTestWithoutTableDriven"),
		test.GomegaSubTest(SubTestWithTableDriven(), "SubTestWithTableDriven"),
	)
}

func SubTestWithoutTableDriven() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		expectedPath := filepath.Join("testdata", "golden", "TestWithAppTest", "SubTestWithoutTableDriven.json")
		if got := GetGoldenFilePath(t); got != expectedPath {
			t.Errorf("GetGoldenFilePath() = %v, want %v", got, expectedPath)
		}
	}
}

func SubTestWithTableDriven() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		tests := []struct {
			name string
			want string
		}{
			{
				name: "test a standard non subtest test",
				want: filepath.Join("testdata", "golden", "TestWithAppTest", "SubTestWithTableDriven", "test_a_standard_non_subtest_test.json"),
			},
			{
				name: "test_a_standard_non_subtest_test_with_snake_case_already",
				want: filepath.Join("testdata", "golden", "TestWithAppTest", "SubTestWithTableDriven", "test_a_standard_non_subtest_test_with_snake_case_already.json"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := GetGoldenFilePath(t); got != tt.want {
					t.Errorf("GetGoldenFilePath() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}
