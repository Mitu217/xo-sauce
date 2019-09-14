package internal

// TemplateType represents a template type.
type TemplateType uint

// the order here will be the alter the output order per file.
const (
	StructTemplate TemplateType = iota
	FuncTemplate
)

// String returns the name for the associated template type.
func (tt TemplateType) String() string {
	var s string
	switch tt {
	case StructTemplate:
		s = "xosauce_struct"
	case FuncTemplate:
		s = "xosauce_func"
	default:
		panic("unknown TemplateType")
	}
	return s
}

// Editable
type EditableType bool

const (
	FileEditableType    EditableType = true
	FileNotEditableType EditableType = false
)

// Suffix ...
func (e EditableType) FileSuffix() string {
	var s string
	switch e {
	case FileEditableType:
		s = ".go"
	case FileNotEditableType:
		s = ".generated.go"
	}
	return s
}

// Package ...
func (e EditableType) HeaderTemplate() string {
	var s string
	switch e {
	case FileEditableType:
		s = "xosauce_package_editable.go.tpl"
	case FileNotEditableType:
		s = "xosauce_package_noteditable.go.tpl"
	default:
		panic("unknown PackageType")
	}
	return s
}

// Field contains field information.
type Field struct {
	Name    string
	Type    string
	NilType string
	IsArray bool
	IsPtr   bool
}

// Struct ...
type Struct struct {
	Package string
	Name    string
	Fields  []*Field
}

// Func ...
type Func struct {
	Package string
	Name    string
	Params  []*Field
	Return  *Field
}
