package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"
)

// CleanupGenerator will read delete.*.tmpl files, and delete any generated
// code that matches the regex defined in them, as well as any empty directories after the deletion
type CleanupGenerator struct {
	order      int
	nameRegex  *regexp.Regexp
	templateFS fs.FS
}

type CleanupOption struct {
	GeneratorOption
	Order int
}

func newCleanupGenerator(gOpt GeneratorOption, opts ...func(opt *CleanupOption)) *CleanupGenerator {
	o := &CleanupOption{
		GeneratorOption: gOpt,
	}
	for _, fn := range opts {
		fn(o)
	}

	return &CleanupGenerator{
		order:      o.Order,
		nameRegex:  regexp.MustCompile("^(?:delete)(.+)(?:.tmpl)"),
		templateFS: o.TemplateFS,
	}
}

func (g *CleanupGenerator) Generate(tmplPath string, tmplInfo fs.FileInfo) error {
	if tmplInfo.IsDir() || !g.nameRegex.MatchString(path.Base(tmplPath)) {
		// Skip over it
		return nil
	}

	// Go through the output dir, if anything matches the regex, delete the file
	regexContent, err := fs.ReadFile(g.templateFS, tmplPath)
	if err != nil {
		return err
	}
	deleteRegex := regexp.MustCompile(string(regexContent))
	outputFS := os.DirFS(cmdutils.GlobalArgs.OutputDir)
	err = deleteFilesMatchingRegex(outputFS, deleteRegex)
	if err != nil {
		return err
	}

	err = deleteDuplicateReferenceFiles(outputFS)
	if err != nil {
		return err
	}

	err = deleteEmptyDirectories(outputFS)
	if err != nil {
		return err
	}
	return nil
}

func deleteDuplicateReferenceFiles(outputFS fs.FS) error {
	if err := fs.WalkDir(outputFS, ".", func(p string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if strings.HasSuffix(p, "ref") {
				originalFile := strings.TrimRight(p, "ref")
				originalContent, err := fs.ReadFile(outputFS, originalFile)
				if err != nil {
					return err
				} else if originalContent == nil {
					return nil
				}

				content, err := fs.ReadFile(outputFS, p)
				if err != nil {
					return err
				} else if content == nil {
					return nil
				}

				if string(content) == string(originalContent) {
					toRemove := fmt.Sprintf("%v/%v", cmdutils.GlobalArgs.OutputDir, p)
					logger.Debugf("[Cleanup] deleting duplicate reference file %v", toRemove)
					err := os.Remove(toRemove)
					if err != nil {
						return err
					}
					globalCounter.Cleanup(toRemove)
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func deleteFilesMatchingRegex(outputFS fs.FS, deleteRegex *regexp.Regexp) error {
	if err := fs.WalkDir(outputFS, ".", func(p string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			content, err := fs.ReadFile(outputFS, p)
			if err != nil {
				return err
			} else if content == nil {
				return nil
			}
			if deleteRegex.MatchString(string(content)) {
				toRemove := fmt.Sprintf("%v/%v", cmdutils.GlobalArgs.OutputDir, p)
				logger.Debugf("[Cleanup] deleting empty file %v", toRemove)
				err := os.Remove(toRemove)
				if err != nil {
					return err
				}
				globalCounter.Cleanup(toRemove)
			}
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func deleteEmptyDirectories(outputFS fs.FS) error {
	var emptyDirList []string
	if err := fs.WalkDir(outputFS, ".", func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			entries, err := fs.ReadDir(outputFS, p)
			if len(entries) == 0 {
				toRemove := fmt.Sprintf("%v/%v", cmdutils.GlobalArgs.OutputDir, p)
				emptyDirList = append(emptyDirList, toRemove)
			}
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	for _, toRemove := range emptyDirList {
		logger.Debugf("[Cleanup] deleting empty dir %v", toRemove)
		err := os.Remove(toRemove)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *CleanupGenerator) Order() int {
	return g.order
}
