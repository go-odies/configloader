FROM build-base AS builder

COPY . .

RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/api /app/api

WORKDIR /app

CMD ["/app/api"]
