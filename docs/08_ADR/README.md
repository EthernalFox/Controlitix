# 8. Architecture Decision Records (ADR)

Архитектурные решения Controlitix фиксируются в формате ADR (Architecture Decision Records). Каждая запись описывает контекст, варианты, принятое решение и его последствия.

## Формат

Каждый ADR — отдельный `.md`-файл с именем `ADR-NNNN-kebab-case-title.md` и следующей структурой:

```
# ADR-NNNN: <Заголовок>

- Статус: proposed | accepted | deprecated | superseded by ADR-XXXX
- Дата: YYYY-MM-DD
- Авторы: <список>

## Контекст
Описание проблемы, ограничений и требований.

## Рассмотренные варианты
Список вариантов с краткой оценкой.

## Решение
Что именно решено и почему.

## Последствия
Позитивные и негативные эффекты принятого решения, что становится сложнее/проще.

## Связанные документы
Ссылки на другие ADR, разделы документации, внешние материалы.
```

## Список решений

| ADR | Статус | Заголовок |
|---|---|---|
| [ADR-0001](ADR-0001-monorepo.md) | accepted | Монорепозиторий для всех сервисов и фронтендов |
| [ADR-0002](ADR-0002-go-backend.md) | accepted | Go как язык backend-сервисов |
| [ADR-0003](ADR-0003-postgres-schemas.md) | accepted | Разделение PostgreSQL на схемы и разрешение кросс-схемных FK |
| [ADR-0004](ADR-0004-timescaledb-history.md) | accepted | TimescaleDB для архива значений тегов |
| [ADR-0005](ADR-0005-docker-compose-mvp.md) | accepted | Docker Compose как способ развёртывания MVP |
| [ADR-0006](ADR-0006-nginx-gateway.md) | accepted | Nginx как единая точка входа (gateway) |
| [ADR-0007](ADR-0007-service-accounts-env-secrets.md) | accepted | ТУЗ и секреты в env для межсервисной аутентификации |
| [ADR-0008](ADR-0008-no-api-versioning-mvp.md) | accepted | Отсутствие версионирования HTTP API в MVP |
| [ADR-0009](ADR-0009-telegram-only-notifications.md) | accepted | Только Telegram как канал уведомлений в MVP |
| [ADR-0010](ADR-0010-kafka-event-bus.md) | accepted | Kafka как шина событий |
| [ADR-0011](ADR-0011-konva-pixi-split.md) | accepted | Konva в редакторе, Pixi в просмотрщике |
| [ADR-0012](ADR-0012-ms-auth-pluggable-identity.md) | accepted | ms-auth — JWT + расширяемые источники идентичности |

## Принципы принятия решений

- **MVP-ориентированность:** MVP существует под конкретного пилотного заказчика (self-hosted, один хост на объект, ~250 устройств). Решения оцениваются против этого контекста, а не абстрактного масштаба.
- **Обратимость:** предпочитаем обратимые решения; сложные необратимые (multi-tenancy, k8s, API-версионирование) откладываем.
- **Документированность:** каждое значимое архитектурное решение — отдельный ADR.
- **Не изобретать лишнего:** если существующий инструмент (TimescaleDB, Nginx, Kafka) закрывает задачу — используем его, не пишем собственное.
