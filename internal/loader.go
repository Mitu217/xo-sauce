package internal

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Loader ...
type Loader struct {
	StructList func(*ast.File) []*Struct
	FuncList   func(*ast.File) []*Func
}

// NewLoader
func NewLoader() *Loader {
	return &Loader{}
}

func (l Loader) generateCodeFromDB(args *ArgType) error {
	params := []string{args.DSN}
	if args.Out != "" {
		params = append(params, "-o", args.XoPath)
	}
	if args.TemplatePath != "" {
		params = append(params, "--template-path", args.TemplatePath)
	}
	output, err := exec.Command("xo", params...).CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

// loadStructList ...
func (l Loader) loadStructList(f *ast.File) []*Struct {
	ss := []*Struct{}
	ast.Inspect(f, func(n ast.Node) bool {
		gd, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}
		for _, spec := range gd.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				if _, ok := ts.Type.(*ast.StructType); ok {
					s := &Struct{
						Name:    ts.Name.Name,
						Package: f.Name.Name,
					}
					ss = append(ss, s)
				}
			}
		}
		return true
	})
	return ss
}

// loadFuncList ...
func (l Loader) loadFuncList(f *ast.File) []*Func {
	funcList := []*ast.FuncDecl{}
	ast.Inspect(f, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Recv == nil {
			funcList = append(funcList, fd)
		}
		return true
	})

	fns := []*Func{}
	for _, fd := range funcList {
		// get param fields
		params := []*Field{}
		for _, p := range fd.Type.Params.List {
			for _, n := range p.Names {
				param := &Field{}
				param.Name = n.Name
				// field type
				switch p.Type.(type) {
				case *ast.Ident:
					param.Type = p.Type.(*ast.Ident).Name
					// FIXME: for xo
					if param.Type == "XODB" {
						param.Type = fmt.Sprintf("%s.%s", f.Name.Name, param.Type)
					}
				case *ast.SelectorExpr:
					x := p.Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
					sel := p.Type.(*ast.SelectorExpr).Sel.Name
					param.Type = fmt.Sprintf("%s.%s", x, sel)
				}
				params = append(params, param)
			}
		}
		// get return field
		var res *Field
		for _, p := range fd.Type.Results.List {
			r := &Field{}
			// field type
			switch p.Type.(type) {
			case *ast.StarExpr:
				r.Type = p.Type.(*ast.StarExpr).X.(*ast.Ident).Name
				r.IsPtr = true
				r.NilType = "nil"
			case *ast.ArrayType:
				elt := p.Type.(*ast.ArrayType).Elt
				switch elt.(type) {
				case *ast.StarExpr:
					r.Type = elt.(*ast.StarExpr).X.(*ast.Ident).Name
					r.IsPtr = true
				}
				r.IsArray = true
				r.NilType = "nil"
			}
			if r.Type != "" {
				res = r
			}
		}
		fn := &Func{
			Name:    fd.Name.Name,
			Package: f.Name.Name,
			Params:  params,
			Return:  res,
		}
		fns = append(fns, fn)
	}
	return fns
}

// LoadFile ...
func (l Loader) LoadCodes(args *ArgType) error {
	if err := l.generateCodeFromDB(args); err != nil {
		return err
	}

	fis, err := ioutil.ReadDir(args.XoPath)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		// delete table files in black list
		shouldDeleted := false
		for _, b := range args.BlackList {
			if strings.Index(fi.Name(), b) != -1 {
				shouldDeleted = true
				break
			}
		}
		if shouldDeleted {
			if err := os.Remove(filepath.Join(args.XoPath, fi.Name())); err != nil {
				return err
			}
			continue
		}

		// parse file
		name := strings.Split(fi.Name(), ".")[0]
		f, err := parser.ParseFile(fset, filepath.Join(args.XoPath, fi.Name()), nil, 0)
		if err != nil {
			return err
		}

		// load struct
		ss := l.loadStructList(f)
		for _, s := range ss {
			if err := args.ExecuteTemplate(StructTemplate, FileEditableType, name, s); err != nil {
				return err
			}
		}

		// load func
		fns := l.loadFuncList(f)
		for _, fn := range fns {
			if err := args.ExecuteTemplate(FuncTemplate, FileNotEditableType, name, fn); err != nil {
				return err
			}
		}
	}

	return nil
}
