# Handoff: Controlitix editor-ui — Design System (Liquid Glass)

## Overview

Дизайн-система и набор экранов для **editor-ui** — инженерного приложения Controlitix (SCADA-редактор мнемосхем). Покрывает:

- Foundations (токены light/dark, типографика, spacing, alarm-семантика).
- Components (Button, Input, Badge, Tabs, Switch, TagSelect, Card, Modal).
- Layout (AppShell в двух режимах: List и Editor).
- Editor — детально (canvas states, Data Binding, контекстное меню, save/publish states).
- Screens (Login, таблица устройств, Tag Drawer, системные экраны 403/404/500/Session, empty + skeleton).

Стилистика — **жидкое стекло** (liquid glass): backdrop-filter blur, полупрозрачные фоны, мягкие тени, ambient pastel orbs на фоне. Применяется только на chrome (Header / Panel / Aside / Footer / Modal / Popover) — карточки данных, таблицы и значения тегов остаются плоскими ради SCADA-плотности и читаемости.

## About the Design Files

Файлы в этой папке — **дизайн-референс в HTML**. Это прототип, показывающий внешний вид, поведение и токены, **не production-код для копирования**.

Задача — воспроизвести эти дизайны в существующем кодбейсе `editor-ui` (React 19 + Mantine 8 + Konva), используя установленные паттерны:

- Все базовые компоненты должны быть **Mantine 8**, обёрнутые через адаптерный слой `shared/ui/components`.
- Кастомные варианты (`Button` с variants `primary`/`secondary`/`danger`/`ghost`) — поверх Mantine через `createTheme()`.
- Liquid glass — это **визуальный слой поверх токенов**, реализуется через CSS-переменные и `backdrop-filter`. Токены палитры (`bg-primary`, `accent-blue`, alarms) **не меняются** — они уже зафиксированы в Design-System.md.

## Fidelity

**High-fidelity.** Все цвета, размеры, шрифты, радиусы, отступы — финальные. Hex-значения, размеры в px, font-family указаны точно. Liquid glass-параметры (blur, opacity, shadow) тоже финальные.

## Reference docs (already in repo)

В `reference/`:

- `Design-System.md` — источник правды для токенов, тем, spacing, типографики, AppShell, alarms.
- `Design-Brief-editor-ui.md` — техзадание на дизайн (scope экранов, состояния, acceptance criteria).
- `Figure-Types.md` — whitelist 11 типов фигур редактора.
- `Information-Architecture.md` — sitemap и роуты.

⚠ Brief в п. 10 запрещает glassmorphism. Эта работа — **сознательное отклонение**: жидкое стекло применено по запросу заказчика только на chrome, без ущерба читаемости данных. Зафиксировать в `docs/06_Дизайн/Design-Decisions.md`.

## Design Tokens

### Light theme (editor-ui default)

```
--bg-primary:       #FFFFFF
--bg-secondary:     #F8F9FA
--bg-surface:       #FFFFFF
--bg-elevated:      #F1F3F5
--border-default:   #DEE2E6
--text-primary:     #212529
--text-secondary:   #495057
--text-muted:       #868E96
--accent-blue:      #228BE6
--canvas-bg:        #F5F5F5
--toolbar-bg:       #FAFAFA
```

### Dark theme

```
--bg-primary:       #111422
--bg-secondary:     #1A1D2E
--bg-surface:       #232740
--bg-elevated:      #2D3154
--border-default:   #3A3F5C
--text-primary:     #E8EAF0
--text-secondary:   #9CA3B8
--text-muted:       #6B7394
--accent-blue:      #4C9AFF
--canvas-bg:        #111422
--toolbar-bg:       #232740
```

### Alarm semantics (invariant in both themes)

```
OK / Normal:        #4CAF50  · круг
Lo / Hi:            #FFA726  · треугольник
LoLo / HiHi:        #EF5350  · ромб (мигает)
Uncertain:          #42A5F5  · полая окружность
Bad Quality:        #78909C  · штриховка
Comm Loss:          #546E7A  · разорванная окружность
Offline:            #37474F  · затемнённый квадрат
Acknowledged:       #7E57C2  · круг с галочкой
```

Цвет НИКОГДА не единственный носитель — всегда форма иконки/паттерн.

### Spacing

```
xs: 4px    sm: 8px    md: 16px    lg: 24px    xl: 32px
```
Padding внутри SCADA-панелей — 12px (исключение из 8-сетки).

### Radius

- `xs`: 2px (мелкие элементы)
- `sm`: 4px (defaultRadius Mantine — все формы, карточки данных)
- `md`: 8px (бейджи, шорткат-чипы)
- `lg`: 14–16px (chrome лейаута, modal, popover, glass-panel)

