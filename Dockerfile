FROM alpine:3.20
WORKDIR /
COPY bleTest /

CMD ["/bleTest"]