FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN make build

FROM alpine:3.12

RUN apk add --no-cache git

COPY --from=builder /app/bin/gover /bin/gover
COPY --from=builder /app/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
