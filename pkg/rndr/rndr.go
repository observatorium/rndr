package rndr

import (
	"context"
	"fmt"

	"github.com/brancz/locutus/render/jsonnet"
	"github.com/brancz/locutus/rollout"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

func Render(ctx context.Context, logger log.Logger, t TemplateDefinition, values []byte, outDir string) error {
	var (
		typName  string
		renderer rollout.Renderer
	)

	// TODO(bwplotka): Allow passing more parameters (e.g kubernetes options).
	switch {
	case t.Renderer.Jsonnet != nil:
		renderer = jsonnet.NewRenderer(logger, t.Renderer.Jsonnet.Entry, nil)
		typName = "jsonnet"
	case t.Renderer.Helm != nil:
		return errors.Errorf("helm renderer is not implemented")
	case t.Renderer.Process != nil:
		return errors.Errorf("process renderer is not implemented")
	default:
		return errors.Errorf("no renderer was specified")
	}

	r, err := renderer.Render(values)
	if err != nil {
		return errors.Wrapf(err, "render %v", typName)
	}

	fmt.Println(r.Objects, r.Rollout)
	return nil
}
