FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/detective-chicken-server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/detective-chicken-agent ./cmd/agent

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/detective-chicken-server /usr/local/bin/detective-chicken-server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/detective-chicken-server"]
