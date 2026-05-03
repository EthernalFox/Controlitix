# Controlitix - руководство для frontend-агентов

## Цель документа

- Синхронизировать агентов по фронтенд-проектам Controlitix и зафиксировать ключевые ожидания MVP.
- Дать быстрые ссылки на доменные артефакты и API, чтобы ускорить работу над задачами editor-ui и viewer-ui.

## Проекты и области ответственности

### editor-ui (редактор)

- Назначение: редактирование объектов, мнемосхем и фигур, привязка фигур к тегам, ручная публикация состояния схемы; CRUD устройств, теги управляются через устройство (отдельного раздела нет) (см. `02_Архитектура/Микросервисы/editor-ui.md`).
- Стек: React 19, TypeScript 5, Zustand (история и состояние), KonvaJS (canvas), Mantine UI 8 (панели и формы).
- Методология разработки FSD. Хранилища Zustand и работа с API располагаются в `entities`; атомарные UI-компоненты держим в `shared/ui`; сценарное поведение и отображение описываем во `features`; `widgets` собирают фичи в композиции, `app` отвечает только за провайдеры и инфраструктуру. Layout используем единый и собираем его на уровне `pages`; для экранов редактора добавляем отдельные виджеты (например `DrawHeader`, `DrawFooter`, панели) на базе `shared/ui`. Для каждой фичи или виджета создаём папку `ComponentName` с `ComponentName.tsx`, бизнес-логику выносим в `useModels.tsx`, типы - в `types.ts`, стили - в `styles.scss`.
- Интеграции: REST к `ms-editor` для объектов, устройств, тегов, схем и фигур (`04_API/ms-editor.md`).
- Основные функции MVP: управление списком объектов и схем, редактор холста с трансформациями и правилами цвета или видимости, привязка фигур к тегам, CRUD устройств (Modbus TCP/RTU, SNMP) и тегов через устройства, undo/redo >= 50 шагов, автосохранение черновиков в браузере.
- UX-потоки: создание схемы и публикация, привязка фигур к тегу, создание устройства и тегов описаны в `05_Сценарии-использования/Инженер_*.md`.
- Права: базовый режим - редактирование доступно всем залогиненным; режим только чтение возможен флагом (planned), RBAC вне MVP (planned).
- НФ-требования: 60 fps при >= 500 фигурах, быстрые undo/redo (<= 10 мс), автосохранение каждые ~5 секунд, стабильность форм с оптимистичным сохранением, базовая клавиатурная доступность.

### Линтинг

- ESLint: @eslint/js (recommended) + typescript-eslint (recommended), плагины eslint-plugin-import, eslint-plugin-react, eslint-plugin-prettier.
- Покрываются файлы `**/*.{js,mjs,cjs,ts,jsx,tsx}`, глобалы браузера подключаются через globals.browser.
- Ключевые правила: `react/react-in-jsx-scope` отключено; `comma-dangle`: ["error", "never"]; `import/order`: группы builtin/external, internal, parent/sibling/index, алиасы `@{app,entities,features,shared,pages,widgets}/**` относятся к internal, алфавитная сортировка, пустые строки между блоками.
- Резолвер импортов понимает алиасы @app, @entities, @features, @shared, @pages, @widgets и расширения .js, .jsx, .ts, .tsx.
- Prettier: printWidth 80, tabWidth 2, useTabs false, singleQuote false, quoteProps consistent, semi true, trailingComma none, bracketSpacing true, arrowParens always, endOfLine lf, jsxSingleQuote false, jsxBracketSameLine false.

## Токен-система и тема

### Архитектура (после спек 0047–0050)

Дизайн-токены определяются **один раз** в `cssVariablesResolver` внутри `editor-ui/src/shared/libs/theme/theme.ts` и автоматически инжектируются Mantine в DOM при смене цветовой схемы.

```
theme.ts
  └── cssVariablesResolver → Mantine инжектирует в :root и [data-mantine-color-scheme="light/dark"]
        ├── variables {}   → theme-agnostic токены (alarm-цвета, размеры лейаута)
        ├── light {}       → светлая тема (bg, text, border, glass, accent)
        └── dark {}        → тёмная тема
```

### Пространства имён

| Namespace | Использование |
|---|---|
| `var(--mantine-*)` | Стандартные Mantine-свойства: spacing, radius, shadow, font-family, color shades |
| `var(--ctrx-*)` | Семантические токены проекта: alarm, glass, canvas, surface, accent |

**Запрещено:**
- Хардкодить hex-цвета и `rgba()` в `.module.css` файлах.
- Использовать `var(--bg-*)`, `var(--text-*)`, `var(--border-*)` (старый неймспейс, удалён в spec 0047).
- Импортировать из `@mantine/core` напрямую в виджеты, фичи и страницы — только через `@shared/ui`.

### ThemeProvider

Единственный `MantineProvider` в `app/providers/ThemeProvider/ThemeProvider.tsx`:
- `theme` — из `shared/libs/theme`
- `cssVariablesResolver` — из `shared/libs/theme`
- `colorSchemeManager = localStorageColorSchemeManager({ key: "controlitix-theme" })`
- `defaultColorScheme="light"`

### Как добавить новый токен

1. Добавить переменную в нужный блок (`variables`/`light`/`dark`) в `cssVariablesResolver` в `theme.ts`.
2. Именовать в формате `--ctrx-<категория>-<имя>`.
3. Задокументировать в `docs/06_Дизайн/Tokens.md`.

### Паттерн использования в CSS-модулях