### Typography

- **Inter** — UI и заголовки. Веса 400 / 500 / 600.
- **JetBrains Mono** — значения тегов, координаты, identifier'ы, addresses. Веса 400 / 500.

| Размер | Контекст |
|---|---|
| 11px | бейджи, метаданные |
| 12–13px | таблицы, панели, секции |
| 14px | основной текст |
| 13px / 600 | заголовки секций |
| 18–20px / 600 | заголовки страниц |
| 32px / 600 | hero / 4xx-коды |

### Liquid glass primitives

```css
--glass-bg:         rgba(255, 255, 255, 0.55);  /* light */
--glass-bg-strong:  rgba(255, 255, 255, 0.72);  /* для inputs */
--glass-bg-soft:    rgba(255, 255, 255, 0.32);  /* подложки */
--glass-border:     rgba(255, 255, 255, 0.7);
--glass-blur:       saturate(180%) blur(20px);
--glass-shadow:
  0 1px 0 rgba(255,255,255,0.7) inset,
  0 -1px 0 rgba(0,0,0,0.04) inset,
  0 8px 32px -8px rgba(15, 23, 42, 0.18),
  0 2px 8px -2px rgba(15, 23, 42, 0.08);
--glass-highlight:  linear-gradient(180deg, rgba(255,255,255,0.5) 0%, rgba(255,255,255,0) 50%);
```

Dark-вариант — те же ключи, но bg-rgba опускается до 35–55%, борд в `rgba(255,255,255,0.08)`, тень глубже.

`.glass` — базовый класс: `background: var(--glass-bg)` + `backdrop-filter: var(--glass-blur)` + border + shadow + 14px radius. У псевдоэлемента `::before` лежит light-highlight градиент.

**Где применять glass:** Header, Panel, Aside, Footer, Modal, Popover, Dropdown, FAB, бейдж-чипы внутри toolbar'а.

**Где НЕ применять:** строки таблиц, ячейки данных, превью значений тегов, тело drawer'а с формой, числовые badges. Эти поверхности — flat (`bg-surface` / `bg-secondary`).

## Layout — AppShell

Каркас — Mantine `AppShell`, обёрнутый в `shared/ui/Layout`. Два **взаимоисключающих** режима:

### List mode (страницы списков)

| Слот | Размер | Содержимое |
|---|---|---|
| Header | 56px / 64px@md | logo + breadcrumbs + search + user |
| Navbar | 240px / 280px@md, collapsible | Дерево объектов с tabs `Overview / Devices / Tags / Diagrams` |
| Main | flex | таблицы / карточки / формы |
| Aside | — | не используется |
| Footer | — | не используется |

### Editor mode (`/diagrams/:id/edit`)

| Слот | Размер | Содержимое |
|---|---|---|
| Header | 56–64px | logo + breadcrumbs + Undo/Redo + Grid/Snap toggles + autosave status + Publish + user |
| Panel (left) | 240/280px, collapsible | Палитра 11 фигур (3-col grid) + LayerList |
| Main | flex | Konva canvas с grid background и transformer |
| Aside (right) | 280/320px, collapsible | Tabs `Properties` / `Data Binding` |
| Footer | 48–56px | Undo/Redo + Grid/Snap switches + Zoom slider + save status |

Navbar и Panel **никогда не показываются вместе** — это ключевое правило AppShell-логики.

## Components

### Button — 4 варианта (все glass)

| Вариант | Mantine | Применение |
|---|---|---|
| `primary` | `variant="filled" color="deepBlue"` | Publish, Save, Create — главное действие |
| `secondary` | `variant="outline" color="deepBlue"` | Preview, Cancel — второстепенное |
| `danger` | `variant="filled" color="red"` | Delete |
| `ghost` | `variant="subtle" color="deepBlue"` | Toolbar-иконки, действия в панелях |

Все варианты — на `backdrop-filter: blur(20px)` подложке с полупрозрачным фоном (см. CSS в `Controlitix Design System.html`, селекторы `.btn-primary` / `.btn-secondary` / `.btn-danger` / `.btn-ghost`).

Размеры: дефолт padding `6px 14px`, font 13px/500. `btn-sm` — `4px 10px`/12px. `btn-icon` — 30×30 квадрат.

### Input

- Высота: 30px (size="sm" в Mantine).
- Background: `var(--glass-bg-strong)` + 1px `var(--glass-border-inner)` + radius 8px.
- Focus: `border: 1px solid var(--accent-blue)` + `box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-blue) 18%, transparent)`.
- Error: то же, но через `--alarm-crit`.
- Для чисел/адресов/identifier'ов — класс `input-mono` (font-family JetBrains Mono).

