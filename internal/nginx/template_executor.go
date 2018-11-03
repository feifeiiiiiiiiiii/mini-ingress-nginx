package nginx

import (
	"bytes"
	"html/template"
	"path"
)

// TemplateExecutor executes NGINX configuration templates
type TemplateExecutor struct {
	mainTemplate    *template.Template
	ingressTemplate *template.Template
}

// NewTemplateExecutor create a NewTemplateExecutor
func NewTemplateExecutor(mainTemplatePath string, ingressTemplatePath string) (*TemplateExecutor, error) {
	nginxTemplate, err := template.New(path.Base(mainTemplatePath)).ParseFiles(mainTemplatePath)
	if err != nil {
		return nil, err
	}

	ingressTemplate, err := template.New(path.Base(ingressTemplatePath)).ParseFiles(ingressTemplatePath)
	if err != nil {
		return nil, err
	}

	return &TemplateExecutor{
		mainTemplate:    nginxTemplate,
		ingressTemplate: ingressTemplate,
	}, nil
}

// UpdateIngressTemplate updates the ingress template
func (te *TemplateExecutor) UpdateIngressTemplate(templateString *string) error {
	newTemplate, err := template.New("ingressTemplate").Parse(*templateString)
	if err != nil {
		return err
	}
	te.ingressTemplate = newTemplate

	return nil
}

// ExecuteIngressConfigTemplate generates the content of a NGINX configuration file for an Ingress resource
func (te *TemplateExecutor) ExecuteIngressConfigTemplate(cfg *IngressNginxConfig) ([]byte, error) {
	var configBuffer bytes.Buffer
	err := te.ingressTemplate.Execute(&configBuffer, cfg)

	return configBuffer.Bytes(), err
}

// ExecuteMainConfigTemplate generates the content of the main NGINX configuration file
func (te *TemplateExecutor) ExecuteMainConfigTemplate() ([]byte, error) {
	cfg := &MainConfig{}
	var configBuffer bytes.Buffer
	err := te.mainTemplate.Execute(&configBuffer, cfg)

	return configBuffer.Bytes(), err
}
