package generator

type GenerationContext struct {
	templatePath string
	filename     string
	//	 Add the template (template.Template) here
	model interface{}
}

func NewGenerationContext(templatePath string, filename string, model interface{}) *GenerationContext {
	return &GenerationContext{
		templatePath: templatePath,
		filename:     filename,
		model:        model,
	}
}
