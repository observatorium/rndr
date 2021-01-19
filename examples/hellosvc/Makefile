GOBIN ?= $(firstword $(subst :, ,${GOPATH}))/bin
RNDR =? $(GOBIN)/rndr

kubernetes:
	@$(RNDR) --template=hellosvc-tmpl-jsonnet/hellosvc.tmpl.yaml kubernetes --values=dont-know-what-to-put-hellosvc.values.yaml -o hellosvc-tmpl-jsonnet/.gen/kubernetes/

kubernetes-special:
	@$(RNDR) --template=hellosvc-tmpl-jsonnet/hellosvc.tmpl.yaml kubernetes --values=my-special-hellosvc.values.yaml -o hellosvc-tmpl-jsonnet/.gen/kubernetes-special/