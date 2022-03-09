package cmdutils

func LoadYamlConfig(bind interface{}, filepath string, additionalLookupDirs ...string) error {
	_, e := BindYamlFile(bind, filepath, additionalLookupDirs...)
	return e
}

