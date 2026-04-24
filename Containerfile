# build
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o headnscale .

# run
FROM gcr.io/distroless/static-debian12@sha256:739834468f307223e74360e20d82996e382d5d74268e6f183701657c645e7518

WORKDIR /app

COPY --from=builder /build/headnscale /app/headnscale

USER nonroot:nonroot

ENTRYPOINT ["/app/headnscale"]
