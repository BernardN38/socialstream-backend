FROM alpine:latest

RUN mkdir /app

COPY mediaApp /app

CMD [ "/app/mediaApp"]