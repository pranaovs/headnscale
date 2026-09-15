# build
FROM golang:1.26-alpine@sha256:ce864e7223ac17b1775e6fd0b4c0db580c2eb50e7953a427916379e4b92a1628 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o headnscale .

# run
FROM gcr.io/distroless/static-debian13@sha256:f2ea2709ac8db56323cbd7d014277f32cb572d9ea124b0076f7aafe5980678fe

WORKDIR /app

COPY --from=builder /build/headnscale /app/headnscale

ENTRYPOINT ["/app/headnscale"]
