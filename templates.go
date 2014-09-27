package templates

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultViewsDir = "views" // directory containing the base html templates
)

// Templates represent the parsed html templates
type Templates struct {
	Templates  map[string]*template.Template
	Extensions map[string]bool
	ViewsDir   string
	Dir        string

	Views    map[string]string
	Partials map[string]string
}

// New constructs a Templates type
func New() *Templates {
	t := &Templates{
		Templates: make(map[string]*template.Template),
		ViewsDir:  DefaultViewsDir,
		Views:     make(map[string]string),
		Partials:  make(map[string]string),
	}

	return t
}

// Get looks up an html template by view name
func (t *Templates) Get(name string) *Template {
	tmpl := &Template{
		templates: t,
		name:      name,
	}

	if _, ok := t.Templates[name]; ok {
		return tmpl
	}

	tmpl.err = fmt.Errorf("template not found")

	return tmpl
}

// Template is a helper type returned by templates.Template to create a user friendly API
type Template struct {
	templates *Templates
	name      string
	err       error
}

// Render executes the template
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	if t.err != nil {
		return t.err
	}

	return t.templates.Templates[t.name].ExecuteTemplate(w, name, data)
}

// AddExts is a helper method to add file extensions to filter template extensions.
// AddExts should be called before Parse
func (t *Templates) AddExts(extensions []string) {
	exts := make(map[string]bool)
	for _, ext := range extensions {
		exts[ext] = true
	}
	t.Extensions = exts
}

// Parse parses the html templates found in the passed in directory
func (t *Templates) Parse(dir string) (*Templates, error) {
	t.Dir = dir
	if err := filepath.Walk(dir, t.parseFile); err != nil {
		return t, err
	}

	if len(t.Views) == 0 {
		return t, fmt.Errorf("no views were found")
	}

	// create view templates
	for name, tmpl := range t.Views {
		var err error
		t.Templates[name], err = template.New(name).Parse(tmpl)
		if err != nil {
			return t, err
		}
	}

	// add partials to the view templates
	for _, baseTmpl := range t.Templates {
		for name, tmpl := range t.Partials {
			var err error
			baseTmpl, err = baseTmpl.New(name).Parse(tmpl)
			if err != nil {
				return t, err
			}
		}
	}

	return t, nil
}

func (t *Templates) parseFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	ext := filepath.Ext(f.Name())
	if f.IsDir() || !t.check(ext) {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	subPath := strings.Replace(path, t.Dir+"/", "", 1)
	if strings.HasPrefix(subPath, t.ViewsDir+"/") {
		t.Views[subPath] = string(contents)
	} else {
		t.Partials[subPath] = string(contents)
	}

	return nil
}

// checkExt is a helper function to check if the passed in extension exist
func (t *Templates) check(ext string) bool {
	if len(t.Extensions) == 0 {
		return true
	}

	for x := range t.Extensions {
		if ext == x {
			return true
		}
	}

	return false
}
