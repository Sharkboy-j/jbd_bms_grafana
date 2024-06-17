FROM golang:1.22-alpine3.20 AS builder
WORKDIR /build
COPY  . .
RUN go mod download
RUN go build -o /bleTest main.go

FROM alpine:3.20
WORKDIR /

COPY --from=builder /bleTest /
RUN ls -la

CMD ["/bleTest"]