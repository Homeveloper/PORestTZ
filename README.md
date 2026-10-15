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

## Нагрузочное тестирование (1000 RPS на один кошелёк)

В `cmd/loadtest` — отдельная утилита, которая шлёт HTTP-запросы с заданным
темпом на один и тот же кошелёк и считает коды ответов и задержки

```bash
docker compose up -d --build
go run ./cmd/loadtest -rps=1000 -duration=10s
