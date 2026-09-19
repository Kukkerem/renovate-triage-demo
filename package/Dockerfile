# Multi-stage build for the demo provider binary.
FROM golang:1.26.5-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/provider ./cmd/provider

# Runtime base — pinned by both tag and digest.
FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
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
