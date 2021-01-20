package rndr

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/observatorium/rndr/pkg/rndr/jsonnet"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func Render(ctx context.Context, logger log.Logger, t TemplateDefinition, values []byte, outDir string) (err error) {
	// TODO(bwplotka): Parse values & validate through API (!).
	// TODO(bwplotka): Allow passing more parameters (e.g kubernetes options).
	var objectGroups map[string][]string

	switch {
	case t.Renderer.Jsonnet != nil:
		// TODO(bwplotka): This is a hack to make sure we only accept YAML.
		// Use provided definition (requires dynamic invoke of Go).
		// Something like https://github.com/golang/mock/blob/master/mockgen/mockgen.go#L378.
		v := make(map[string]interface{})
		if err := yaml.Unmarshal(values, v); err != nil {
			return err
		}
		valuesJSON := []byte("{}")
		if len(v) > 0 {
			valuesJSON, err = json.Marshal(v)
			if err != nil {
				return err
			}
		}
		objectGroups, err = jsonnet.Render(logger, t.Name, *t.Renderer.Jsonnet, valuesJSON)
	case t.Renderer.Helm != nil:
		return errors.Errorf("helm renderer is not implemented")
	case t.Renderer.Process != nil:
		return errors.Errorf("process renderer is not implemented")
	default:
		return errors.Errorf("no renderer was specified")
	}

	if err != nil {
		return err
	}
	// TODO: Put into files.
	fmt.Println(objectGroups)
	return nil
}
