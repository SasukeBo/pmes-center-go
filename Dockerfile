# Use the official golang image as parent image.
FROM golang

# Set the working directory
WORKDIR /go/src/

RUN mkdir -p github.com/SasukeBo/ftpviewer

WORKDIR /go/src/github.com/SasukeBo/ftpviewer

COPY . .

EXPOSE 44761

CMD [ "go", "run", "main.go"]
