# Note: Do not place `ARG` above `FROM`.
#       It will not be able to be referenced by RUN.

# Build Geth in a stock Go builder container
FROM golang:1.21.3-bullseye as builder

# Support setting various labels on the final image
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

# automatically set by buildkit, can be changed with --platform flag
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN apt update && apt install -y git

# Get dependencies - will also be cached if we won't change go.mod/go.sum
COPY go.mod /go-ethereum/
COPY go.sum /go-ethereum/
RUN cd /go-ethereum && go mod download

ADD . /go-ethereum
RUN cd /go-ethereum && \
      GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM="$(echo $TARGETVARIANT | cut -c2-)" \
      go run build/ci.go install -static -arch $TARGETARCH ./cmd/geth

# Pull Geth into a second stage deploy debian container
FROM debian:11.9-slim

RUN apt update && \
    apt install -y ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /go-ethereum/build/bin/geth /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp
ENTRYPOINT ["geth"]

# Add some metadata labels to help programmatic image consumption
ARG COMMIT=""
ARG VERSION=""
ARG BUILDNUM=""

LABEL commit="$COMMIT" version="$VERSION" buildnum="$BUILDNUM"
