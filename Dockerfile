# Basic image
FROM golang:1.17 as builder

WORKDIR /

COPY sidecar/ /sidecar

# RUN go get -d -v github.com/find-in-docs/documents/pkg/config
# RUN go get -d -v github.com/find-in-docs/documents/pkg/data
# RUN go get -d -v github.com/find-in-docs/sidecar/pkg/client
# RUN go get -d -v github.com/find-in-docs/sidecar/pkg/utils
# RUN go get -d -v github.com/spf13/viper
# RUN go get -d -v github.com/find-in-docs/sidecar/pkg/utils
# RUN go get -d -v github.com/find-in-docs/sidecar/protos/v1/messages
 
COPY documents/ /service

WORKDIR /service

RUN go build -o bin/documents pkg/main/main.go

FROM alpine:latest

WORKDIR /service

COPY --from=builder /service/bin/documents ./

CMD [ "/service/documents" ]
