# Design System — Controlitix MVP

## Figma-файлы

| Артефакт | Ссылка |
|---|---|
| **Design System & Wireframes** (основной файл) | [Figma Design](https://www.figma.com/design/PcVBgUkUUs6rIF3DZXo0wQ) |
| Site Map: editor-ui | [FigJam](https://www.figma.com/online-whiteboard/create-diagram/8583b198-a2ef-4d9a-930e-cd8ba7cfe190) |
| Site Map: viewer-ui | [FigJam](https://www.figma.com/online-whiteboard/create-diagram/e89f8e88-638b-4fd6-8cb4-65390d25856d) |
| User Flow: Инженер создаёт мнемосхему | [FigJam](https://www.figma.com/online-whiteboard/create-diagram/69612c12-c3c3-48fe-b33c-65f31fb7c3c2) |
| User Flow: Оператор мониторит тревоги | [FigJam](https://www.figma.com/online-whiteboard/create-diagram/f330fb2f-8a60-47ea-8913-6f3519cb8f6a) |

## Страницы в основном Figma-файле

- **Design System & HMI** — палитра тёмной темы, семантика цветов тревог, правила HMI.
- **editor-ui** — wireframe редактора мнемосхем (toolbar, left panel, canvas, right panel, status bar).
- **viewer-ui** — wireframe просмотра мнемосхемы с панелью тревог, экран трендов, логин.

---

## UI Kit — Mantine 8

Базовая библиотека компонентов — [Mantine UI 8](https://mantine.dev).

### Почему Mantine

- `AppShell` — встроенный лейаут с header / navbar / aside / footer.
- Двойная тема (light / dark) через `MantineProvider` и CSS-переменные — переключение без перезагрузки.
- Компактные размеры (`size="xs"`, `size="sm"`) из коробки — критично для SCADA-плотности.
- Готовые компоненты форм, модалок, drawer, tabs, notifications, select с поиском — покрывают потребности CRUD-экранов и панелей редактора.
- Альтернатива shadcn/ui рассматривалась и отклонена: требует Tailwind, не имеет AppShell, каждый компонент собирается вручную — нецелесообразно для инженерного UI.

### Как используется

Mantine-компоненты **не импортируются напрямую** в виджеты и фичи. Между ними стоит адаптерный слой `shared/ui/components`, который:
- Реэкспортирует Mantine-компоненты (AppShell, Tabs, Select, Modal, TextInput, …).
- Оборачивает компоненты с кастомной API-поверхностью, если дизайн-система Controlitix отличается от дефолтов Mantine (Button с вариантами `primary` / `secondary` / `danger` / `ghost`).

Это позволяет при необходимости заменить Mantine-компонент без изменения потребителей.

### Конфигурация темы

Тема настраивается через `createTheme()` + `MantineProvider`. Ключевые точки:

| Параметр | Значение |
|---|---|
| `primaryColor` | `"primary"` (кастомная палитра из 10 оттенков) |
| `fontFamily` | Inter, sans-serif |
| `fontFamilyMonospace` | JetBrains Mono, monospace |
| `defaultRadius` | `"sm"` (4px) |
| `spacing` | `{ xs: 4, sm: 8, md: 16, lg: 24, xl: 32 }` |
| `colors` | кастомные палитры `primary`, `secondary` + семантические цвета тревог |

Переключатель темы — в меню пользователя (header). Состояние хранится в `localStorage`.

---

## Темы

### Распределение по приложениям

| Приложение | Тема по умолчанию | Обоснование |
|---|---|---|
| **editor-ui** | **Light** | Инженер работает при нормальном офисном освещении |
| **viewer-ui** | **Dark** | Оператор работает в диспетчерской, часто в полутени |

Обе темы обязательны в MVP. Пользователь может переключить тему вручную.

### Light Theme — editor-ui (основная)

| Токен | Hex | Использование |
|---|---|---|
| `bg-primary` | `#FFFFFF` | Основной фон |
| `bg-secondary` | `#F8F9FA` | Панели, sidebar |
| `bg-surface` | `#FFFFFF` | Карточки, модалки |
| `bg-elevated` | `#F1F3F5` | Hover, tooltip |
| `border-default` | `#DEE2E6` | Границы, разделители |
| `text-primary` | `#212529` | Основной текст |
| `text-secondary` | `#495057` | Вторичный текст |
| `text-muted` | `#868E96` | Неактивный текст |
| `accent-blue` | `#228BE6` | Акцент, ссылки, выделение |
| `canvas-bg` | `#F5F5F5` | Фон холста редактора |
| `toolbar-bg` | `#FAFAFA` | Фон footer-toolbar |

### Dark Theme — viewer-ui (основная)

| Токен | Hex | Использование |
|---|---|---|
| `bg-primary` | `#111422` | Основной фон |
| `bg-secondary` | `#1A1D2E` | Панели, sidebar |
| `bg-surface` | `#232740` | Карточки, модалки |
| `bg-elevated` | `#2D3154` | Hover, tooltip |
| `border-default` | `#3A3F5C` | Границы, разделители |
| `text-primary` | `#E8EAF0` | Основной текст |
| `text-secondary` | `#9CA3B8` | Вторичный текст |
| `text-muted` | `#6B7394` | Неактивный текст |
| `accent-blue` | `#4C9AFF` | Акцент, ссылки, выделение |
| `canvas-bg` | `#111422` | Фон холста |
| `toolbar-bg` | `#232740` | Фон footer-toolbar |

### Маппинг токенов на Mantine

Токены реализуются через кастомные `colors` в `createTheme()`:

```
bg-primary    → Mantine body background (theme.other или CSS var)
bg-secondary  → AppShell.Navbar / AppShell.Aside background
bg-surface    → Card, Modal background
border-default → theme.colors.gray[3] (light) / custom dark[4] (dark)
accent-blue   → theme.primaryColor shade [6]
```

Семантические цвета тревог маппятся как дополнительные `colors` в теме, чтобы их можно было использовать через `color="alarm-ok"` в Badge, Text и других компонентах.

---

## Цветовая семантика тревог

Строгий стандарт — одинаков в обеих темах, на мнемосхемах, в списке тревог, на графиках трендов и в Telegram-уведомлениях.

| Состояние | Hex | Цвет | Когда применяется |
|---|---|---|---|
| OK / Normal | `#4CAF50` | Зелёный | Значение в норме |
| Lo / Hi (Warning) | `#FFA726` | Янтарный | Превышение Lo/Hi уставки |
| LoLo / HiHi (Alarm) | `#EF5350` | Красный | Аварийное превышение |
| Uncertain | `#42A5F5` | Синий | Качество данных неопределённо |
| Bad Quality | `#78909C` | Серо-голубой | Данные недостоверны |
| Communication Loss | `#546E7A` | Тёмно-серый | Потеря связи с устройством |
| Offline | `#37474F` | Очень тёмный серый | Устройство выключено / недоступно |
| Acknowledged | `#7E57C2` | Фиолетовый | Тревога квитирована |

Цвета тревог **не меняются** между light и dark темами — они должны быть одинаково узнаваемы в обоих режимах.

---

## Типографика

| Назначение | Шрифт | Начертания |
|---|---|---|
| Основной текст, заголовки, формы | Inter | 400 Regular, 500 Medium, 600 Semi Bold |
| Значения тегов, координаты, числа | JetBrains Mono | 400 Regular, 500 Medium |

Моноширинный шрифт обязателен для числовых значений — цифры не должны «прыгать» по ширине при обновлении.

### Размеры

| Контекст | Размер |
|---|---|
| Вторичная информация, бейджи | 11px |
| Таблицы, панели, боковые панели | 12–13px |
| Основной текст | 14px |
| Заголовки секций в панелях | 13px / Semi Bold (600) |
| Заголовки страниц | 18–20px / Semi Bold (600) |

### Маппинг на Mantine

```
theme.fontFamily        = "Inter, sans-serif"
theme.fontFamilyMonospace = "JetBrains Mono, monospace"
theme.fontSizes         = { xs: 11, sm: 13, md: 14, lg: 16, xl: 20 }
theme.headings.fontFamily = "Inter, sans-serif"
```

---

## Spacing

Базовая сетка — **8px**. Все отступы кратны 4px или 8px.

| Токен | Значение | Когда |
|---|---|---|
| `xs` | 4px | Минимальный gap между связанными элементами |
| `sm` | 8px | Межэлементный spacing внутри секции |
| `md` | 16px | Padding внутри панелей и карточек |
| `lg` | 24px | Межсекционный gap |
| `xl` | 32px | Крупные разделы |

Отступы внутри панелей — **12px** (исключение из кратности 8 ради плотности).

### Маппинг на Mantine

```
theme.spacing = { xs: "4px", sm: "8px", md: "16px", lg: "24px", xl: "32px" }
```

Компоненты: `<Stack gap="sm">`, `<Group gap={8}>`, `<Card p="md">`.

---

## Лейаут — AppShell

Лейаут строится на `AppShell` из Mantine, обёрнутом в кастомный компонент `Layout` (`shared/ui/Layout`).

### Слоты лейаута

```
┌──────────────────────────────────────────────┐
│  Header                                      │
├──────────┬────────────────────┬──────────────┤
│  Navbar  │                    │  Aside       │
│  или     │    Main (content)  │  (правая     │
│  Panel   │                    │   панель)    │
│  (левая  │                    │              │
│   панель)│                    │              │
├──────────┴────────────────────┴──────────────┤
│  Footer                                      │
└──────────────────────────────────────────────┘
```

### Два режима лейаута

editor-ui использует два принципиально разных набора слотов в зависимости от экрана:

**List mode** — страницы списков (объекты, устройства, теги, мнемосхемы):

| Слот | Содержимое |
|------|-----------|
| Header | Logo + заголовок страницы + иконка пользователя |
| Navbar | Навигационное дерево объектов |
| Main | Контент: таблицы, карточки, формы |
| Aside | Не используется |
| Footer | Не используется |

**Editor mode** — редактор мнемосхемы:

| Слот | Содержимое |
|------|-----------|
| Header | Logo + toolbar (Align, Grid, Snap…) + Publish + user |
| Panel | Палитра фигур + список слоёв |
| Main | Konva canvas |
| Aside | Вкладки: Properties, Data Binding |
| Footer | Toolbar: Undo/Redo, Zoom, Snap, статус сохранения |

Navbar и Panel **взаимоисключающие** — никогда не показываются одновременно.

### Размеры слотов

| Слот | base | md+ | Collapsible |
|------|------|-----|-------------|
| Header | 56px (height) | 64px | Нет |
| Navbar | 240px (width) | 280px | Да, по кнопке |
| Panel | 240px (width) | 280px | Да, по кнопке |
| Aside | 280px (width) | 320px | Да, по кнопке |
| Footer | 48px (height) | 56px | Нет |

При сворачивании панели её ширина → 0px, контентная область занимает освободившееся место.

---

## Компоненты

### Базовые (реэкспорт из Mantine через shared/ui)

| Компонент | Mantine-источник | Назначение |
|---|---|---|
| AppShell | `@mantine/core` | Каркас лейаута |
| Button | `Button` + кастомные варианты | Действия |
| TextInput | `TextInput` | Текстовые поля |
| Select | `Select` | Выбор из списка (с поиском для тегов) |
| Tabs | `Tabs` | Вкладки в правой панели |
| Modal | `Modal` | Модалки (создание объекта, подтверждение публикации) |
| Card | `Card` | Карточки в списках |
| Table | `Table` | Таблицы устройств, тегов |
| Stack | `Stack` | Вертикальные группы |
| Group | `Group` | Горизонтальные группы |
| Text | `Text` | Типографика |
| Switch | `Switch` | Переключатели (Snap, Grid) |
| Loader | `Loader` | Индикация загрузки |
| Notification | `Notification` | Тосты (сохранено / ошибка) |

### Кастомные варианты Button

| Вариант | Вид | Когда |
|---|---|---|
| `primary` | Заливка accent-blue | Главное действие: Publish, Save, Create |
| `secondary` | Обводка accent-blue | Второстепенное: Preview, Cancel |
| `danger` | Заливка красным | Удаление |
| `ghost` | Прозрачный фон, текст accent-blue | Toolbar-кнопки, действия в панелях |

### Кастомные (поверх Mantine)

| Компонент | Назначение | Каталог |
|---|---|---|
| Layout | AppShell-обёртка с collapse-логикой и двумя режимами | `shared/ui/Layout` |
| Header | Трёхзонная шапка (before / main / after) | `shared/ui/Layout/Header` |
| Panel | Левая панель с секциями top / center / bottom | `shared/ui/Layout/Panel` |
| Aside | Правая панель с секциями top / center | `shared/ui/Layout/Aside` |
| Footer | Нижняя панель | `shared/ui/Layout/Footer` |
| Navbar | Навигационная панель | `shared/ui/Layout/Navbar` |
| ColorPicker | Выбор цвета для Fill / Stroke в Properties | TBD |
| TagSelect | Select с поиском тега для Data Binding | TBD |
| AlarmBadge | Бейдж состояния тревоги с иконкой + цветом | TBD (viewer-ui) |

---

## Правила HMI (SCADA-специфика)

1. **Мигание** — только для активных неквитированных тревог LoLo/HiHi. Ничего другого не мигает.
2. **Цвет НЕ единственный носитель информации** — всегда дублируем иконкой/паттерном (accessibility, цветослепота).
3. **Моноширинный шрифт для значений** — обязательно. Цифры не должны «прыгать».
4. **Quality всегда рядом со значением** — бейдж OK/Hi/HiHi/Uncertain/Bad/Offline рядом с числом.
5. **Тёмная тема по умолчанию для viewer-ui** — диспетчерские в полутени, глаза работают часами.
6. **Светлая тема по умолчанию для editor-ui** — инженер работает при офисном освещении.
7. **Плотность информации** — SCADA исторически плотная. Не раздувать отступами. Использовать Mantine `size="xs"` / `size="sm"` для компонентов панелей.
8. **Стабильное позиционирование** — значения не должны сдвигать соседние элементы при обновлении. Моноширинный шрифт + фиксированная ширина контейнеров значений.

---

## Связанные документы

- [`README.md`](README.md) — обзор раздела дизайна.
- [`Information-Architecture.md`](Information-Architecture.md) — sitemap и навигационная модель editor-ui.
- [`Figure-Types.md`](Figure-Types.md) — набор типов фигур и контракт figure_params.
- [`HMI-Visual-Language.md`](HMI-Visual-Language.md) — TBD, детальное описание отображения оборудования.
- [`Components.md`](Components.md) — TBD, каталог кастомных компонентов (детали).
- [`../09_План-работ/README.md`](../09_План-работ/README.md) — Phase 0 описывает полный scope дизайн-фазы.
- [`../../specs/0001.ready.editor-ui-layout-design-system.md`](../../specs/0001.ready.editor-ui-layout-design-system.md) — спецификация на реализацию.
