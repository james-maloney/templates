/*

Package templates is a thin wrapper around html/template.

The templates package main function is to create a collection
of templates found in a templates directory

The templates directory structure only requires that a 'views' directory exist and it contains at least on html template.
An html template will be created for each template found in the views directory.

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

Usage

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
		// the first method call specifies the 'index.html' view and the Render call
		// specifies that the 'base.html' template should be rendered to os.Stdout
		err := tmpls.Template("index").Render(os.Stdout, "base", nil)
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
