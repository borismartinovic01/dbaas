FROM alpine:latest

RUN mkdir /app

COPY fileConfigApp /app

CMD ["/app/fileConfigApp"]