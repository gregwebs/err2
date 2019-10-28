package main

import (
	"flag"
	"os"
	"text/template"
)

type data struct {
	Package     string
	Type        string
	Name        string
	CallPackage string
}

func main() {
	var d data
	flag.StringVar(&d.Package, "package", "err2", "The package name used for the generated file")
	flag.StringVar(&d.Type, "type", "", "The actual type used for the wrapper being generated e.g. *File")
	flag.StringVar(&d.Name, "name", "", "The name used for the helper being generated. This should start with a capital letter so that it is exported.")
	flag.Parse()

	if d.Name == "" || d.Type == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if d.Package != "err2" {
		d.CallPackage = "err2."
	}

	t := template.Must(template.New("queue").Parse(codeTemplate))
	_ = t.Execute(os.Stdout, d)
}

var codeTemplate = `package {{.Package}}

type _{{.Name}} struct{}

// {{.Name}} is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var {{.Name}} _{{.Name}}

// Try is a helper method to call func() ({{.Type}}, error) functions
// with it and be as fast as Check(err).
func (o _{{.Name}}) Try(v {{.Type}}, err error) {{.Type}} {
	{{.CallPackage}}Check(err)
	return v
}
`
