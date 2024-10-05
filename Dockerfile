FROM postgres:latest

ENV POSTGRES_USER=admin
ENV POSTGRES_PASSWORD=adminpass
ENV POSTGRES_DB=tg

EXPOSE 5432

# Команда запускается автоматически при старте контейнера и инициализирует PostgreSQL
