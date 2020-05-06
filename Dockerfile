# Use the official golang image as parent image.
FROM golang

COPY ./deploy.sh /deploy.sh

RUN chmod -R 755 /deploy.sh

RUN apt-get update

RUN apt-get install lsof

CMD ["/deploy.sh"]
