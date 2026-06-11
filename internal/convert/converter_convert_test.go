package convert

import (
	"context"
	"strings"
	"testing"
	"text/template"
)

func TestExecuteTemplate(t *testing.T) {
	t.Parallel()

	tmpl := template.Must(template.New("test").Parse("{{.}}"))

	tests := []struct {
		name     string
		ctx      context.Context
		tmpl     *template.Template
		tmplName string
		data     any
		wantErr  bool
	}{
		{
			name:     "successful execution",
			ctx:      context.Background(),
			tmpl:     tmpl,
			tmplName: "test",
			data:     "hello",
			wantErr:  false,
		},
		{
			name:     "cancelled context",
			ctx:      func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			tmpl:     tmpl,
			tmplName: "test",
			data:     "hello",
			wantErr:  true,
		},
		{
			name:     "template execution failure",
			ctx:      context.Background(),
			tmpl:     template.Must(template.New("test").Parse("{{.Undefined}}")),
			tmplName: "test",
			data:     struct{}{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf strings.Builder
			err := executeTemplate(tt.ctx, &buf, tt.tmpl, tt.tmplName, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