### AlarmBadge

```
.badge {
  font-size: 11px / 600;
  padding: 3px 8px;
  border-radius: 999px;
  font-family: JetBrains Mono;
  background: color-mix(in srgb, var(--bg-color) 15%, transparent);
  color: var(--bg-color);
  border: 1px solid var(--bg-color);
}
```

Иконка (`.icon` 8×8) — обязательная, форма зависит от типа тревоги (см. Alarm semantics).

### Tabs

Pill-сегмент: контейнер на `var(--glass-bg-soft)` + 4px padding + 1px inner border + radius 10px. Активный таб — `var(--bg-surface)` + лёгкая тень. В dark — `rgba(76,154,255,0.18)`.

### Switch

Track 32×18 / radius 999px. Off — `var(--glass-bg-soft)`. On — `linear-gradient(180deg, #3A9BEE, #228BE6)`. Knob 12×12 с тенью, transition 0.2s.

### Modal

Glass-card в центре viewport'а на dimmed backdrop'е (`rgba(0,0,0,0.06)` + `backdrop-filter: blur(2px)` для всего поля). Card — `glass` + `padding: 18px` + radius 14px + max-width 240–280px.

3 состояния публикации: `saving` (spinner) → `published` (зелёный круг ✓) → `error` (красный круг !).

### Card (data)

**Flat**, без glass — `background: var(--bg-surface)` + `border: 1px solid var(--border-default)` + `radius: sm (4px)` + `padding: 18px`. Применяется для DiagramCard, ObjectCard, табличных строк.

### TagSelect

Search input + список с подсказкой. Item: 6×8 padding, radius 6px. Активный — `rgba(34,139,230,0.12)` + 1px accent border. Справа от tag_name — unit (моно, alarm-цвет).

## Editor canvas — детали

### Состояния

| State | Визуал |
|---|---|
| Empty | Centered prompt с dashed-квадратом 64×64, accent ＋, текст «Перетащите фигуру из палитры», ниже — keyboard hints `R · C · L · T` |
| Default (no select) | Aside показывает свойства схемы (имя, описание, dimensions). Properties и Data Binding неактивны. |
| Single select | Selected фигура с dashed outline 2px accent + 4 corner handles (8×8 white + 1.5px accent border). Aside — Properties + Data Binding активны. |
| Multi-select (≥2) | Bounding box dashed `var(--accent-blue)` + soft fill `rgba(34,139,230,0.05)`. Properties — только общие поля (fill). Data Binding скрыт. |
| Drag in progress | Тень под фигурой + snap-линии (1px accent dashed) при включённом Snap. |
| Save error banner | В footer: `rgba(239,83,80,0.1)` + 1px alarm-crit border + retry-кнопка. |
| Connection lost | Top banner `rgba(255,167,38,0.1)` + `⚡` иконка + текст «Соединение потеряно · автосохранение приостановлено». |

### Canvas grid

Background:
```
linear-gradient(to right, color-mix(in srgb, var(--text-muted) 18%, transparent) 1px, transparent 1px),
linear-gradient(to bottom, color-mix(in srgb, var(--text-muted) 18%, transparent) 1px, transparent 1px);
background-size: 20px 20px;
```

### Cursor coords overlay

Сверху-слева canvas'а — glass-чип с `font-family: JetBrains Mono` 11px: `cursor: 142.0, 88.5`.

### Контекстное меню (ПКМ)

Glass-strong popover, items 7×10 padding / 12px / radius 6px. Hover: `var(--glass-bg-soft)`. Шорткаты справа моно-цветом muted. Раздел разделён 1px hairline. Delete — alarm-crit.

Items: Дублировать (⌘D) · Привязать к тегу (B) · На передний план · На задний план · — · Удалить (Del).

### Keyboard shortcuts

| Действие | Шорткат |
|---|---|
| Undo / Redo | ⌘Z / ⌘⇧Z |
| Удалить | Del |
| Дублировать | ⌘D |
| Снять выделение | Esc |
| Zoom in/out | + / − |
| Fit to screen | ⌘0 |
| Опубликовать | ⌘↑ |
| Привязать тег | B |

## Screens

### Login

Полноэкранная stage с pastel-orbs background (4 orbs blur 80px) → центрирована glass-card 380px / radius 18px / padding 32px. Логотип 36×36 с gradient mark, h1 «С возвращением, инженер», 2 inputs, primary-кнопка full-width.

### Devices table

Glass-card обёртка с inner padding=0. Шапка таблицы (`thead th`): font 10px / 600 / uppercase / letter-spacing 0.08em / `var(--glass-bg-soft)`. Строки: `td` font 12px, hover `var(--glass-bg-soft)`. Первая ячейка — статус-точка 8×8 alarm-цветом. id и endpoint — моно.

