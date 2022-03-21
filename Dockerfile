#############      builder                                  #############
FROM eu.gcr.io/gardener-project/3rd/golang:1.17.8 AS builder

ENV BINARY_PATH=/go/bin
WORKDIR /go/src/github.com/23technologies/machine-controller-manager-provider-ionos

COPY . .
RUN hack/build.sh

#############      base                                     #############
FROM eu.gcr.io/gardener-project/3rd/alpine:3.15 as base

RUN apk add --update bash curl tzdata
WORKDIR /

#############      machine-controller               #############
FROM base AS machine-controller
LABEL org.opencontainers.image.source="https://github.com/23technologies/machine-controller-manager-provider-ionos"

COPY --from=builder /go/bin/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
