package testutils

import (
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestPackageDirectory(t *testing.T) {
	RuntimeTest = TestPackageDirectory
	g := gomega.NewWithT(t)
	dir := PackageDirectory()
	g.Expect(dir).To(HaveSuffix("test/utils"))
}

func TestProjectDirectory(t *testing.T) {
	RuntimeTest = TestPackageDirectory
	g := gomega.NewWithT(t)
	dir := ProjectDirectory()
	g.Expect(dir).To(HaveSuffix("go-lanai"))
}
