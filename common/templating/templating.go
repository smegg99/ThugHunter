// common/templating/templating.go
package templating

import (
	"bytes"
	"fmt"
	"text/template"
)

// Resolve renders a Go text/template string with the given data.
// data can be any struct or map; its exported fields / keys become
// the {{.PLACEHOLDER}} variables.
func Resolve(tmplStr string, data any) (string, error) {
	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", tmplStr, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", tmplStr, err)
	}
	return buf.String(), nil
}
