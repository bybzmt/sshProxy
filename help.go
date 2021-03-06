package main

import (
	"strings"
	"text/template"

	flaggy "github.com/integrii/flaggy"
)

// defaultHelpTemplate is the help template used by default
// {{if (or (or (gt (len .StringFlags) 0) (gt (len .IntFlags) 0)) (gt (len .BoolFlags) 0))}}
// {{if (or (gt (len .StringFlags) 0) (gt (len .BoolFlags) 0))}}
const defaultHelpTemplate = `{{.CommandName}}{{if .Description}} - {{.Description}}{{end}}{{if .PrependMessage}}
{{.PrependMessage}}{{end}}
{{if .UsageString}}
  Usage:
    {{.UsageString}}{{end}}{{if .Positionals}}

  Positional Variables: {{range .Positionals}}
    {{.Name}}{{if .Required}} (Required){{end}}{{if .Description}} - {{.Description}}{{end}}{{if .DefaultValue}} (default: {{.DefaultValue}}){{end}}{{end}}{{end}}{{if .Subcommands}}

  Subcommands: {{range .Subcommands}}
    {{.LongName}}{{if .ShortName}} ({{.ShortName}}){{end}}{{if .Position}}{{if gt .Position 1}}  (position {{.Position}}){{end}}{{end}}{{if .Description}} - {{.Description}}{{end}}{{end}}
{{end}}{{if (gt (len .Flags) 0)}}
  Flags: {{if .Flags}}{{range .Flags}}
    {{if .ShortName}}-{{.ShortName}} {{else}}   {{end}}{{if .LongName}}--{{.LongName}} {{end}}{{if .Description}} {{Indent $ .LongName}}{{.Description}}{{if .DefaultValue}} (default: {{.DefaultValue}}){{end}}{{end}}{{end}}{{end}}
{{end}}{{if .AppendMessage}}{{.AppendMessage}}
{{end}}{{if .Message}}
{{.Message}}{{end}}
`

func indexHelper(d flaggy.Help, longName string) string {
	max := 0
	for _, f := range d.Flags {
		if len(f.LongName) > max {
			max = len(f.LongName)
		}
	}

	l := max - len(longName)
	return strings.Repeat(" ", l)
}

func newHelpTemplate() *template.Template {
	fn := map[string]interface{}{"Indent": indexHelper}

	tmpl, err := template.New("Help").Funcs(fn).Parse(defaultHelpTemplate)
	if err != nil {
		panic(err)
	}

	return tmpl
}
