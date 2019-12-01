FROM golang:latest AS builder
RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN go get github.com/google/uuid
RUN go get github.com/jmoiron/sqlx
RUN go get github.com/lib/pq

COPY ./certs /certs

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goflow .

FROM alpine:latest

RUN apk add --no-cache git ca-certificates

COPY --from=builder /certs /certs
COPY --from=builder /app/goflow /goflow

CMD ["/goflow"]