module github.com/sudachen/spacemesh-playground

go 1.13

replace github.com/spacemeshos/go-spacemesh => ./go-spacemesh

replace github.com/spacemeshos/poet => ./poet

require (
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/grpc-ecosystem/grpc-gateway v1.9.0
	github.com/spacemeshos/amcl v0.0.2
	github.com/spacemeshos/go-spacemesh v0.1.12
	github.com/spacemeshos/poet v0.1.0
	github.com/spacemeshos/post v0.0.0-20191225190235-dfb8a5803e6d
	github.com/spacemeshos/smutil v0.0.0-20190604133034-b5189449f5c5
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.4.0
	github.com/sudachen/smwlt v0.0.0-20200621022533-ad9c7e263ec2
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7
	google.golang.org/grpc v1.23.1
)
