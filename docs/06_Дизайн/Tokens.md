# Справочник токенов — Controlitix

Токены разделены на два источника:
- **`var(--mantine-*)`** — стандартные Mantine 8 CSS-переменные. Переключаются автоматически при смене `colorScheme`. Использовать в CSS-модулях для поверхностей, текста, границ, spacing, radius.
- **`var(--ctrx-*)`** — расширение через `cssVariablesResolver`. Только для того, чего Mantine не покрывает: alarm-цвета, glass morphism, canvas, орбы, размеры лейаута.
- **`theme.other.alarm`** — типизированный JS-доступ к alarm-цветам через `useMantineTheme().other.alarm`.

**Запрещено:** hex / rgba() напрямую в `.module.css`. Старые `--bg-*`, `--text-*`, `--border-*`, `--accent-*` — удалены (spec 0047).

---

## Mantine native vars — справка

Основные переменные, автоматически переключаемые Mantine:

| CSS var | Light | Dark | Назначение |
|---|---|---|---|
| `--mantine-color-body` | `#fff` | dark[7] | Фон страницы |
| `--mantine-color-default` | `#fff` | dark[6] | Фон карточек, панелей |
| `--mantine-color-default-hover` | gray[0] | dark[5] | Hover-состояние |
| `--mantine-color-default-border` | gray[4] | dark[4] | Разделители, обводки |
| `--mantine-color-text` | black | white | Основной текст |
| `--mantine-color-dimmed` | gray[6] | dark[2] | Вторичный текст |
| `--mantine-color-placeholder` | gray[5] | dark[3] | Плейсхолдеры, метки |
| `--mantine-color-anchor` | deepBlue[6] | deepBlue[4] | Ссылки |
| `--mantine-color-error` | red[6] | red[4] | Ошибки |
| `--mantine-color-bright` | black | white | Инверсный текст |
| `--mantine-color-deepBlue-{0-9}` | — | — | Акцентный синий (10 шейдов) |
| `--mantine-font-family` | `Inter, sans-serif` | | Основной шрифт |
| `--mantine-font-family-monospace` | `JetBrains Mono, monospace` | | Моноширинный |
| `--mantine-spacing-{xs/sm/md/lg/xl}` | 4/8/16/24/32px | | Отступы |
| `--mantine-radius-{xs/sm/md/lg/xl}` | 2/4/8/14/18px | | Скругления |

---

## Alarm-цвета — `--ctrx-alarm-*`

**Theme-invariant** — одинаковы в обеих темах. Размещены в блоке `variables` resolver'а.

| CSS var | Hex | Цвет | Состояние |
|---|---|---|---|
| `--ctrx-alarm-ok` | `#4caf50` | Зелёный | OK / Normal |
| `--ctrx-alarm-warn` | `#ffa726` | Янтарный | Lo / Hi (Warning) |
| `--ctrx-alarm-crit` | `#ef5350` | Красный | LoLo / HiHi (Alarm) |
| `--ctrx-alarm-uncertain` | `#42a5f5` | Синий | Uncertain |
| `--ctrx-alarm-bad` | `#78909c` | Серо-голубой | Bad Quality |
| `--ctrx-alarm-comm` | `#546e7a` | Тёмно-серый | Communication Loss |
| `--ctrx-alarm-offline` | `#37474f` | Очень тёмный серый | Offline |
| `--ctrx-alarm-ack` | `#7e57c2` | Фиолетовый | Acknowledged |

**JS-доступ** (в компонентах, Konva-рендерере, тестах):
```tsx
const theme = useMantineTheme();
const color = theme.other.alarm.crit; // '#ef5350'
```

**CSS-доступ** (в .module.css):
```css
.badge[data-state="crit"] { background: var(--ctrx-alarm-crit); }
```

> Мигание (`animation: blink`) — **только** для `crit` (LoLo/HiHi) неквитированных тревог. Ничего другого не мигает.

---

## Canvas — `--ctrx-canvas-*`

Тема-зависимые. Используются в Konva-холсте и CSS-оверлеях сетки.

| CSS var | Light | Dark | Назначение |
|---|---|---|---|
| `--ctrx-canvas-bg` | `#f5f5f5` | `#111422` | Фон Konva-холста |
| `--ctrx-canvas-grid` | `#dee2e6` | `#2d3154` | Линии сетки |

---

## Glass Morphism — `--ctrx-glass-*`

