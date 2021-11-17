package data

const (
	_ = iota
	ErrorTranslatorOrderGorm // gorm error -> data error
	ErrorTranslatorOrderData // data error -> data error with status code
)

const (
	GormConfigurerGroup = "gorm_config"
)
