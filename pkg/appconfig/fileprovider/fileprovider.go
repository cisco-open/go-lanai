package fileprovider

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/parser"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"embed"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

var logger = log.New("Config.File")

type ConfigProvider struct {
	appconfig.ProviderMeta
	reader   io.Reader
	filepath string
}

func NewProvider(precedence int, filePath string, reader io.Reader) *ConfigProvider {
	fileExt := strings.ToLower(path.Ext(filePath))
	switch fileExt {
	case ".yml", ".yaml":
		return &ConfigProvider{
			ProviderMeta: appconfig.ProviderMeta{Precedence: precedence},
			reader:       reader,
			filepath:     filePath,
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
		logger.Warnf("Unknown appconfig file extension: %s", fileExt)
		return nil
	}
}

func (configProvider *ConfigProvider) Name() string {
	return fmt.Sprintf("file:%s", configProvider.filepath)
}

func (configProvider *ConfigProvider) Load(_ context.Context) (loadError error) {
	defer func(){
		if loadError != nil {
			configProvider.Loaded = false
		} else {
			configProvider.Loaded = true
		}
	}()

	encoded, loadError := io.ReadAll(configProvider.reader)
	if loadError != nil {
		return loadError
	}

	settings, loadError := parser.NewYamlPropertyParser()(encoded)
	if loadError != nil {
		return loadError
	}
	configProvider.Settings = settings

	return nil
}

func NewFileProvidersFromBaseName(precedence int, baseName string, ext string, conf bootstrap.ApplicationConfig) (provider *ConfigProvider, exists bool) {

	raw := conf.Value(appconfig.PropertyKeyConfigFileSearchPath)
	var searchPaths []string
	switch v := raw.(type) {
	case string:
		searchPaths = []string{v}
	case []string:
		searchPaths = v
	case []interface{}:
		searchPaths = make([]string, len(v))
		for i, elem := range v {
			if s, ok := elem.(string); ok {
				searchPaths[i] = s
			}
		}
	}

	for _, dir := range searchPaths {
		fullPath := path.Join(dir, baseName + "." + ext)
		info, err := os.Stat(fullPath)
		if !os.IsNotExist(err) && !info.IsDir() {
			file, _ := os.Open(fullPath)
			return NewProvider(precedence, fullPath, file), true
		}
	}

	return nil, false
}

func NewEmbeddedFSProvider(precedence int, path string, fs embed.FS) (provider *ConfigProvider, exists bool) {
	file, e := fs.Open(path)
	if e != nil {
		return nil, false
	}
	return NewProvider(precedence, path, file), true
}