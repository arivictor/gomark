FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache modules first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/site ./cmd/site

FROM gcr.io/distroless/static-debian12:nonroot AS site

WORKDIR /app

COPY --from=builder /out/site /app/bin/site
COPY --from=builder /src/cmd/site/content /app/cmd/site/content
COPY --from=builder /src/templates /app/site/templates
COPY --from=builder /src/public /app/site/public

EXPOSE 8080

CMD ["/app/bin/site"]
