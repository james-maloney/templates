/*
Package templates is a thin wrapper around html/template.

The templates package main function is to create a collection
of templates found in a templates directory

The templates directory structure only requires that a 'views' directory exist and it contains at least one HTML template.
An HTML template will be created for each template found in the views directory.

All other views that are not in the 'views' directory will be made available to each view template

Example directory structure

	templates/
		base.html
		views/
			index.html
			about.html
		partials/
			css.html
			nav.html
			scripts.html

Template Examples

templates/base.html

	<!DOCTYPE html>
	<html>
		<head>
			{{template "head" . }}
			<title>{{.Title}}</title>
		</head>
		<body>
			<header>
				<h1 class="logo">Templates Example Site</h1>
				{{ template "partials/nav.html" . }}
			</header>
			{{template "body" . }}
			{{template "footer" . }}
		</body>
	</html>

templates/views/index.html

	{{ define "head" }}
		{{template "partials/css.html" . }}
	{{ end }}

	{{ define "body" }}
		<h2>Hello World</h2>
		<p>A simple index page</p>
	{{ end }}

	{{ define "footer" }}
		{{template "partials/scripts.html" . }}
		<footer>
			About Page Footer
		</footer>
	{{ end }}

templates/partials/nav.html

	{{ if .Menu }}
	<nav>
		<ul>
		{{range $item := .Menu}}
			<li><a {{range $key, $value := $item.Attrs }}{{$key}}="{{$value}}"{{end}}>{{$item.Name}}</a></li>
		{{ end }}
		</ul>
	</nav>
	{{ end }}

Usage:

	// templates collection
	var tmpls *templates.Templates

	// path to template directory
	var templatesPath = "templates/"

	func init() {
		var err error
		templs, err = templates.New().Parse(templatesPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	fund main() {
		// the first method call specifies the 'views/index.html' view and the Render call
		// specifies that the 'base.html' template should be rendered to os.Stdout
		err := tmpls.Get("views/index.html").Render(os.Stdout, "base.html", nil)
		if err != nil {
			// handle error
		}
	}

Example Site

	cd example
	go run main.go -tmpl-dir=`pwd`

View site at http://localhost:8083
*/
package templates
