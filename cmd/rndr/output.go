package main

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/observatorium/rndr/pkg/kingpinv2"
	"github.com/observatorium/rndr/pkg/rndr"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

func registerOutput(cmd *kingpin.Application, g *run.Group, future func() log.Logger) {
	o := cmd.Command("output", "Render output defined in spec given values.")
	spec := o.Flag("spec", "Path to the YAML file with spec defined in github.com/observatorium/rndr/pkg/rndr.Spec").
		Short('s').Required().ExistingFile()
	outDir := o.Flag("output", "Output directory").Short('o').Default(".gen").ExistingDir()
	values := kingpinv2.Flag(o, "values", "Values YAML as defined in passed --template api").Required().PathOrContent()

	o.Action(func(_ *kingpin.ParseContext) error {
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			logger := future()

			specFile, err := filepath.Abs(*spec)
			if err != nil {
				return errors.Wrap(err, "abs")
			}
			bSpec, err := ioutil.ReadFile(specFile)
			if err != nil {
				return errors.Wrap(err, "read spec file")
			}
			s, err := rndr.ParseSpec(bSpec, filepath.Dir(specFile))
			if err != nil {
				return err
			}

			if s.Template == nil {
				return errors.New("template is not specified. Ref or empty template is not yet supported")
			}

			vYAML, err := values.Content()
			if err != nil {
				return err
			}

			return rndr.RenderTemplate(ctx, logger, s.Name, *s.Template, vYAML, *outDir)
		}, func(err error) {
			cancel()
		})
		return nil
	})
}
