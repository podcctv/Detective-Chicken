FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/jijian-server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/jijian-agent ./cmd/agent

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/jijian-server /usr/local/bin/jijian-server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/jijian-server"]
