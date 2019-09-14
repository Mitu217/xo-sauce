package internal

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/knq/snaker"
)

// NewTemplateFuncs returns a set of template funcs bound to the supplied args.
func (a *ArgType) NewTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"goparamlist":  a.goparamlist,
		"goreturnlist": a.goreturnlist,
		"reniltype":    a.reniltype,
		"retype":       a.retype,
	}
}

// goparamlist converts a list of fields into their named Go parameters,
// skipping any Field with Name contained in ignoreNames. addType will cause
// the go Type to be added after each variable name. addPrefix will cause the
// returned string to be prefixed with ", " if the generated string is not
// empty.
//
// Any field name encountered will be checked against goReservedNames, and will
// have its name substituted by its corresponding looked up value.
//
// Used to present a comma separated list of Go variable names for use with as
// either a Go func parameter list, or in a call to another Go func.
// (ie, ", a, b, c, ..." or ", a T1, b T2, c T3, ...").
func (a *ArgType) goparamlist(fields []*Field, addType bool, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	i := 0
	vals := []string{}
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		s := "v" + strconv.Itoa(i)
		if len(f.Name) > 0 {
			n := strings.Split(snaker.CamelToSnake(f.Name), "_")
			s = strings.ToLower(n[0]) + f.Name[len(n[0]):]
		}

		// add the go type
		if addType {
			s += " " + f.Type
		}

		// add to vals
		vals = append(vals, s)

		i++
	}

	// concat generated values
	return strings.Join(vals, ", ")
}

func (a *ArgType) goreturnlist(fields []*Field, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	vals := []string{}
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		s := f.Type
		vals = append(vals, s)
	}

	// concat generated values
	return strings.Join(vals, ", ")
}

// retype checks typ against known types, and prefixing
// ArgType.CustomTypePackage (if applicable).
func (a *ArgType) retype(res *Field) string {
	prefix := ""
	// add array symbol
	if res.IsArray {
		prefix = prefix + "[]"
	}
	// add ptr symbol
	if res.IsPtr {
		prefix = prefix + "*"
	}
	return prefix + res.Type
}

// reniltype checks typ against known nil types (similar to retype), prefixing
// ArgType.CustomTypePackage (if applicable).
func (a *ArgType) reniltype(res *Field) string {
	return res.NilType
}
