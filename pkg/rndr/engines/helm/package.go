package helm

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type PackageOptions struct {

}

func Package(_ context.Context, logger log.Logger, name, author string, opts PackageOptions, outDir string) (err error) {
	return errors.New("not implemented")

}