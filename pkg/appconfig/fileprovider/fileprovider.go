package fileprovider

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/parser"
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

type ConfigProvider struct {
	appconfig.ProviderMeta
	reader io.Reader
	propertyParser parser.PropertyParser
}

func NewProvider(precedence int, filePath string, reader io.Reader) *ConfigProvider {
	fileExt := strings.ToLower(path.Ext(filePath))
	switch fileExt {
	case ".yml", ".yaml":
		return &ConfigProvider{
			ProviderMeta:   appconfig.ProviderMeta{Precedence: precedence},
			reader:         reader,
			propertyParser: parser.NewYamlPropertyParser(),
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
		fmt.Printf("Unknown appconfig file extension: %s", fileExt)
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

func NewFileProvidersFromBaseName(precedence int, baseName string, ext string) *ConfigProvider {
	fullPath := path.Join(configsDirectory, baseName + "." + ext)
	info, err := os.Stat(fullPath)
	if !os.IsNotExist(err) && !info.IsDir() {
		file, _ := os.Open(fullPath);
		return NewProvider(precedence, fullPath, file)
	}
	return nil
}