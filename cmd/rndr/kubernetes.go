package main

import (
	"context"

	"github.com/observatorium/rndr/pkg/rndr"
	"github.com/oklog/run"
	"github.com/openproto/protoconfig/go/kingpinv2"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

func registerKubernetesManifestsCommand(cmd *kingpin.CmdClause, g *run.Group, rc rndrConfig) *kingpin.CmdClause {
	c := cmd.Command("manifests", "Generate Kubernetes manifests")
	values := kingpinv2.Flag(c, "values", "Values YAML as defined in passed --template api").Required().PathOrContent()

	c.Action(func(_ *kingpin.ParseContext) error {
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			bTmpl, err := rc.tmpl.Content()
			if err != nil {
				return err
			}

			t, err := rndr.ParseTemplate(bTmpl)
			if err != nil {
				return err
			}

			vTmpl, err := values.Content()
			if err != nil {
				return err
			}
			return rndr.Render(ctx, rc.logger, t, vTmpl, rc.outDir)
		}, func(err error) {
			cancel()
		})
		return nil
	})
	return c
}

func registerKubernetesOperatorCommand(cmd *kingpin.CmdClause, g *run.Group, rc rndrConfig) *kingpin.CmdClause {
	c := cmd.Command("operator", "Generate Kubernetes operator")

	// TODO(bwplotka): Allow building all into Go binary? What if files are too large to be part of binary?
	c.Action(func(_ *kingpin.ParseContext) error {
		_, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return errors.New("not implemented")
		}, func(err error) {
			cancel()
		})
		return nil
	})
	return c
}

func registerKubernetesHelmCommand(cmd *kingpin.CmdClause, g *run.Group, rc rndrConfig) *kingpin.CmdClause {
	c := cmd.Command("helm", "Generate Helm operator")
	c.Action(func(_ *kingpin.ParseContext) error {
		_, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			return errors.New("not implemented")
		}, func(err error) {
			cancel()
		})
		return nil
	})
	return c
}
