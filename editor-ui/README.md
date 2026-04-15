# editor-ui

Frontend редактора Controlitix (React + TypeScript + Vite) по FSD.

## Быстрый старт

- Требования: Node.js >= 18 (Vite 6 не работает на Node 16)
- Установка: `npm ci`
- Dev: `npm run dev`
- Build: `npm run build` (выполняет `tsc -b` и `vite build`)

## Архитектура (FSD)

- `src/app` — провайдеры и инфраструктура (router/theme и т.п.)
- `src/shared` — переиспользуемые библиотеки и UI-атомы
- `src/entities` — состояние/модели домена + работа с API
- `src/features` — сценарии/фичи (UI + поведение)
- `src/widgets` — композиция фич/сборка экранов

Подробнее: `docs/02_Архитектура/Микросервисы/editor-ui.md`.

## Shared UI

UI строится через собственные компоненты-обёртки Mantine, чтобы можно было быстро заменить библиотеку:

- Public API: `src/shared/ui/index.ts`
- Компоненты: `src/shared/ui/components/*`
- Лейаут: `src/shared/ui/Layout/*`

Правило импортов: импортируем только через `index.ts` (см. `agents.md`), например:

- `import { Button, Switch } from "@shared/ui";`
- `import { Layout, Header, Navbar } from "@shared/ui";`

## Theme (3 темы)

Тема вынесена в `src/shared/libs/theme/*`:

- Темы: `light`, `dark`, `ultraDark`
- Получение темы: `getTheme(name)` / `themes[name]`
- `ThemeProvider` хранит выбор в `localStorage` по ключу `controlitix-theme`
- По умолчанию (если ключа нет) берётся системная настройка через `prefers-color-scheme`

Хук для UI:

- `useAppTheme()` (контекст): `themeName`, `setThemeName(name)`, `toggleTheme()`

## Layout (slot-паттерн)

Каркас страницы строится на Mantine `AppShell`, но через слот-компоненты:

- `Layout` принимает слоты: `header`, `navbar`, `panel`, `aside`, `footer` и `children` как `main`
- Слоты типизированы через `RequireAtLeastOne`: хотя бы один слот обязателен
- Есть управление: `navbarCollapsed`, `panelHidden`, `asideHidden` + `useLayout().toggleNavbar()`

## Процесс разработки

Правила работы и линтинга: `agents.md`.
