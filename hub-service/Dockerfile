FROM alpine:latest

RUN mkdir /app

COPY ./build/hubApp /app

CMD /app/hubApp --port 4000 --db-dsn ${POSTGRES_DSN} --redis ${REDIS_URL}
