# Wallet REST API

Тестовое задание: REST-сервис для операций с кошельками (DEPOSIT/WITHDRAW) на Go + PostgreSQL.

## Стек
- Go 1.25
- PostgreSQL 16
- Docker / Docker Compose
- chi (роутер), pgx (драйвер БД), shopspring/decimal (деньги без потери точности)

## Запуск

```bash
docker compose up -d --build