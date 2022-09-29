package codegen

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

const SpecialDirDelimiter = "@"

// FileSystemMapper reads an embed.FS & maps base template names to locations of where generated code based on the template will be output to
// as well as where that template is located in the srcRoot hierarchy
type FileSystemMapper struct {
	// Map the template name to the output directories any files that use that template will be put
	// e.g inner.tmpl -> <targetRoot>/inner/
	outputDirs map[string]string
	// Maps the template name to the full path within the embedded template FS
	templatePaths map[string]string
	// Generated code that uses templates inside this directory will be output to a directory mirroring
	// it's relative location to this directory
	// e.g templates/srcRoot/inner/inner.tmpl -> <targetRoot>/inner/inner.go
	srcRootDir string
	// Directory for storing templates that can be imported by other templates
	commonDir string

	templates embed.FS
}

func NewFileSystemMapper(templates embed.FS, targetRoot string) (FileSystemMapper, error) {
	ret := FileSystemMapper{
		outputDirs:    make(map[string]string),
		templatePaths: make(map[string]string),
		templates:     templates,
	}

	if srcRootDir, err := getPathForDirInFileSystem("srcRoot", templates); err != nil {
		return FileSystemMapper{}, err
	} else {
		ret.srcRootDir = srcRootDir
	}

	if commonDir, err := getPathForDirInFileSystem("common", templates); err != nil {
		return FileSystemMapper{}, err
	} else {
		ret.commonDir = commonDir
	}

	if err := ret.populateTemplateOutputDestinations(targetRoot); err != nil {
		return FileSystemMapper{}, err
	}

	return ret, nil
}

func getPathForDirInFileSystem(dirName string, templates embed.FS) (string, error) {
	var value string
	fs.WalkDir(templates, ".",
		func(fullPath string, d fs.DirEntry, err error) error {
			if value != "" || !d.IsDir() {
				return nil
			}

			if path.Base(fullPath) == dirName {
				value = fullPath
			}
			return nil
		})
	if value == "" {
		return "", fmt.Errorf("%s directory does not exist in template hierarchy", dirName)
	}
	return value, nil
}

func (f *FileSystemMapper) populateTemplateOutputDestinations(targetRoot string) error {
	return fs.WalkDir(f.templates, f.srcRootDir,
		func(templatePath string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				basename := path.Base(templatePath)
				outputDir := path.Join(targetRoot, interiorFolderStructure(templatePath, f.srcRootDir))
				f.outputDirs[basename] = outputDir
				if f.templatePaths[basename] != "" {
					return fmt.Errorf("template %v already exists in filesystem", basename)
				}
				f.templatePaths[basename] = templatePath
			}
			return nil
		})
}

func interiorFolderStructure(templatePath string, srcRootDir string) string {
	// templates/srcRoot/inner/dirs -> inner/dirs
	return strings.ReplaceAll(path.Dir(templatePath), srcRootDir, "")
}

// GetPathToTemplate returns the full path of the template in the hierarchy
// e.g inner.tmpl -> srcRoot/inner/inner.tmpl
func (f *FileSystemMapper) GetPathToTemplate(tmpl string) string {
	return f.templatePaths[tmpl]
}

// GetOutputFilePath will return the resolved destination path for filename using template tmpl
// Files generated from templates inside `srcRoot` will be output to a similar hierarchy to where the template is located.
// e.g. If you use `srcRoot/inner/inner.tmpl`, then your generated file will be sent to `<targetRoot>/inner/inner.go`
// If modifiers are provided, it will:
// Get the relative hierachy of the template in the srcRoot directory
// Replace all special foldernames (surrounded by @s) with values from modifiers
// e.g modifers containing VERSION == v2:
// templates/srcRoot/inner/@VERSION@/inner.tmpl -> outputdir/inner/v2/inner.go
func (f *FileSystemMapper) GetOutputFilePath(tmpl string, filename string, modifiers map[string]string) string {
	result := path.Join(f.outputDirs[tmpl], filename)
	if modifiers != nil {
		result = replaceSpecialFoldersWithModifiers(result, modifiers)
	}
	return result
}

func replaceSpecialFoldersWithModifiers(unresolvedPath string, modifiers map[string]string) string {
	specialFolders := findSpecialFolders(unresolvedPath)
	if specialFolders == nil {
		return unresolvedPath
	}

	parts := strings.Split(unresolvedPath, "/")
	for _, toReplace := range specialFolders {
		replaceWith := modifierValue(toReplace, modifiers)
		parts = replacePathParts(parts, toReplace, replaceWith)
	}
	return path.Join(parts...)
}

func findSpecialFolders(result string) []string {
	regexString := fmt.Sprintf("(%v[A-Z]\\w+%v)", SpecialDirDelimiter, SpecialDirDelimiter)
	r, err := regexp.Compile(regexString)
	if err != nil {
		return nil
	}

	matches := r.FindStringSubmatch(result)
	if len(matches) <= 1 {
		return nil
	}
	return matches[1:]
}

func modifierValue(toReplace string, modifiers map[string]string) string {
	pathWithoutDelim := strings.ReplaceAll(toReplace, SpecialDirDelimiter, "")
	replaceWith := modifiers[pathWithoutDelim]
	if replaceWith == "" {
		fmt.Printf("%v could not be resolved, replacing with empty value\n", pathWithoutDelim)
	}
	return replaceWith
}

func replacePathParts(parts []string, toReplace string, replaceWith string) []string {
	var newPathParts []string
	for _, part := range parts {
		newPathParts = append(newPathParts, strings.ReplaceAll(part, toReplace, replaceWith))
	}
	return newPathParts
}
