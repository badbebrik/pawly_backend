# Pawly Backend

Задеплоен на `https://pawly-app.ru`.

Локально API поднимается на `http://localhost:8000`.

## Локальный запуск

Из корня `pawly_backend`:

```bash
docker compose -f deploy/docker-compose.dev.yaml --profile tools run --rm migrate all up
docker compose -f deploy/docker-compose.dev.yaml up -d --build
```

Остановить:

```bash
docker compose -f deploy/docker-compose.dev.yaml down
```
