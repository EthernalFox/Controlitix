# Отчет по Controlitix (editor-ui и ms-editor)

## Общий обзор

- В корне проекта два отдельных репозитория: `editor-ui/` (frontend) и `ms-editor/` (backend).
- Документация по API и данным живет в `docs/`, и используется как источник контрактов.
- Важно: в архитектуре запрещены кросс‑схемные FK.

## editor-ui

- Архитектура FSD: базовые UI‑компоненты в `shared/ui`, состояние и API клиенты в `entities`, композиция экранов в `pages`.
- Навигация: `AppNavbar` ведет на объекты, мнемосхемы и устройства; хедер без меню (переход на главную через логотип).
- Страница Draw собрана через единый `Layout` в `editor-ui/src/pages/DrawPage/DrawPage.tsx`:
  - Header: `DrawHeader` (кнопки выравнивания/сеток/публикации).
  - Левая панель: `DrawElementsPanel` (элементы/слои).
  - Правая панель: `DrawPropertiesPanel` (свойства выделенного элемента).
  - Footer: `DrawFooter` (палитра фигур).
  - Контент: `DrawCanvas` как placeholder под будущий canvas.
- Список фигур вынесен в `editor-ui/src/entities/shapes/shapes.ts` и используется в футере.
- Базовый слой API в `editor-ui/src/shared/modules/api` (BaseClient/BaseApiClient/BaseFileClient, форматтеры).
- Клиенты ms-editor для объектов и мнемосхем находятся в `editor-ui/src/entities/objects/api` и `editor-ui/src/entities/mimics/api`; пока не подключены к Zustand‑сторам.
- Страницы объектов/мнемосхем используют моки и заготовки, без реальной интеграции с API.

## ms-editor

- Clean Architecture:
  - `cmd/ms-editor` — точка входа.
  - `internal/domain` — сущности и доменные ошибки.
  - `internal/usecase` — сценарии (CRUD объектов/мнемосхем/фигур).
  - `internal/transport/http` — HTTP роутинг и DTO.
  - `internal/infrastructure` — подключение к Postgres и репозиторий.
- HTTP API соответствует `docs/04_API/ms-editor.md` и включает дополнительные списковые эндпоинты для editor-ui:
  - `GET /objects`
  - `GET /objects/{id}/diagrams`
  - `GET /diagrams/{id}/figures`
- Репозитории в `internal/infrastructure/repository` пока заглушки (возвращают `ErrNotImplemented`).
- Конфигурация через `.env` и переменные окружения:
  - `MS_EDITOR_HTTP_ADDRESS`
  - `MS_EDITOR_LOG_LEVEL`
  - `MS_EDITOR_DATABASE_URL`
- Миграции через Goose в `ms-editor/migrations/000001_init.sql`:
  - Таблицы `public.objects`, `public.mimic`, `public.figures`, `public.figure_params`.
  - Внешние ключи только внутри схемы `public`, `tag_id` без FK.
- Docker окружение:
  - `ms-editor/Dockerfile` — сборка сервиса.
  - `ms-editor/docker-compose.yml` — Postgres + сервис.
- Swagger UI доступен по `/swagger`, спецификация — `/swagger/openapi.yaml`.

## Что еще не реализовано

- Реальные запросы к Postgres в репозиториях ms-editor (CRUD логика).
- Подключение editor-ui сторонов (Zustand) к API клиентам и отказ от моков.
- Полная спецификация ошибок и примеры в OpenAPI.

## Как запускать

- ms-editor: см. `ms-editor/README.md` (Docker Compose и миграции Goose).
- editor-ui: запуск остается по проектным скриптам (не обновлялись в рамках отчета).
