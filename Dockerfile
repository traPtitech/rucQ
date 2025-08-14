#syntax=docker/dockerfile:1
FROM golang:1.24.5-bookworm AS builder

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

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /usr/bin/rucq /usr/bin/rucq

CMD [ "/usr/bin/rucq" ]
