package templates

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/james-maloney/bufpool"
)

var (
	// Views is a default Templates instance
	Views = New()
)

func New() *Templates {
	return &Templates{
		Templates:  map[string]*template.Template{},
		views:      map[string]string{},
		partials:   map[string]string{},
		funcs:      template.FuncMap{},
		hasPartial: map[string]map[string]bool{},
		pool:       bufpool.New(),
	}
}

// AddView adds a new template that has access to all partial templates
func AddView(name string, tmpl string) {
	Views.AddView(name, tmpl)
}

// AddPartial adds a template that the view templates will have access too.
func AddPartial(name string, tmpl string) {
	Views.AddPartial(name, tmpl)
}

// AddFunc adds a template function to all views
func AddFunc(name string, f interface{}) {
	Views.AddFunc(name, f)
}

func RenderBytes(baseView, view string, data interface{}) ([]byte, error) {
	return Views.RenderBytes(baseView, view, data)
}

func MustRender(w http.ResponseWriter, baseView, view string, data interface{}) {
	Views.MustRender(w, baseView, view, data)
}

func MustRenderOne(w http.ResponseWriter, view string, data interface{}) {
	Views.MustRenderOne(w, view, data)
}

// Templates represents a collection of templates
type Templates struct {
	Templates  map[string]*template.Template
	Extensions map[string]bool

	dir         string
	stripPrefix string
	views       map[string]string
	partials    map[string]string
	hasPartial  map[string]map[string]bool
	funcs       template.FuncMap

	pool *bufpool.Pool
}

// AddView adds a new template that has access to all partial templates
func (t *Templates) AddView(name string, tmpl string) {
	t.Templates[name] = template.Must(template.New(name).Funcs(t.funcs).Parse(tmpl))
	t.hasPartial[name] = map[string]bool{}
	t.addPartials()
}

// AddPartial adds a template that the view templates will have access too.
func (t *Templates) AddPartial(name string, tmpl string) {
	t.partials[name] = tmpl
	t.addPartials()
}

func (t *Templates) AddFunc(name string, f interface{}) {
	t.funcs[name] = f
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

// ParseDir adds templates and partials from a directory of template files
func (t *Templates) ParseDir(dir string, stripPrefix string) (*Templates, error) {
	t.dir = dir
	t.stripPrefix = stripPrefix
	if err := filepath.Walk(dir, t.parseFile); err != nil {
		return t, err
	}

	if len(t.views) == 0 || len(t.partials) == 0 {
		return t, fmt.Errorf("no views were found")
	}

	t.addPartials()
	for name, tmpl := range t.views {
		t.AddView(name, tmpl)
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

	subPath := strings.Replace(path, t.stripPrefix, "", 1)
	if strings.Contains(path, "/view/") || strings.Contains(path, "/views/") {
		t.views[subPath] = string(contents)
	} else {
		t.partials[subPath] = string(contents)
	}

	return nil
}

// check is a helper function to check if the passed in extension exist
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

// addPartials is a helper function to make sure all views have all partials
func (t *Templates) addPartials() {
	for bName, bTmpl := range t.Templates {
		for pName, pTmpl := range t.partials {
			if _, ok := t.hasPartial[bName][pName]; !ok {
				t.hasPartial[bName][pName] = true
				bTmpl = template.Must(bTmpl.New(pName).Funcs(t.funcs).Parse(pTmpl))
			}
		}
	}
}

// MustRender will render the template to w. Any errors will panic, this should only be used if you are
// recovering from panics
func (t *Templates) MustRender(w http.ResponseWriter, baseView, view string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	b, err := t.RenderBytes(baseView, view, data)
	if err != nil {
		panic(err.Error())
	}

	w.Write(b)
}

func (t *Templates) MustRenderOne(w http.ResponseWriter, view string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf := t.pool.Get()
	tmpl, ok := t.Templates[view]
	if !ok {
		t.pool.Put(buf)
		panic(fmt.Sprintf("templates: '%s' not found", view))
	}

	if err := tmpl.Execute(buf, data); err != nil {
		t.pool.Put(buf)
		panic(fmt.Sprintf("templates: error executing template '%s', error: '%v'", view, err))
	}

	w.Write(buf.Bytes())
	t.pool.Put(buf)
}

func (t *Templates) Execute(w io.Writer, baseView, view string, data interface{}) error {
	tmpl, ok := t.Templates[view]
	if !ok {
		return fmt.Errorf("templates: '%s' not found", view)
	}

	if err := tmpl.ExecuteTemplate(w, baseView, data); err != nil {
		return fmt.Errorf("templates: error executing template '%s', error: '%v'", baseView, err)
	}

	return nil
}

func (t *Templates) ExecuteOne(w io.Writer, view string, data interface{}) error {
	tmpl, ok := t.Templates[view]
	if !ok {
		return fmt.Errorf("templates: '%s' not found", view)
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("templates: error executing template '%s', error: '%v'", view, err)
	}

	return nil
}

func (t *Templates) RenderBytes(baseView, view string, data interface{}) ([]byte, error) {
	buf := t.pool.Get()
	tmpl, ok := t.Templates[view]
	if !ok {
		t.pool.Put(buf)
		return nil, fmt.Errorf("templates: '%s' not found", view)
	}

	if err := tmpl.ExecuteTemplate(buf, baseView, data); err != nil {
		t.pool.Put(buf)
		return nil, fmt.Errorf("templates: error executing template '%s', error: '%v'", baseView, err)
	}

	b := buf.Bytes()
	t.pool.Put(buf)
	return b, nil
}
