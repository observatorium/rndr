GOBIN ?= $(firstword $(subst :, ,${GOPATH}))/bin
RNDR ?= $(GOBIN)/rndr

from-jsonnet-gen:
	@mkdir -p tmpl/jsonnet/.gen/kubernetes/
	@$(RNDR) output --spec="tmpl/jsonnet/hellosvc.rndr.yaml" \
 		 --values-file="1-dont-know-what-to-put-hellosvc.values.yaml" \
 		 -o "tmpl/jsonnet/.gen/kubernetes"
	@mkdir -p tmpl/jsonnet/.gen/kubernetes-special/
	@$(RNDR) output --spec="tmpl/jsonnet/hellosvc.rndr.yaml" \
 		 --values-file="2-my-special-hellosvc.values.yaml" \
 		 -o "tmpl/jsonnet/.gen/kubernetes-special"
	@$(RNDR) package --spec="tmpl/jsonnet/hellosvc.rndr.yaml" helm

assert-equal-output:
	@git --no-pager diff --no-index "expected/" "tmpl/jsonnet/.gen/"

test:
	@$(MAKE) from-jsonnet-gen
	@$(MAKE) assert-equal-output
	@echo "Check Passed"
