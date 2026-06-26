module github.com/zhong/pritunl-cli

go 1.26.4

require (
	github.com/zhong/pritunl-go-sdk v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/zhong/pritunl-go-sdk => ../pritunl-go-sdk
