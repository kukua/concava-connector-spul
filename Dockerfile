FROM golang:1.5.1
MAINTAINER Maurits van Mastrigt <maurits@kukua.cc>

WORKDIR /data
COPY ./ /data/

EXPOSE 3333
EXPOSE 5555
CMD ["go", "run", "src/connector.go"]
