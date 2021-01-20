GOBIN ?= $(firstword $(subst :, ,${GOPATH}))/bin
RNDR ?= $(GOBIN)/rndr

kubernetes:
	@mkdir -p hellosvc-tmpl-jsonnet/.gen/kubernetes/
	@$(RNDR) --log.level=debug --template="hellosvc-tmpl-jsonnet/hellosvc.tmpl.yaml" kubernetes manifests \
 		 --values-file="1-dont-know-what-to-put-hellosvc.values.yaml" \
 		 -o "hellosvc-tmpl-jsonnet/.gen/kubernetes"

kubernetes-special:
	@mkdir -p hellosvc-tmpl-jsonnet/.gen/kubernetes-special/
	@$(RNDR) --template="hellosvc-tmpl-jsonnet/hellosvc.tmpl.yaml" kubernetes manifests \
 		 --values-file="2-my-special-hellosvc.values.yaml" \
 		 -o "hellosvc-tmpl-jsonnet/.gen/kubernetes-special"