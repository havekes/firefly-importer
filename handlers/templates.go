package handlers

import (
	"embed"
	"html/template"
	"log"
)

//go:embed templates
var templateFS embed.FS

var Templates *template.Template

func init() {
	var err error
	Templates, err = template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}
}
