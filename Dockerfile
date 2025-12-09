FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG RELEASE_VERSION

# Install our build tools
RUN apk add --update ca-certificates

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH
ENV LDFLAGS="-X 'main.VERSION=${RELEASE_VERSION}' "

RUN echo 'nobody:*:65534:65534:nobody:/_nonexistent:/bin/false' >> /etc_passwd

COPY . ./

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build $DEBUGFLAGS -o bin/vault-autounseal-operator github.com/camaeel/vault-autounseal-operator/cmd/vault-autounseal-operator

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/vault-autounseal-operator /vault-autounseal-operator
COPY --from=builder /etc_passwd /etc/passwd

USER nobody

ENTRYPOINT ["/vault-autounseal-operator"]
