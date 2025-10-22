module github.com/uwu-octane/antBackend/tests

go 1.25.2

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/uwu-octane/antBackend/api v0.0.0
	google.golang.org/grpc v1.71.0
)

require (
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace (
	github.com/uwu-octane/antBackend/api => ../api
	github.com/uwu-octane/antBackend/auth => ../auth
	github.com/uwu-octane/antBackend/user => ../user
)
