# Basic image
FROM golang:1.18 as builder

COPY . /src
WORKDIR /src
RUN make static

FROM alpine:3.11.3
COPY --from=builder /src/gopensearch-linux-amd64 /usr/bin/gopensearch

EXPOSE 9200

ENTRYPOINT [ "/usr/bin/gopensearch" ]