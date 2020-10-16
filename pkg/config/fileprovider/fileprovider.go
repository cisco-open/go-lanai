package fileprovider

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/config"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type ContentReader func() ([]byte, error)

type PropertyParser func(reader ContentReader) (map[string]interface{}, error)

type ConfigProvider struct {
	config.ProviderMeta
	contentReader ContentReader
	propertyParser PropertyParser
}

func fileContentReader(filePath string) ContentReader {
	return func() (bytes []byte, err error) {
		return ioutil.ReadFile(filePath)
	}
}

func newProvider(description string, precedence int, filePath string, reader ContentReader) *ConfigProvider {
	fileExt := strings.ToLower(path.Ext(filePath))
	switch fileExt {
	case ".yml", ".yaml":
		return &ConfigProvider{
			ProviderMeta:   config.ProviderMeta{Description: description, Precedence: precedence},
			contentReader:  reader,
			propertyParser: NewYamlPropertyParser(),
		}
	/*
	case ".ini":
		return NewCachedLoader(NewINIFile(name, fileName, reader))
	case ".json", ".json5":
		return NewCachedLoader(NewJSONFile(name, fileName, reader))
	case ".toml":
		return NewCachedLoader(NewTOMLFile(name, fileName, reader))
	case ".properties":
		return NewCachedLoader(NewPropertiesFile(name, fileName, reader))
	 */
	default:
		fmt.Printf("Unknown config file extension: ", fileExt)
		return nil
	}
}

func (configProvider *ConfigProvider) Load() {
	configProvider.Valid = false

	configProvider.Settings = make(map[string]interface{})

	settings, error := configProvider.propertyParser(configProvider.contentReader)
	//TODO: should error be propagated up?
	if error == nil {
		for k, v := range settings {
			configProvider.Settings[config.NormalizeKey(k)] = v
		}
	}
}

func NewFileProvidersFromBaseName(description string, precedence int, baseName string, ext string) *ConfigProvider {
	fullPath := path.Join(".", baseName + "." + ext)
	info, err := os.Stat(fullPath)
	if !os.IsNotExist(err) && !info.IsDir() {
		return newProvider(description, precedence, fullPath, fileContentReader(fullPath))
	}
	return nil
}