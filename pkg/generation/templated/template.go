package templated

import (
	"bytes"
	"fmt"
	"text/template"
)

// TemplateData represents the data structure available in templates
type TemplateData struct {
	Ref map[string]string
}

// RenderTemplate executes a Go template with the provided secret data
// The secret data is made available as .Ref.<key> in the template
func RenderTemplate(templateStr string, secretData map[string][]byte) ([]byte, error) {
	if templateStr == "" {
		return nil, fmt.Errorf("template string cannot be empty")
	}

	// Convert []byte values to strings for easier template usage
	refData := make(map[string]string)
	for key, value := range secretData {
		refData[key] = string(value)
	}

	data := TemplateData{
		Ref: refData,
	}

	// Parse and execute the template
	tmpl, err := template.New("secret").Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
