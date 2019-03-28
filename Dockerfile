# build stage
FROM golang:alpine AS build-env
RUN apk add make git bash
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"
ADD ./ /go/src/github.com/bmeg/sifter
RUN cd /go/src/github.com/bmeg/sifter && make hub-build

# final stage
FROM alpine
RUN apk add ca-certificates
WORKDIR /data
VOLUME /data
ENV PATH="/app:${PATH}"
COPY --from=build-env /go/bin/sifter /app/
