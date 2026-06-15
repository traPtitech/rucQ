#syntax=docker/dockerfile:1
FROM golang:1.26.1-trixie@sha256:1d414b0376b53ec94b9a2493229adb81df8b90af014b18619732f1ceaaf7234a AS builder

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod

RUN --mount=type=cache,target=${GOCACHE} \
    --mount=type=cache,target=${GOMODCACHE} \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download

RUN --mount=type=cache,target=${GOCACHE} \
    --mount=type=cache,target=${GOMODCACHE} \
    --mount=type=bind,target=. \
    go build -o /usr/bin/rucq main.go

FROM gcr.io/distroless/static-debian12:nonroot@sha256:d093aa3e30dbadd3efe1310db061a14da60299baff8450a17fe0ccc514a16639

WORKDIR /app

COPY --from=builder /usr/bin/rucq /usr/bin/rucq

CMD [ "/usr/bin/rucq" ]
