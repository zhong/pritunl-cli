module github.com/example/pritunl-cli

go 1.26.4

require (
	github.com/example/pritunl-go-sdk v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/example/pritunl-go-sdk => ../pritunl-go-sdk
