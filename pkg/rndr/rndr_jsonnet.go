package rndr

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/brancz/locutus/render/jsonnet"
	"github.com/efficientgo/tools/core/pkg/errcapture"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

var applyAllLocutusJsonnetTmpl = template.Must(template.New("").Parse(`
local values = import '{{ .LocutusVirtualConfigPath }}';
local manifests = (import '{{ .FunctionFile }}')(values);
{
  objects: manifests,
  rollout: {
    apiVersion: 'workflow.kubernetes.io/v1alpha1',
    kind: 'Rollout',
    metadata: {
	  template: '{{ .Name }}',
      name: 'rndr-generated-jsonnet',
    },
    spec: {
      groups: [
        {
          steps: [
            {
              action: 'CreateOrUpdate',
              object: item,
            }
            for item in std.objectFields(manifests)
          ],
        },
      ],
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
func locutusify(entry string, templName string, functionFile string) (err error) {
	if err := os.RemoveAll(entry); err != nil {
		return err
	}

	f, err := os.Create(entry)
	if err != nil {
		return err
	}
	defer errcapture.Do(&err, f.Close, "close locutus entry")

	return applyAllLocutusJsonnetTmpl.Execute(f, struct {
		Name                     string
		LocutusVirtualConfigPath string
		FunctionFile             string
	}{
		Name:                     templName,
		FunctionFile:             functionFile,
		LocutusVirtualConfigPath: jsonnet.VirtualConfigPath,
	})
}

func (r *rndr) renderJsonnet(values []byte) error {
	entry := filepath.Join(r.tmpDir, "main.jsonnet")
	if err := locutusify(
		entry,
		r.t.Name,
		r.t.Renderer.Jsonnet.Function,
	); err != nil {
		return errors.Wrap(err, "locutusify")
	}

	level.Debug(r.logger).Log("msg", "rendered jsonnet intermidiate file", "path", entry)

	res, err := jsonnet.NewRenderer(r.logger, entry, nil).Render(values)
	if err != nil {
		return errors.Wrapf(err, "render jsonnet from options %+v", *r.t.Renderer.Jsonnet)
	}

	fmt.Println(res.Objects, res.Rollout)
	return nil
}
