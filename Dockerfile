FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache modules first for faster rebuilds.
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/site ./cmd/site
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/runner ./cmd/runner

FROM gcr.io/distroless/static-debian12:nonroot AS site

WORKDIR /app

COPY --from=builder /out/site /app/site
COPY --from=builder /src/cmd/site/content /app/cmd/site/content
COPY --from=builder /src/internal/site/templates /app/internal/site/templates
COPY --from=builder /src/internal/site/public /app/internal/site/public

EXPOSE 8080

CMD ["/app/site"]

FROM golang:1.24-alpine AS runner

WORKDIR /app

COPY --from=builder /out/runner /app/runner

EXPOSE 8080

CMD ["/app/runner"]

FROM golang:1.24-alpine AS runtime

WORKDIR /app

COPY --from=builder /out/site /app/site
COPY --from=builder /out/runner /app/runner
COPY --from=builder /src/cmd/site/content /app/cmd/site/content
COPY --from=builder /src/internal/site/templates /app/internal/site/templates
COPY --from=builder /src/internal/site/public /app/internal/site/public

EXPOSE 8080

CMD ["/app/site"]
