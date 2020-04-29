# Use the official golang image as parent image.
FROM golang

# Set the working directory
WORKDIR /go/src/

RUN mkdir -p github.com/SasukeBo/ftpviewer

WORKDIR /go/src/github.com/SasukeBo/ftpviewer

COPY . .

RUN go build

RUN apt-get update

RUN apt-get install lsof

CMD ["nohup", "./ftpviewer", "&"]
