package jsonnet

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/brancz/locutus/render/jsonnet"
	"github.com/brancz/locutus/rollout"
	"github.com/efficientgo/tools/core/pkg/errcapture"
	"github.com/efficientgo/tools/core/pkg/logerrcapture"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

type TemplateRenderer struct {
	// Functions represent a local or absolute paths to .jsonnet files with
	// single `function(values) {` that renders manifests in right order.
	// Each function's manifests will be part of different groups allowing parallel rollout if requested.
	Functions []string
}

var applyAllLocutusJsonnetTmpl = template.Must(template.New("").Parse(`
local values = import '{{ .LocutusVirtualConfigPath }}';

local groups = [
{{- range .FunctionFiles }}
local manifests = (import '{{ . }}')(values);
{{-  end }}
];


{
  name: '{{ . }}',
  steps: [
	{
	  action: 'CreateOrUpdate',
	  object: item,
	}
	for item in std.objectFields(manifests)
  ],
},

];



{
  objects: {
		
	for item in std.objectFields(groups)
  },
  rollout: {
    apiVersion: 'workflow.kubernetes.io/v1alpha1',
    kind: 'Rollout',
    metadata: {
	  template: '{{ .Name }}',
      name: 'rndr-generated-jsonnet',
    },
    spec: {
      groups: groups,
    },
  },
}`))

// locutusify takes function file path and creates Locutus boilerplate in specified location
// that takes locutus specific jsonnet file that imports and executes function file with values taken from
// locutus specific path specified by jsonnet.VirtualConfigPath.
// NOTE(bwplotka): We are reinventing invocation part for two reason:
// * To simplify input passing and not leak locutus existence. User does not need to know exact "virtual config path" which might be non-intuitive to learn
// about in the first place.
// * To make sure jsonnet templates can be run natively without mocking any input (e.g for testing purposes)
// * To reduce user knowledge about locutus rollout abstractions. `rndr` is meant as a tool which can be used reliably
// for both operator, helm and GitOps flows. If the operator flow is not necessary (for example for stateless services) we
// want to make sure no operator will be deployed. This significantly reduces simplifies that stack if it can be simplified.
// TODO(bwplotka): Potentially something to upstream on Locutus side.
func locutusify(entry string, templName string, functionFiles []string) (err error) {
	if err := os.RemoveAll(entry); err != nil {
		return err
	}

	f, err := os.Create(entry)
	if err != nil {
		return err
	}
	defer errcapture.Do(&err, f.Close, "close locutus entry")

	return applyAllLocutusJsonnetTmpl.Execute(io.MultiWriter(f, os.Stdout), struct {
		Name                     string
		LocutusVirtualConfigPath string
		FunctionFiles            []string
	}{
		Name:                     templName,
		FunctionFiles:            functionFiles,
		LocutusVirtualConfigPath: jsonnet.VirtualConfigPath,
	})
}

func Render(logger log.Logger, name string, c TemplateRenderer, valuesJSON []byte) (groups map[string][]string, err error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "rndr")
	if err != nil {
		return nil, err
	}
	defer logerrcapture.Do(logger, func() error { return os.RemoveAll(tmpDir) }, "remove tmp dir")

	entry := filepath.Join(tmpDir, "main.jsonnet")
	if err := locutusify(
		entry,
		name,
		c.Functions,
	); err != nil {
		return nil, errors.Wrap(err, "locutusify")
	}

	level.Debug(logger).Log("msg", "rendered jsonnet intermidiate file", "path", entry)

	res, err := jsonnet.NewRenderer(logger, entry, nil).Render(valuesJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "render jsonnet from options %+v", c)
	}

	if res.Rollout == nil || len(res.Rollout.Spec.Groups) == 0 {
		return nil, errors.Errorf("no rollout resource rendered by locutus entry %v", entry)
	}
	ret := make(map[string][]string, len(res.Rollout.Spec.Groups))

	// We control boilerplate so we expect purely CreateOrUpdate actions.
	for _, g := range res.Rollout.Spec.Groups {
		for _, s := range g.Steps {
			if s.Action != (&rollout.CreateOrUpdateObjectAction{}).Name() {
				return nil, errors.Errorf("rollout step rendered by locutus has unexpected action %v", s)
			}
			ret[g.Name] = append(ret[g.Name], s.Object)
		}
	}
	return ret, nil
}
