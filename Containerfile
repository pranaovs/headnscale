# build
FROM golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o headnscale .

# run
FROM gcr.io/distroless/static-debian13

WORKDIR /app

COPY --from=builder /build/headnscale /app/headnscale

ENTRYPOINT ["/app/headnscale"]
