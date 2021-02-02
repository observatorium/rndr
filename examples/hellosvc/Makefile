GOBIN ?= $(firstword $(subst :, ,${GOPATH}))/bin
RNDR ?= $(GOBIN)/rndr

from-jsonnet-gen:
	@mkdir -p tmpl/jsonnet/.gen/kubernetes/
	@$(RNDR) --template="tmpl/jsonnet/hellosvc.tmpl.yaml" kubernetes manifests \
 		 --values-file="1-dont-know-what-to-put-hellosvc.values.yaml" \
 		 -o "tmpl/jsonnet/.gen/kubernetes"
	@mkdir -p tmpl/jsonnet/.gen/kubernetes-special/
	@$(RNDR) --template="tmpl/jsonnet/hellosvc.tmpl.yaml" kubernetes manifests \
 		 --values-file="2-my-special-hellosvc.values.yaml" \
 		 -o "tmpl/jsonnet/.gen/kubernetes-special"

assert-equal-output:
	@git --no-pager diff --no-index "expected/" "tmpl/jsonnet/.gen/"

test:
	@$(MAKE) from-jsonnet-gen
	@$(MAKE) assert-equal-output
