FROM golang:1.11


COPY . /go/src/connection-log
WORKDIR /go/src/connection-log

ENV GO111MODULE=on

RUN go build

EXPOSE 8080

CMD ./connection-log