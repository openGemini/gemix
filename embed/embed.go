package embed

import (
	goembed "embed"
	"io/fs"
)

//go:embed templates
var embededFiles goembed.FS

// ReadTemplate read the template file embed.
func ReadTemplate(path string) ([]byte, error) {
	return embededFiles.ReadFile(path)
}

// ReadTemplateDir read the template dirs embed.
func ReadTemplateDir(name string) ([]fs.DirEntry, error) {
	return embededFiles.ReadDir(name)
}

//go:embed examples
var embedExamples goembed.FS

// ReadExample read an example file
func ReadExample(path string) ([]byte, error) {
	return embedExamples.ReadFile(path)
}
