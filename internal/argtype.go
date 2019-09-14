package internal

// ArgType is the type that specifies the command line arguments.
type ArgType struct {
	// DSN is the database string (ie, pgsql://user@blah:localhost:5432/dbname?args=)
	DSN string `arg:"positional,required,help:data source name"`

	// Out is the output path.
	Out string `arg:"-o,help:output path"`

	// TemplatePath is the path to use the user supplied templates instead of
	// the built in versions.
	TemplatePath string `arg:"--template-path,help:user supplied template path"`

	Path string `arg:"-"`

	Package string `arg:"-"`

	XoPath string `arg:"-"`

	XoPackage string `arg:"-"`

	// Loader is the load model.
	Loader *Loader `arg:"-"`

	// templateSet is the set of templates to use for generating data.
	templateSet *TemplateSet `arg:"-"`

	// Generated is the generated templates after a run.
	Generated []TBuf `arg:"-"`
}

// NewDefaultArgs returns the default arguments.
func NewDefaultArgs() *ArgType {
	return &ArgType{
		Loader: NewLoader(),
	}
}
