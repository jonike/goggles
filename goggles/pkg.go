package goggles

import (
	"bytes"
	"fmt"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/KyleBanks/depth"
)

// Pkg represents a go source package.
type Pkg struct {
	depth.Pkg

	files *token.FileSet
	Docs  struct {
		Name      string    `json:"name"`
		Import    string    `json:"import"`
		Package   string    `json:"package"`
		Constants string    `json:"constants"`
		Variables string    `json:"variables"`
		Functions string    `json:"functions"`
		Types     []PkgType `json:"types"`
	} `json:"docs"`
}

type PkgType struct {
	Name      string `json:"name"`
	Header    string `json:"header"`
	Constants string `json:"constants"`
	Variables string `json:"variables"`
	Functions string `json:"functions"`
}

// makeDocs retrieves the documentation for a package and attaches it to the Pkg.
func (p *Pkg) makeDocs() error {
	p.files = token.NewFileSet()
	doc, err := p.parseDocs()
	if err != nil {
		return err
	}

	p.Docs.Name = doc.Name
	p.Docs.Import = fmt.Sprintf("import \"%v\"", p.Name)
	p.Docs.Package = strings.TrimSpace(doc.Doc)
	p.Docs.Constants = p.printValues(doc.Consts)
	p.Docs.Variables = p.printValues(doc.Vars)
	p.Docs.Functions = p.printFuncs(doc.Funcs)
	p.Docs.Types = p.printTypes(doc.Types)

	return nil
}

// parseDocs parses the package documentation.
func (p *Pkg) parseDocs() (*doc.Package, error) {
	filter := func(file os.FileInfo) bool {
		name := file.Name()
		return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}

	pkgs, err := parser.ParseDir(p.files, filepath.Join(srcdir(), p.Name), filter, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		return doc.New(pkg, ".", 0), nil
	}

	return nil, nil
}

func (p *Pkg) printValues(vals []*doc.Value) string {
	var b bytes.Buffer
	for _, v := range vals {
		fmt.Fprintf(&b, "```\n%s\n%s\n```\n", p.printToken(v.Decl), p.printToken(v.Doc))
	}
	return b.String()
}

func (p *Pkg) printFuncs(funcs []*doc.Func) string {
	var b bytes.Buffer
	for _, f := range funcs {
		var receiver string
		if f.Recv != "" {
			receiver = fmt.Sprintf("(%s)", f.Recv)
		}

		//fmt.Fprintf(&b, "###func %s %s\n%s\n```\n%s\n```\n", receiver, f.Name, f.Doc, p.printToken(f.Decl))
		println(receiver)
		fmt.Fprintf(&b, "%s\n```\n%s\n```\n", f.Doc, p.printToken(f.Decl))
	}

	return b.String()
}

func (p *Pkg) printTypes(types []*doc.Type) []PkgType {
	var pkgTypes []PkgType

	for _, t := range types {
		pkgTypes = append(pkgTypes, PkgType{
			Name:      t.Name,
			Header:    fmt.Sprintf("type %v", t.Name),
			Constants: p.printValues(t.Consts),
			Variables: p.printValues(t.Vars),
			Functions: p.printFuncs(t.Funcs),
		})
	}

	return pkgTypes
}

func (p *Pkg) printToken(t interface{}) string {
	var b bytes.Buffer
	conf := printer.Config{
		Mode:     printer.TabIndent,
		Tabwidth: 4,
	}
	err := conf.Fprint(&b, p.files, t)
	if err != nil {
		return ""
	}

	return b.String()
}