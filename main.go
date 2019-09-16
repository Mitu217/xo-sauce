package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/mitu217/xo-sauce/internal"
)

func main() {
	args := internal.NewDefaultArgs()
	arg.MustParse(args)
	if err := processArgs(args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	if err := args.Loader.LoadCodes(args); err != nil {
		panic(err)
	}

	if err := writeTypes(args); err != nil {
		panic(err)
	}
}

// processArgs processs cli args.
func processArgs(args *internal.ArgType) error {
	var err error

	// get working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// determine out path
	if args.Out == "" {
		args.Out = cwd
	} else {
		fi, err := os.Stat(args.Out)
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return errors.New("output path is not directory")
		}
	}

	args.Path = args.Out
	args.Package = path.Base(args.Path)
	args.XoPath = args.Out + "/xo"
	args.XoPackage = path.Base(args.XoPath)

	// check user template path
	if args.TemplatePath != "" {
		fi, err := os.Stat(args.TemplatePath)
		if err != nil {
			return errors.New("template path must exist")
		}
		if !fi.IsDir() {
			return errors.New("template path is not directory")
		}
	}

	return nil
}

// files is a map of filenames to open file handles.
var files = map[string]*os.File{}

// getFile builds the filepath from the TBuf information, and retrieves the
// file from files. If the built filename is not already defined, then it calls
// the os.OpenFile with the correct parameters depending on the state of args.
func getFile(args *internal.ArgType, t *internal.TBuf) (*os.File, error) {
	var f *os.File
	var err error

	// determine filename
	filename := strings.ToLower(t.Name) + t.EditableType.FileSuffix()
	filepath := path.Join(args.Path, filename)

	// lookup file
	f, ok := files[filepath]
	if ok {
		return f, nil
	}

	// default open mode
	mode := os.O_RDWR | os.O_CREATE | os.O_TRUNC

	// stat file to determine if file already exists
	fi, err := os.Stat(filepath)
	if err == nil && fi.IsDir() {
		return nil, errors.New("filepath cannot be directory")
	}

	// open file
	f, err = os.OpenFile(filepath, mode, 0666)
	if err != nil {
		return nil, err
	}

	// file didn't originally exist, so add package header
	err = args.TemplateSet().Execute(f, t.EditableType.HeaderTemplate(), args)
	if err != nil {
		return nil, err
	}

	// store file
	files[filepath] = f

	return f, nil
}

// writeTypes writes the generated definitions.
func writeTypes(args *internal.ArgType) error {
	var err error

	out := internal.TBufSlice(args.Generated)

	// sort segments
	sort.Sort(out)

	// loop, writing in order
	for _, t := range out {
		var f *os.File

		// check if generated template is only whitespace/empty
		bufStr := strings.TrimSpace(t.Buf.String())
		if len(bufStr) == 0 {
			continue
		}

		// get file and filename
		f, err = getFile(args, &t)
		if err != nil {
			return err
		}

		// should only be nil when type == xo
		if f == nil {
			continue
		}

		// write segment
		_, err = t.Buf.WriteTo(f)
		if err != nil {
			return err
		}
	}

	// build goimports parameters, closing files
	params := []string{"-w"}
	for k, f := range files {
		params = append(params, k)

		// close
		err = f.Close()
		if err != nil {
			return err
		}
	}

	// process written files with goimports
	output, err := exec.Command("goimports", params...).CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}
