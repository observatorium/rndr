module github.com/observatorium/rndr

go 1.15

require (
	github.com/brancz/locutus v0.0.0-20210118164634-ff6bf1183da1
	github.com/efficientgo/tools/core v0.0.0-20210112005647-abeaf368c334
	github.com/go-kit/kit v0.10.0
	github.com/oklog/run v1.1.0
	github.com/openproto/protoconfig/go v0.0.0-20210118220306-2c3249881f2c
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

replace (
github.com/brancz/locutus => ../locutus
)