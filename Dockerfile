FROM golang:1-stretch

RUN apt install -y make

RUN useradd --create-home --shell /bin/bash breath
USER breath

RUN mkdir -p /home/breath/{.cache/go-build,go/bin}
ENV GOPATH=/home/breath/go

COPY . /home/breath/app
WORKDIR /home/breath/app

RUN make build

ENTRYPOINT make run
