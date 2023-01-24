package generator

type GenerationContext struct {
	templatePath string
	filename     string
	regenRule    string
	//	 Add the template (template.Template) here
	model interface{}
}

func NewGenerationContext(templatePath string, filename string, regenRule string, model interface{}) *GenerationContext {
	return &GenerationContext{
		templatePath: templatePath,
		filename:     filename,
		regenRule:    regenRule,
		model:        model,
	}
}
