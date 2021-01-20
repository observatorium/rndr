package rndr

import (
	"testing"

	"github.com/efficientgo/tools/core/pkg/testutil"
	"github.com/observatorium/rndr/pkg/rndr/golang"
	"github.com/observatorium/rndr/pkg/rndr/jsonnet"
)

func TestParseTemplate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, err := ParseTemplate([]byte{}, "")
		testutil.NotOk(t, err)
	})
	t.Run("valid", func(t *testing.T) {
		tmpl, err := ParseTemplate([]byte(`name: "helloservice"
authors: "team@example.com"

# api defines the definition of values.
api:
  go:
    default: "github.com/observatorium/rndr/examples/hellosvc/api.Default()"
    struct: "github.com/observatorium/rndr/examples/hellosvc/api.HelloService"

# renderer defines the rendering engine.
renderer:
  jsonnet:
    functions: 
    - hellosvc.libsonnet
    - second
`), "")
		testutil.Ok(t, err)
		testutil.Equals(t, TemplateDefinition{
			Name:    "helloservice",
			Authors: "team@example.com",

			API: TemplateAPI{Go: &golang.TemplateAPI{
				Default: "github.com/observatorium/rndr/examples/hellosvc/api.Default()",
				Struct:  "github.com/observatorium/rndr/examples/hellosvc/api.HelloService",
			}},
			Renderer: TemplateRenderer{
				Jsonnet: &jsonnet.TemplateRenderer{Functions: []string{"hellosvc.libsonnet", "second"}},
			},
		}, tmpl)
	})
	t.Run("unparsable", func(t *testing.T) {
		_, err := ParseTemplate([]byte(`f: "helloservice"
`), "")
		testutil.NotOk(t, err)
	})
	t.Run("parsable but too many entries for one-ofs", func(t *testing.T) {
		_, err := ParseTemplate([]byte(`name: "helloservice"
authors: "team@example.com"

# api defines the definition of values.
api:
  go:
    default: "github.com/observatorium/rndr/examples/hellosvc/api.Default()"
    struct: "github.com/observatorium/rndr/examples/hellosvc/api.HelloService"
  proto:
    entry: "Config"
    file: "openproto/protoconfig.proto"
  

# renderer defines the rendering engine.
renderer:
  jsonnet:
    file: hellosvc.libsonnet
  helm:
    chart: prometheus
    repo: 
  process:
    command: "./my-cmd"
    inputEnvVar: "INPUT"
    arguments:
    - "--config=${INPUT}
`), "")
		testutil.NotOk(t, err)
	})
}
