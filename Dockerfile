FROM golang:1.22.4-alpine3.20 AS builder
WORKDIR /build
COPY  . .
RUN go mod download
RUN go build -o /bleTest .

FROM alpine:3.20
WORKDIR /
COPY --from=builder /bleTest /

CMD ["/bleTest"]