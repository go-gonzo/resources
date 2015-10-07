package resources

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/omeid/gonzo/context"

	"github.com/omeid/go-resources"
	"github.com/omeid/gonzo"
)

type Config resources.Config

var FilenameFormat = "%s_resource.go"

// A build stage creates a new Package and adds all the files coming through the channel to
// the package and returns the result of build as a File on the output channel.
func Build(config Config) gonzo.Stage {
	return func(ctx context.Context, in <-chan gonzo.File, out chan<- gonzo.File) error {

		res, err := build(ctx, in, config)

		var buff *bytes.Buffer
		err = res.Build(buff)
		if err != nil {
			return err
		}

		path := fmt.Sprintf(FilenameFormat, strings.ToLower(res.Var))
		sf := gonzo.NewFile(ioutil.NopCloser(buff), gonzo.NewFileInfo())
		sf.FileInfo().SetName(path)
		sf.FileInfo().SetSize(int64(buff.Len()))
		out <- sf
		return nil
	}
}

func build(ctx context.Context, files <-chan gonzo.File, config Config) (*resources.Package, error) {

	res := resources.New()
	res.Config = resources.Config(config)
	for {
		select {
		case file, ok := <-files:
			if !ok {
				return res, nil
			}
			path, _ := filepath.Rel(file.FileInfo().Base(), file.FileInfo().Name())
			res.Add(filepath.ToSlash(path), file)
			ctx.Infof("Adding %s.\n", path)
			defer file.Close() //Close files AFTER we have build our package.
		case <-ctx.Done():
			return res, ctx.Err()
		}
	}
}
