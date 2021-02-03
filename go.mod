module github.com/observatorium/rndr

go 1.15

require (
	cuelang.org/go v0.2.2 // indirect
	github.com/brancz/locutus v0.0.0-20210118164634-ff6bf1183da1
	github.com/bwplotka/bingo v0.3.0 // indirect
	github.com/efficientgo/tools/core v0.0.0-20210120193558-db1e3eb63de3
	github.com/go-kit/kit v0.10.0
	github.com/observatorium/rndr/examples/hellosvc/api/go v0.0.0-20210202202431-a8647360f0a9 // indirect
	github.com/oklog/run v1.1.0
	github.com/openproto/protoconfig/go v0.0.0-20210120170055-746d71fbb221
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)

replace github.com/brancz/locutus => ../locutus