Toolbar таблицы: title + search input + tab-pills (Все / Modbus / SNMP) + primary-кнопка `+ Устройство`.

### Tag Drawer

Glass-strong panel 380px width, открывается справа поверх dimmed page. Полная форма:

- Имя тега, Устройство (Select)
- Адрес (per-protocol): для Modbus — `Тип регистра` + `Address`; для SNMP — `OID`.
- Data type, Unit
- Scaling: `raw_min` / `raw_max` / `eng_min` / `eng_max` (4 моно-инпута)
- Setpoints: 4 моно-инпута `LoLo / Lo / Hi / HiHi`, каждый окрашен alarm-цветом 30%-границей.
- Footer: ghost Cancel + primary Save.

### System screens (4 columns)

Каждый — glass-card / radius 14px / padding 24px / centered:

- **403** — gradient-text «403» (#FFA726→#F08F0E), title «Нет доступа», secondary-button.
- **404** — gradient-text «404» (#42A5F5→#1976D2), title «Не найдено», secondary-button.
- **500** — gradient-text «500» (#EF5350→#C62828), `err_id` моно, secondary «Повторить».
- **Session expired** — иконка ⏱ alarm-ack, primary «Войти».

### Empty / Loading

- Empty: centered иконка 56×56 в glass-soft, заголовок 13px/500, описание 11px muted, primary-кнопка `+ Создать …`.
- Loading skeleton: 3 grid-card 120px height с shimmer-анимацией (`@keyframes shimmer` 1.4s linear infinite, gradient `glass-soft → glass-strong → glass-soft`, `background-size: 200% 100%`).

## Existing codebase mapping

`editor-ui/src/`:

- `shared/ui/components/*` — адаптерный слой над Mantine. Здесь живут `Button`, `TextInput`, `Tabs`, `Modal` и т.д. — они уже есть, нужно только обновить стили под liquid glass через theme + CSS-переменные.
- `shared/ui/Layout/*` — Layout / Header / Navbar / Panel / Aside / Footer. Структура совпадает, нужно добавить glass-классы.
- `widgets/EditorToolbar` — Header в Editor mode.
- `widgets/ShapePalette` — палитра (3-col grid из 11 фигур).
- `widgets/LayerList` — слои.
- `widgets/PropertiesPanel` — Aside с Properties и Data Binding tabs.
- `widgets/EditorFooter` — Footer редактора.
- `widgets/DiagramCard` — карточка схемы (flat).
- `widgets/TagDrawer`, `widgets/DeviceDrawer` — формы.

## Implementation strategy

1. **Tokens layer** — завести CSS-переменные из секции «Design Tokens» в `app/index.css`, привязать к `data-theme="light|dark"` атрибуту `<body>`.
2. **Glass primitive** — создать utility-класс `.glass` / `.glass-strong` / `.glass-soft` (или MantineProvider styles override на AppShell-слоты).
3. **Mantine theme** — обновить `createTheme()`: `primaryColor: "deepBlue"`, custom `colors`, `fontFamily: "Inter"`, `fontFamilyMonospace: "JetBrains Mono"`, `defaultRadius: "sm"`, `spacing` из токенов.
4. **AppShell-обёртки** — добавить class на каждый слот, разделить List / Editor mode props.
5. **Component variants** — обновить `Button.tsx` чтобы glass-стиль шёл из `theme.components.Button.styles`. Аналогично `TextInput`, `Modal`.
6. **Alarm system** — `AlarmBadge` компонент с map (state → color + icon shape). Использовать `quality` бейджи на превью значений в Data Binding.
7. **Editor canvas** — Konva grid layer + transformer styling, drag preview, snap lines.

## Files in this bundle

- `Controlitix Design System.html` — полный hi-fi прототип (5 разделов, light/dark toggle).
- `reference/Design-System.md` — токены и правила HMI.
- `reference/Design-Brief-editor-ui.md` — техзадание.
- `reference/Figure-Types.md` — 11 типов фигур.
- `reference/Information-Architecture.md` — sitemap.

## Accessibility checklist (DoD из brief'а п. 12)

- Контраст текст/фон ≥ 4.5:1 (WCAG AA) во всех состояниях, обеих темах. Glass-фон **не отменяет** это требование — за glass должен быть контрастный backdrop.
- Hit area ≥ 32×32px для всех интерактивных элементов.
- Focus-стили нарисованы для всех инпутов и кнопок (`box-shadow` ring через accent-blue 18%).
- Цвет никогда не единственный носитель — alarms продублированы формой.
- Все анимации ≤ 200ms (исключение — мигание HiHi alarms в viewer-ui).
