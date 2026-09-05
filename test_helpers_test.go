package main

import "html/template"

func templateForTest() *template.Template {
	return template.Must(template.New("home").Parse("{{.Tasks}}"))
}
