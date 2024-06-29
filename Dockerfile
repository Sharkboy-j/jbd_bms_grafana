FROM alpine:3.20
WORKDIR /
COPY bleTest /
RUN chmod +x bleTest
CMD ["/bleTest"]