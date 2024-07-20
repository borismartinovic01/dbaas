FROM alpine:latest

RUN mkdir /app

COPY monitoringApp /app

CMD ["/app/monitoringApp"]