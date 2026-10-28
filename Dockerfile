# Build on the runner's native platform and cross-compile, instead of emulating the target platform.
FROM --platform=$BUILDPLATFORM golang:1.27.3 AS builder

ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/main main.go

FROM alpine:3.24.3
COPY --from=builder /out/main /app/main

# nobody:nobody. A numeric ID lets Kubernetes verify runAsNonRoot.
USER 65534:65534

CMD /app/main $PORT
