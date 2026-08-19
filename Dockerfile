FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/detective-chicken-server ./cmd/server && \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/detective-chicken-agent-amd64 ./cmd/agent && \
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o /out/detective-chicken-agent-arm64 ./cmd/agent && \
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags="-s -w" -o /out/detective-chicken-agent-armv7 ./cmd/agent

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/detective-chicken-server /usr/local/bin/detective-chicken-server
COPY --from=builder /out/detective-chicken-agent-* /usr/local/share/detective-chicken/
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/detective-chicken-server"]
