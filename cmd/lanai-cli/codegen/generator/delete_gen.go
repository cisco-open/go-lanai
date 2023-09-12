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

// DeleteGenerator will read delete.*.tmpl files, and delete any generated
// code that matches the regex defined in them, as well as any empty directories after the deletion
type DeleteGenerator struct {
	priorityOrder int
	nameRegex     *regexp.Regexp
	templateFS    fs.FS
	outputFS      fs.FS
}

type DeleteOption struct {
	Option
	PriorityOrder int
}

func newDeleteGenerator(opts ...func(opt *DeleteOption)) *DeleteGenerator {
	o := &DeleteOption{
		PriorityOrder: defaultProjectPriorityOrder,
	}
	for _, fn := range opts {
		fn(o)
	}

	return &DeleteGenerator{
		priorityOrder: o.PriorityOrder,
		nameRegex:     regexp.MustCompile("^(?:delete)(.+)(?:.tmpl)"),
		templateFS:    o.TemplateFS,
		outputFS:      o.OutputFS,
	}
}

func (d *DeleteGenerator) Generate(tmplPath string, tmplInfo fs.FileInfo) error {
	if tmplInfo.IsDir() || !d.nameRegex.MatchString(path.Base(tmplPath)) {
		// Skip over it
		return nil
	}

	// Go through the output dir, if anything matches the regex, delete the file
	regexContent, err := fs.ReadFile(d.templateFS, tmplPath)
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
					logger.Infof("delete generator deleting duplicate reference file %v\n", toRemove)
					err := os.Remove(toRemove)
					if err != nil {
						return err
					}
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
				logger.Infof("delete generator deleting empty file %v\n", toRemove)
				err := os.Remove(toRemove)
				if err != nil {
					return err
				}
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
		logger.Infof("delete generator deleting empty dir %v\n", toRemove)
		err := os.Remove(toRemove)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DeleteGenerator) Order() int {
	return defaultDeletePriorityOrder
}
