FROM alpine:latest

RUN mkdir /app

COPY postApp /app

CMD [ "/app/postApp"]