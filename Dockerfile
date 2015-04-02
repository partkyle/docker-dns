FROM golang:1.4.2

ENV PROJECT=github.com/partkyle/docker-dns

COPY . /go/src/$PROJECT

RUN go install $PROJECT

CMD docker-dns
