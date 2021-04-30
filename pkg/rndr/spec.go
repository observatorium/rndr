package rndr

import (
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Spec specifies the renderable definition file.
type Spec struct {
	Name    string
	Authors string

	Template *Template
	TemplateRef *TemplateRef


	// Packages is a map of packages made using provided renderable spec.
	Packages map[string]Package
}

type TemplateRef struct {
	// TODO
}

// ParseSpec parses Spec from bytes.
// TODO(bwplotka): Version it.
// TODO(bwplotka): Validate one-offs.
func ParseSpec(b []byte, dir string) (Spec, error) {
	s := Spec{}
	if err := yaml.Unmarshal(b, &s); err != nil {
		return Spec{}, errors.Wrapf(err, "parsing template content %q", string(b))
	}

	if s.Name == "" {
		return Spec{}, errors.New("name not specified, but required")
	}

	if s.Authors == "" {
		return Spec{}, errors.New("authors not specified, but required")
	}

	if s.Template != nil {
		switch {
		case s.Template.API.Go != nil:
			if s.Template.API.Go.Struct == "" {
				return Spec{}, errors.New("api.go.struct not specified, but required")
			}
		case s.Template.API.Proto != nil:
			if s.Template.API.Proto.Message == "" {
				return Spec{}, errors.New("api.proto.message not specified, but required")
			}
			if s.Template.API.Proto.File == "" {
				return Spec{}, errors.New("api.proto.file not specified, but required")
			}
			s.Template.API.Proto.File = abs(s.Template.API.Proto.File, dir)
		default:
			return Spec{}, errors.New("template api has to be specified, got none")
		}
	} else {
		// TODO: Implement ref.
		return Spec{}, errors.New("template has to be specified, got none")
	}


	// TODO(bwplotka): Add validation for renderers.
	switch {
	case s.Template.Renderer.Jsonnet != nil:
		if len(s.Template.Renderer.Jsonnet.Functions) == 0 {
			return Spec{}, errors.New("jsonnet template renderer has to have at least single function file specified, got none")
		}
		for i := range s.Template.Renderer.Jsonnet.Functions {
			s.Template.Renderer.Jsonnet.Functions[i] = abs(s.Template.Renderer.Jsonnet.Functions[i], dir)
		}

	case s.Template.Renderer.Helm != nil:
	case s.Template.Renderer.Process != nil:
		s.Template.Renderer.Process.Command = abs(s.Template.Renderer.Process.Command, dir)
	default:
		return Spec{}, errors.New("template renderer has to be specified, got none")
	}

	for p, o := range s.Packages {
		switch {
		case o.OLM != nil:
		case o.KubeOperator != nil:
		case o.OpenshiftTemplate != nil:
		case o.Helm != nil:
		default:
			return Spec{}, errors.Errorf("package type has to be specified, got none for %v", p)
		}
	}

	return s, nil
}

func abs(path string, relDir string) string {
	if relDir == "" {
		return path
	}

	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(relDir, path)
}

