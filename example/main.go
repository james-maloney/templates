package main

import (
	"bytes"
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/james-maloney/templates"
)

var (
	// templates global that will contain all of our parsed temlates from the templates directory
	tmpls *templates.Templates
)

var tmplDir = flag.String("tmpl-dir", "templates", "Path to the templates directory")

var (
	css     = []string{"http://foo.com/main.css"}
	scripts = []string{"http://foo.com/main.js"}
)

// parse the templates in the template directory
func init() {
	flag.Parse()

	var err error
	tmpls, err = templates.New().Parse(*tmplDir + "/templates")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	// Index page route
	http.HandleFunc("/", IndexHandler)

	// About page route
	http.HandleFunc("/about", AboutHandler)

	// Start http server
	if err := http.ListenAndServe(":8083", nil); err != nil {
		log.Fatal(err)
	}
}

// IndexHandler serves the index page
func IndexHandler(w http.ResponseWriter, req *http.Request) {
	buf := &bytes.Buffer{}

	// render the index page to buf
	err := tmpls.Get("views/index.html").Render(buf, "base.html", map[string]interface{}{
		"Title":   "Index Page Title",
		"Css":     css,
		"Scripts": scripts,
		"Menu":    activeNav("index"),
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// write the index page
	w.Write(buf.Bytes())
}

// AboutHandler serves the about page
func AboutHandler(w http.ResponseWriter, req *http.Request) {
	buf := &bytes.Buffer{}

	// render the about page to buf
	err := tmpls.Get("views/about.html").Render(buf, "base.html", map[string]interface{}{
		"Title":   "About Page Title",
		"Css":     css,
		"Scripts": scripts,
		"Menu":    activeNav("about"),
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// write the about page
	w.Write(buf.Bytes())
}

type navItem struct {
	Name  string
	Attrs map[template.HTMLAttr]string
}

func activeNav(active string) []navItem {
	// create menu items
	about := navItem{
		Name: "About",
		Attrs: map[template.HTMLAttr]string{
			"href":  "/about",
			"title": "About Page",
		},
	}
	home := navItem{
		Name: "Home",
		Attrs: map[template.HTMLAttr]string{
			"href":  "/",
			"title": "Home Page",
		},
	}

	// set active menu class
	switch active {
	case "about":
		about.Attrs["class"] = "active"
	case "home":
		home.Attrs["class"] = "active"
	}

	return []navItem{home, about}
}
