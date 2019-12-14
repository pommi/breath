FROM golang:1-stretch

RUN apt install -y make

RUN mkdir -p /root/{.cache/go-build,go/bin}
ENV GOPATH=/root/go

COPY . /root/app
WORKDIR /root/app

RUN make build

ENTRYPOINT ["make", "run"]
