package main

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/observatorium/rndr/pkg/rndr"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

func registerPackage(cmd *kingpin.Application, g *run.Group, future func() log.Logger) {
	p := cmd.Command("package", "Render package defined in spec.")
	spec := p.Flag("spec", "Path to the YAML file with spec defined in github.com/observatorium/rndr/pkg/rndr.Spec").
		Short('s').Required().ExistingFile()
	overrOutDir := p.Flag("output", "Optional override directory for output. Works only when single output was chosen").
		Short('o').String()
	pkgs := p.Arg("name", "List or single package name from provided spec. If empty all package will be rendered in random order.").Strings()

	p.Action(func(_ *kingpin.ParseContext) error {
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

			if *overrOutDir != "" && len(*pkgs) != 1 {
				return errors.New("output dir override not allowed when more than 1 package is specified")
			}

			chosen := make(map[string]rndr.Package, len(s.Packages))
			if len(*pkgs) == 0 {
				for p, pkg := range s.Packages {
					chosen[p] = pkg
				}
			} else {
				for _, p := range *pkgs {
					pkg, ok := s.Packages[p]
					if !ok {
						return errors.Errorf("package with name %q was specified in arg but not present in spec.", p)
					}
					chosen[p] = pkg

				}
			}

			for p, pkg := range chosen {
				if err := rndr.RenderPackage(ctx, logger, s.Name, s.Authors, pkg, overrOutDir); err != nil {
					return errors.Wrapf(err, "render %v", p)
				}
			}
			return nil
		}, func(err error) {
			cancel()
		})
		return nil
	})
}
