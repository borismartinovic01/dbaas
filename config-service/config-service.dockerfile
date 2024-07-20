FROM alpine:latest

RUN mkdir /app

COPY configApp /app

CMD ["/app/configApp"]