FROM golang:1.23.3-alpine3.20 AS builder
WORKDIR /build
COPY . /build
ENV CGO_ENABLED=0  \
  GOCACHE=/.cache/go-build

RUN --mount=type=cache,target=/.cache/go-build go build -o ./service -buildvcs=false -trimpath -ldflags "-s -w" cmd/service/main.go
RUN chmod +x /build/service


#####################
FROM scratch
COPY --from=builder /build/service /usr/bin/service
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/usr/bin/service"]

