package rndr

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/efficientgo/tools/core/pkg/logerrcapture"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type rndr struct {
	ctx    context.Context
	logger log.Logger
	t      TemplateDefinition
	outDir string
	values []byte

	tmpDir string
}

func Render(ctx context.Context, logger log.Logger, t TemplateDefinition, values []byte, outDir string) error {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "rndr")
	if err != nil {
		return err
	}
	defer logerrcapture.Do(logger, func() error { return os.RemoveAll(tmpDir) }, "remove tmp dir")

	r := rndr{
		ctx:    ctx,
		logger: logger,
		t:      t,
		outDir: outDir,
		tmpDir: tmpDir,
	}

	// TODO(bwplotka): Parse values & validate through API (!).

	// TODO(bwplotka): Allow passing more parameters (e.g kubernetes options).
	switch {
	case t.Renderer.Jsonnet != nil:
		return r.renderJsonnet(values)
	case t.Renderer.Helm != nil:
		return errors.Errorf("helm renderer is not implemented")
	case t.Renderer.Process != nil:
		return errors.Errorf("process renderer is not implemented")
	default:
		return errors.Errorf("no renderer was specified")
	}
}
