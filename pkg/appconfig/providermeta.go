package appconfig

type ProviderMeta struct {
	Loaded     bool                   //invalid if not loaded or during load
	Settings   map[string]interface{} //storage for the settings loaded by the auth
	Precedence int                    //the precedence for which the settings will take effect.
}

func (providerMeta ProviderMeta) GetSettings() map[string]interface{} {
	return providerMeta.Settings
}

func (providerMeta ProviderMeta) GetPrecedence() int {
	return providerMeta.Precedence
}

func (providerMeta ProviderMeta) IsLoaded() bool {
	return providerMeta.Loaded
}
