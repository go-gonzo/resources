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
	return func(ctx context.Context, files <-chan gonzo.File, out chan<- gonzo.File) error {

		ctx, cancel := context.WithCancel(ctx)

		res := resources.New()
		res.Config = resources.Config(config)

		var err error 

		buff := &bytes.Buffer{}
		for {
			select {
			case file, ok := <-files:
				if !ok {
					goto BUILD
				}

				if file.FileInfo().IsDir() {
					continue
				}
				path, _ := filepath.Rel(file.FileInfo().Base(), file.FileInfo().Name())
				res.Add(filepath.ToSlash(path), file)
				ctx.Infof("Adding %s", path)
				defer func(path string) {
					ctx.Debug("Closing %s", path)
					file.Close() //Close files AFTER we have build our package.
				}(path)
			case <-ctx.Done():
				err = ctx.Err()
				goto BUILD
			}
		}

		BUILD:
		if err != nil {
			return err
		}

		ctx.Debug("Runnig build...")
		err = res.Build(buff)
		if err != nil {
			cancel()
			return err
		}
		path := fmt.Sprintf(FilenameFormat, strings.ToLower(config.Var))
		sf := gonzo.NewFile(ioutil.NopCloser(buff), gonzo.NewFileInfo())
		sf.FileInfo().SetName(path)
		sf.FileInfo().SetSize(int64(buff.Len()))
		out <- sf
		return nil
	}
}
