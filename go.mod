module go-webrtc

go 1.13

replace go-lib => ../go-lib

require (
	github.com/gin-gonic/gin v1.6.2
	github.com/go-acme/lego/v3 v3.3.0
	github.com/golang/protobuf v1.3.3
	github.com/pion/rtcp v1.2.1
	github.com/pion/webrtc/v2 v2.2.4
	github.com/prometheus/common v0.9.1 // indirect
	go-lib v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.26.0
)
