FROM alpine:3.20
WORKDIR /

COPY  . .
RUN ls -la

CMD ["/bleTest"]