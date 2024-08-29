FROM golang:1.23 AS builder
RUN apt-get update && \
    apt-get install -y gcc-aarch64-linux-gnu gcc-x86-64-linux-gnu && \
    ln -s /usr/bin/aarch64-linux-gnu-gcc /usr/bin/arm64-linux-gnu-gcc  && \
    ln -s /usr/bin/x86_64-linux-gnu-gcc /usr/bin/amd64-linux-gnu-gcc
# 1. Precompile the entire go standard library into the first Docker cache layer: useful for other projects too!
RUN CGO_ENABLED=1 GOOS=linux go install -v -installsuffix cgo -a std
ARG TARGETARCH
ARG APP_SHA
ARG VERSION
WORKDIR /go/src/github.com/to6ka/ledger
# get deps first so it's cached
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH \
    CC=$TARGETARCH-linux-gnu-gcc \
    go build -o numary -tags json1,netgo \
    -ldflags="-X github.com/to6ka/ledger/cmd.Version=${VERSION} \
    -X github.com/to6ka/ledger/cmd.BuildDate=$(date +%s) \
    -X github.com/to6ka/ledger/cmd.Commit=${APP_SHA} \
    -X github.com/to6ka/ledger/cmd.DefaultSegmentWriteKey=none" ./

FROM ubuntu:jammy as app
RUN apt update && apt install -y ca-certificates wget && rm -rf /var/lib/apt/lists/*
COPY --from=builder /go/src/github.com/to6ka/ledger/numary /usr/local/bin/numary
EXPOSE 3068
ENTRYPOINT ["numary"]
ENV OTEL_SERVICE_NAME ledger
CMD ["server", "start"]
