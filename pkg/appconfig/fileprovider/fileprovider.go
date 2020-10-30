package fileprovider

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	configsDirectory = "configs"
)

type PropertyParser func([]byte) (map[string]interface{}, error)

type ConfigProvider struct {
	appconfig.ProviderMeta
	reader io.Reader
	propertyParser PropertyParser
}

func newProvider(description string, precedence int, filePath string, reader io.Reader) *ConfigProvider {
	fileExt := strings.ToLower(path.Ext(filePath))
	switch fileExt {
	case ".yml", ".yaml":
		return &ConfigProvider{
			ProviderMeta:   appconfig.ProviderMeta{Description: description, Precedence: precedence},
			reader:         reader,
			propertyParser: NewYamlPropertyParser(),
		}
	//TODO: impl the following
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
		fmt.Printf("Unknown appconfig file extension: ", fileExt)
		return nil
	}
}

func (configProvider *ConfigProvider) Load() (loadError error) {
	defer func(){
		if loadError != nil {
			configProvider.IsLoaded = false
		} else {
			configProvider.IsLoaded = true
		}
	}()

	encoded, loadError := ioutil.ReadAll(configProvider.reader)
	if loadError != nil {
		return loadError
	}

	settings, loadError := configProvider.propertyParser(encoded)
	if loadError != nil {
		return loadError
	}
	configProvider.Settings = settings

	return nil
}

func NewFileProvidersFromBaseName(description string, precedence int, baseName string, ext string) *ConfigProvider {
	fullPath := path.Join(configsDirectory, baseName + "." + ext)
	info, err := os.Stat(fullPath)
	if !os.IsNotExist(err) && !info.IsDir() {
		file, _ := os.Open(fullPath);
		return newProvider(description, precedence, fullPath, file)
	}
	return nil
}