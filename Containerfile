# build
FROM golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o headnscale .

# run
FROM gcr.io/distroless/static-debian13@sha256:d5f030ca7c5793784e9ea4178a116da360250411d13921a5af27c6cb5a5949bf

WORKDIR /app

COPY --from=builder /build/headnscale /app/headnscale

ENTRYPOINT ["/app/headnscale"]
