package rndr

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/observatorium/rndr/pkg/rndr/engines/helm"
	"github.com/pkg/errors"
)

type Package struct {
	OutputDir  string  `yaml:"outputDir"`

	// One of.
	OLM               *OLMPackage
	KubeOperator      *KubeOperatorPackage  `yaml:"kubeOperator"`
	Helm              *helm.PackageOptions
	OpenshiftTemplate *OpenshiftTemplatesPackage `yaml:"openshiftTemplate"`
}

type OLMPackage struct {

}

type KubeOperatorPackage struct {

}

type OpenshiftTemplatesPackage struct {
	Values string
}


// RenderPackage renders package.
func RenderPackage(ctx context.Context, logger log.Logger, name, author string, s Package , overrOutDir *string) (err error) {
	outDir := s.OutputDir
	if overrOutDir != nil {
		outDir = *overrOutDir
	}

	switch {
	case s.OLM != nil:
		return errors.Errorf("Operator Lifecycle Manager packaging is not implemented")
	case s.KubeOperator != nil:
		return errors.Errorf("kubernestes operator packaging is not implemented")
	case s.OpenshiftTemplate != nil:
		return errors.Errorf("openshift templates packaging is not implemented")
	case s.Helm != nil:
		err = helm.Package(ctx,  logger, name, author, *s.Helm, outDir)
	default:
		return errors.New("packaging has to be specified, got none")
	}
	if err != nil {
		return err
	}
	return nil
}
