FROM ubuntu:latest

RUN mkdir /app

COPY ./build/app /app

CMD /app/app --port 4001 --db-dsn ${POSTGRES_DSN} --redis ${REDIS_URL}