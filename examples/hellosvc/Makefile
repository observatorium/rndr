GOBIN ?= $(firstword $(subst :, ,${GOPATH}))/bin
RNDR ?= $(GOBIN)/rndr

kubernetes:
	@mkdir -p hellosvc-tmpl-jsonnet/.gen/kubernetes/
	@$(RNDR) --template-file="hellosvc-tmpl-jsonnet/hellosvc.tmpl.yaml" kubernetes manifests \
 		 --values-file="dont-know-what-to-put-hellosvc.values.yaml" \
 		 -o "hellosvc-tmpl-jsonnet/.gen/kubernetes"

kubernetes-special:
	@mkdir -p hellosvc-tmpl-jsonnet/.gen/kubernetes-special/
	@$(RNDR) --template-file="hellosvc-tmpl-jsonnet/hellosvc.tmpl.yaml" kubernetes manifests \
 		 --values-file="my-special-hellosvc.values.yaml" \
 		 -o "hellosvc-tmpl-jsonnet/.gen/kubernetes-special"