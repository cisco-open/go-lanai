package log

type Properties struct {
	Levels map[string]LoggingLevel `json:"Levels"`
	Logger map[string]string       `json:"Logger"`
}