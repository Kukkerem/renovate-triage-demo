# Multi-stage build for the demo provider binary.
FROM golang:1.27.1-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/provider ./cmd/provider

# Runtime base — pinned by both tag and digest.
FROM alpine:3.23.3@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
RUN apk --no-cache add ca-certificates

# Must be kept in lockstep with Makefile's TERRAFORM_PROVIDER_VERSION and the
# generated CRDs/schema; renovate.json's customManager extracts both.
ARG TERRAFORM_PROVIDER_VERSION=1.16.1
ENV TERRAFORM_PROVIDER_VERSION=${TERRAFORM_PROVIDER_VERSION}

ENV USER_ID=65532
COPY --from=build /out/provider /usr/local/bin/renovate-triage-demo-provider
USER ${USER_ID}

ENTRYPOINT ["renovate-triage-demo-provider"]

LABEL org.opencontainers.image.title="renovate-triage-demo" \
      org.opencontainers.image.description="Miniature Crossplane-shaped provider used as a Renovate triage demo fixture" \
      org.opencontainers.image.source="https://github.com/kukkerem/renovate-triage-demo" \
      org.opencontainers.image.base.name="docker.io/library/alpine:3.23.3"
