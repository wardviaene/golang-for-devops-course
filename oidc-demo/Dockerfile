#
# Build go project
#
FROM golang:1.18-alpine as go-builder

WORKDIR /oidc-demo

COPY . .

RUN apk add -u -t build-tools curl git && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/*.go && \
    apk del build-tools && \
    rm -rf /var/cache/apk/*

#
# Runtime container
#
FROM alpine:latest  

WORKDIR /app

RUN apk --no-cache add ca-certificates bash curl

COPY --from=go-builder /oidc-demo/server /app/oidc-demo-server

COPY config.yaml /app/config.yaml

ENTRYPOINT ["/app/oidc-demo-server"]
