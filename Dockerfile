FROM --platform=$BUILDPLATFORM golang:1.21-alpine as builder

# Install our build tools
RUN apk add --update ca-certificates

WORKDIR /app

ARG DEBUG
ARG TARGETOS
ARG TARGETARCH
ENV LDFLAGS "-X 'main.VERSION=${RELEASE_VERSION}' "

COPY . ./

RUN if [ DEBUG -eq 1 ]; then export DEBUGFLAGS='-gcflags=all="-N -l"'; fi && \
  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build $DEBUGFLAGS -o bin/vault-autounseal github.com/camaeel/vault-unseal-operator/cmd/vault-unseal-operator

FROM golang:1.21-alpine as debug

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/vault-unseal-operator /vault-unseal-operator
RUN go install github.com/go-delve/delve/cmd/dlv@latest

ENTRYPOINT ["dlv"]

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/vault-unseal-operator /vault-unseal-operator

ENTRYPOINT ["/vault-unseal-operator"]
