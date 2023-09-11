package generator

type GenerationContext struct {
	templatePath string
	filename  string
	regenMode RegenMode
	//	 Add the template (template.Template) here
	model interface{}
}

func NewGenerationContext(templatePath string, filename string, regenMode RegenMode, model interface{}) *GenerationContext {
	return &GenerationContext{
		templatePath: templatePath,
		filename:     filename,
		regenMode:    regenMode,
		model:        model,
	}
}