```css
/* spacing и радиус — Mantine vars */
.panel { padding: var(--mantine-spacing-md); border-radius: var(--mantine-radius-sm); }

/* семантика проекта — ctrx vars */
.alarmOk  { color: var(--ctrx-alarm-ok); }
.glass    { background: var(--ctrx-glass-bg); backdrop-filter: var(--ctrx-glass-blur); }
.focused  { outline: 2px solid var(--ctrx-border-focus); }
```

### Полный справочник токенов

См. `docs/06_Дизайн/Tokens.md`.

## Правила работы

- Коммиты: 1 логическое изменение = 1 коммит (используем частичное добавление `git add -p`, избегаем смешивания рефакторинга/фич/форматирования в одном коммите).
- Репозитории: в каждом проекте свой `.git` (например, `editor-ui/.git`, `docs/.git`), коммиты и история ведутся отдельно по каталогу.
- Импорты: используем public API модулей через `index.ts` (например `@shared/ui`, `@shared/ui/components`, `@shared/ui/Layout`, `@shared/libs/theme`), не импортируем из внутренних файлов вида `.../Component/Component.tsx`.

## Архитектурный контекст

- Потоки данных между сервисами описаны в `02_Архитектура/Потоки-данных.md`; фронтенды зависят от событий Kafka (`tags.values`, `config.changed`, `alarms.events`) и кэшей Redis.
- `ms-poll` нормализует значения тегов и публикует их в Kafka или Redis; `ms-viewer` кэширует последние значения, ведёт историю и обслуживает viewer-ui.
- Публикация схем в editor-ui инициирует событие `config.changed`, после чего viewer-ui должен инвалидацировать кеш и подтянуть актуальные данные.
- Реалтайм подписки viewer-ui проходят через WebSocket-шлюз (или REST-пуллинг как fallback, planned); авторизация и сегментация по объектам требует интеграции с `ms-auth` (planned).

## Доменная модель (MVP)

- Основные сущности описаны в `07_Бизнес-сущности/Сущности.md`; глоссарий терминов - `07_Бизнес-сущности/Глоссарий.md`.
- Схема `public`: `objects`, `mimic`, `figures`, `figure_params` (`03_Модель-данных/public/*.md`) - хранит объекты мониторинга, мнемосхемы и фигуры, включая JSONB-параметры.
- Схема `devices`: `devices`, `device_type`, `devices_params` (`03_Модель-данных/devices/*.md`) - справочник и конфиги устройств.
- Схема `tags`: `tags`, `tag_params`, `units`, `data_types`, `tag_setpoints`, `tag_scaling` (`03_Модель-данных/tags/*.md`) - параметры тегов, масштабирование и уставки для динамики.
- Backend-сервисы должны эмитить `config.changed` при изменении конфигураций, чтобы фронтенды обновляли локальное состояние.

## API и интеграции

- Editor UI опирается на REST-контракты `ms-editor` (`04_API/ms-editor.md`) для CRUD объектов, диаграмм и фигур.
- `ms-editor` также обслуживает устройства и теги, включая справочники `device_type`, `data_types`, `units` (`04_API/ms-editor.md`).
- Viewer UI получает тренды, архивы и тревоги через `ms-viewer` (`02_Архитектура/Микросервисы/ms-viewer.md`).
- Реалтайм: `ms-viewer` перенаправляет Kafka или Redis (`02_Архитектура/Потоки-данных.md`, planned); квитирование тревог выполняется через его API (TBA).

## Пользовательские сценарии

- Инженер: добавить устройство (`05_Сценарии-использования/Инженер_добавить_устройство.md`).
- Инженер: создать мнемосхему (`05_Сценарии-использования/Инженер_создать_мнемосхему.md`).
- Инженер: привязать фигуры к тегам (`05_Сценарии-использования/Инженер_привязать_фигуры_к_тегам.md`).
- Инженер: настроить тренды и отчеты (`05_Сценарии-использования/Инженер_настроить_тренды_и_отчеты.md`).
- Оператор: мониторинг и квитирование тревог (`05_Сценарии-использования/Оператор_мониторинг_и_квитирование_тревог.md`).
- Уведомления Telegram - см. `05_Сценарии-использования/Уведомления_Telegram.md` (planned).

## Нефункциональные и UX ориентиры

- Производительность canvas (editor-ui): 60 fps при 500 фигурах, оптимизация через батч-обновления и requestAnimationFrame.
- Undo или redo должно обслуживать >= 50 шагов локально; хранение истории в Zustand.
- Автосохранение черновика каждые ~5 секунд (localStorage или IndexedDB) с визуальной индикацией состояния.
- Viewer-ui должен поддерживать деградацию до polling и сглаживание дребезга обновлений (planned).
- Базовая клавиатурная доступность: навигация стрелками и модификаторами, фокус-индикация (см. `02_Архитектура/Микросервисы/editor-ui.md`).

## Логи и телеметрия

- Минимальный набор событий: `diagram.saved`, `diagram.published`, `device.created`, `tag.created` (см. `02_Архитектура/Микросервисы/editor-ui.md`).
- Ошибки UI или сети логируются с HTTP-кодами и текстами; единую точку сбора телеметрии требуется согласовать отдельно (planned).
- Для тревог - подписка на `alarms.events`, аудит квитирований и уведомления Telegram через `ms-viewer` (детализация в будущих API-спецификациях, planned).

## Открытые вопросы и TODO

- Детализировать линтинговые правила для editor-ui и viewer-ui (planned).
- Описать точные схемы DTO и коды ошибок для `ms-editor` и realtime шлюза (`04_API/*.md` пока наброски, planned).
- Закрепить стратегию версионности схем и черновиков (planned).
- Подготовить рекомендации по интеграции с `ms-auth` и ролям (RBAC вне MVP, planned).