Тема-зависимые (кроме `--ctrx-glass-blur`). Применяются к chrome-элементам: Header, Panel, Aside, Footer, Modal, Popover (DD-001).

| CSS var | Light | Dark | Назначение |
|---|---|---|---|
| `--ctrx-glass-blur` | `saturate(180%) blur(20px)` | ← invariant | `backdrop-filter` значение |
| `--ctrx-glass-bg` | `rgba(255,255,255,0.55)` | `rgba(35,39,64,0.55)` | Основной фон |
| `--ctrx-glass-bg-strong` | `rgba(255,255,255,0.72)` | `rgba(35,39,64,0.78)` | Усиленный (modal) |
| `--ctrx-glass-bg-soft` | `rgba(255,255,255,0.32)` | `rgba(26,29,46,0.4)` | Мягкий (dropdown) |
| `--ctrx-glass-border` | `rgba(255,255,255,0.7)` | `rgba(255,255,255,0.08)` | Внешняя граница |
| `--ctrx-glass-border-inner` | `rgba(34,139,230,0.08)` | `rgba(76,154,255,0.18)` | Внутренняя граница |
| `--ctrx-glass-shadow` | _сложная тень_ | _сложная тень_ | `box-shadow` |
| `--ctrx-glass-highlight` | `linear-gradient(...)` | `linear-gradient(...)` | Верхний блик |

**Паттерн использования:**
```css
.glassPanel {
  background: var(--ctrx-glass-bg);
  backdrop-filter: var(--ctrx-glass-blur);
  -webkit-backdrop-filter: var(--ctrx-glass-blur);
  border: 1px solid var(--ctrx-glass-border);
  box-shadow: var(--ctrx-glass-shadow);
}
/* Псевдоэлемент для блика */
.glassPanel::before {
  content: "";
  position: absolute;
  inset: 0;
  background: var(--ctrx-glass-highlight);
  pointer-events: none;
  border-radius: inherit;
}
/* Fallback */
@supports not (backdrop-filter: blur(1px)) {
  .glassPanel { background: var(--mantine-color-default); }
}
@media (prefers-reduced-transparency: reduce) {
  .glassPanel { background: var(--mantine-color-default); backdrop-filter: none; }
}
```

---

## Ambient Orbs — `--ctrx-orb-*`

Тема-зависимые. Декоративные пятна в компоненте `AmbientOrbs`.

| CSS var | Light | Dark |
|---|---|---|
| `--ctrx-orb-1` | `#bad7ff` | `#1e3a6b` |
| `--ctrx-orb-2` | `#ffd6e8` | `#4c2a52` |
| `--ctrx-orb-3` | `#c8f0dc` | `#1f4738` |
| `--ctrx-orb-4` | `#ffe4b0` | `#5a4019` |

---

## Размеры лейаута — `--ctrx-header-h` и др.

Theme-invariant. Используются в CSS-модулях Layout-компонентов.

| CSS var | Значение | AppShell prop |
|---|---|---|
| `--ctrx-header-h` | `56px` | `header={{ height: 56 }}` |
| `--ctrx-panel-w` | `240px` | `navbar={{ width: 240 }}` |
| `--ctrx-aside-w` | `280px` | `aside={{ width: 280 }}` |
| `--ctrx-footer-h` | `48px` | `footer={{ height: 48 }}` |

> Mantine AppShell принимает числа в props (не CSS vars). Токены используются только для согласования в CSS-модулях.

---

## Как добавить новый токен

1. Спросить: Mantine уже покрывает это через `var(--mantine-color-*)`?  
   — Да → **не добавлять** `--ctrx-*`, использовать Mantine.  
   — Нет → добавить в `cssVariablesResolver` в `theme.ts`.

2. Выбрать блок:
   - `variables` — одинаково в обеих темах.
   - `light` / `dark` — зависит от темы.

3. Именовать: `--ctrx-<категория>-<имя>` (kebab-case).

4. Если нужен JS-доступ — добавить также в `theme.other` в `createTheme()`.

5. Задокументировать строку в этом файле.

---

## Связанные документы

- [`Design-System.md`](Design-System.md) — общая дизайн-система: типографика, spacing, HMI-правила.
- [`Design-Decisions.md`](Design-Decisions.md) — DD-001: glass morphism.
- [`editor-ui/src/shared/libs/theme/theme.ts`](../../editor-ui/src/shared/libs/theme/theme.ts) — исходник токенов.
- `specs/0047.ready.editor-ui-token-system-v2.md` — спека на реализацию.
