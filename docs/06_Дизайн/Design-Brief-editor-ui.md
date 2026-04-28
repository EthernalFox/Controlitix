# Design Brief — editor-ui (для AI-агента Claude Design)

> Техническое задание на разработку high-fidelity дизайна `editor-ui` в Figma. Адресовано AI-агенту, который умеет читать существующие документы, расширять Figma-библиотеку и генерировать макеты по design-system токенам.

## 1. Цель работы

Получить набор production-ready макетов и интерактивных прототипов всех экранов `editor-ui`, готовых к реализации фронтенд-командой (React 19 + Mantine 8 + Konva). Дизайн должен быть полностью построен на уже зафиксированных токенах (см. п. 4) — никаких новых цветов/шрифтов/радиусов без обоснования.

Главный фокус MVP — **редактор мнемосхемы** (экран 11 в IA). Остальные экраны — поддержка инженерного потока «объект → устройства → теги → мнемосхема → публикация».

## 2. Контекст и обязательное чтение

Перед началом работы агент должен прочитать и не противоречить:

- [docs/06_Дизайн/Design-System.md](Design-System.md) — палитра, токены, типографика, лейаут (AppShell), правила HMI.
- [docs/06_Дизайн/Information-Architecture.md](Information-Architecture.md) — sitemap, роуты, приоритеты экранов.
- [docs/06_Дизайн/Figure-Types.md](Figure-Types.md) — whitelist типов фигур редактора (`rect`, `circle`, `ellipse`, `wedge`, `line`, `image`, `text`, `ring`, `arc`, `tag`, `path`).
- [docs/05_Сценарии-использования/Инженер_создать_мнемосхему.md](../05_Сценарии-использования/Инженер_создать_мнемосхему.md), [Инженер_добавить_устройство.md](../05_Сценарии-использования/Инженер_добавить_устройство.md), [Инженер_привязать_фигуры_к_тегам.md](../05_Сценарии-использования/Инженер_привязать_фигуры_к_тегам.md) — UX-флоу.
- [docs/02_Архитектура/Микросервисы/editor-ui.md](../02_Архитектура/Микросервисы/editor-ui.md) — функциональный скелет.
- [docs/04_API/ms-editor.md](../04_API/ms-editor.md) — DTO и эндпоинты, на которые завязаны формы.

Существующая Figma-библиотека: см. ссылки в `Design-System.md` («Figma-файлы»). Дизайн делать поверх неё, не создавать параллельную.

## 3. Целевая аудитория и контекст использования

- **Пользователь:** инженер АСУ ТП. Работает с десктопа в офисе, освещение нормальное → light theme по умолчанию. Опыт со SCADA, ожидает плотную информационную сетку.
- **Не пользователь:** оператор диспетчерской — это `viewer-ui`, его не трогаем.
- **Устройства:** разрешение от 1366×768 до 2560×1440. Mobile/tablet — вне scope MVP (но лейаут не должен ломаться при ресайзе до ~1280px).

## 4. Жёсткие ограничения

Дизайн **обязан** опираться на следующее, без отклонений:

| Ограничение | Источник |
|---|---|
| Light theme — по умолчанию для editor-ui; dark — обязательная альтернатива | `Design-System.md` § «Темы» |
| Палитра, токены, шрифты, spacing — только из `Design-System.md` | `Design-System.md` |
| Лейаут — Mantine `AppShell`, два режима (List / Editor), Navbar и Panel взаимоисключающие | `Design-System.md` § «Лейаут» |
| Семантика тревог (8 состояний) — одинакова в обеих темах | `Design-System.md` § «Цветовая семантика тревог» |
| Все базовые компоненты — Mantine 8 (без shadcn, без своей кнопки с нуля) | `Design-System.md` § «UI Kit» |
| Цвет НЕ единственный носитель информации | `Design-System.md` § «Правила HMI» |
| Моноширинный шрифт для значений тегов и координат | `Design-System.md` § «Правила HMI» |
| Sitemap и роуты — как в `Information-Architecture.md` | `Information-Architecture.md` |
| Список типов фигур — только из `Figure-Types.md` | `Figure-Types.md` |

## 5. Scope экранов

### In scope (MVP, обязательно)

Приоритет 1 (критический путь):

1. **Login** — форма входа, ошибки, session expired.
2. **Список объектов** — таблица/карточки, поиск, кнопка «Создать».
3. **Карточка объекта (Overview)** — шапка с метаданными, secondary nav.
4. **Устройства объекта** — таблица + drawer/modal формы устройства (создание/редактирование, типы протоколов: modbus_rtu, modbus_tcp, snmp_v1, snmp_v2c, snmp_v3 — поля разные).
5. **Теги объекта** — таблица с фильтрами по device/data_type/unit; drawer/modal формы тега (address per protocol, data_type, unit, scaling, setpoints).
6. **Список мнемосхем объекта** — карточки с превью, статус публикации.
7. **Редактор мнемосхемы (Editor mode)** — главный экран, см. § 6.
8. **Модалка подтверждения публикации** — статусы saving/published/error.

Приоритет 2 (системные состояния):

9. Empty state для каждого списка (объектов / устройств / тегов / мнемосхем).
10. Loading skeletons для тех же списков и для холста редактора.
11. 403 / 404 / 500 / Session expired — отдельные screens.

### Out of scope

- viewer-ui (оператор, тревоги, тренды).
- Глобальный поиск (`/search`) — опционально, только если останется бюджет.
- Mobile/tablet адаптив.
- Settings админского контура.

## 6. Детальные требования по экрану «Редактор мнемосхемы»

Это самый сложный экран, на нём фокус. Слоты лейаута — Editor mode из `Design-System.md`.

### Состав

| Зона | Что должно быть |
|---|---|
| Header | Логотип, имя объекта + имя схемы (хлебные крошки), переключатели Grid/Snap, кнопки Undo/Redo (дублируют footer), Publish (primary), статус автосохранения, аватар |
| Panel (left) | Палитра фигур (11 типов из `Figure-Types.md`, drag из палитры на canvas), под ней — список слоёв с reorder, видимостью, lock |
| Main | Konva canvas: фон, сетка (toggle), zoom, выделение, multi-select, transformer (resize/rotate) |
| Aside (right) | Tabs: **Properties** (геометрия, fill, stroke, opacity, текст для type=text) / **Data Binding** (привязка фигуры к тегу через TagSelect — поиск по имени тега, фильтр по device, отображение data_type и unit; превью текущего значения = «—» в editor-ui) |
| Footer | Undo/Redo, Zoom (slider + %, fit to screen, 100%), toggle Snap, toggle Grid, статус сохранения («Сохранено в 14:32» / «Сохранение…» / «Ошибка сохранения — повторить») |

### Состояния

- **Empty** — мнемосхема без фигур, на canvas — подсказка «Перетащите фигуру из палитры».
- **Default** — есть фигуры, ничего не выделено → правая панель показывает свойства схемы (имя, описание, dimensions).
- **Single select** — выделена одна фигура → Properties + Data Binding активны.
- **Multi select** — выделено ≥2 фигур → Properties показывает только общие свойства (например, fill), Data Binding скрыт (нельзя биндить пачкой в MVP).
- **Drag in progress** — визуальный фидбек (тень, snap-линии при включённом snap).
- **Saving** — индикатор в footer.
- **Save error** — банер «Не удалось сохранить, повторить» в footer + retry.
- **Publish in progress / success / error** — модалка из IA #12.
- **Connection lost (websocket / fetch error)** — баннер сверху, autosave приостанавливается.

### Acceptance criteria

- Видны и кликабельны: все 11 типов фигур в палитре, все элементы toolbar.
- Каждый из 11 figure types имеет default визуал (fill, stroke), отображается на canvas сразу после drop.
- Properties-tab показывает редактируемые поля для каждого типа (минимум: x, y, width, height, fill, stroke, strokeWidth, opacity; для text — text, fontSize, fontFamily; для line/path — points; для image — src; для tag — tag_id).
- Data Binding-tab имеет TagSelect с поиском (≥1 символ → фильтр), показывает device_name и unit рядом с tag_name.
- Footer корректно отображает 3 состояния save и 2 состояния publish.
- Все состояния (empty / default / single / multi / saving / save error / publish modal) присутствуют как отдельные frames в Figma.
- Light и Dark — оба варианта для каждого состояния.
- Прототип: drag фигуры с палитры → drop на canvas → выделение → переход на Properties; drag отдельным flow → drop на тег → подсветка bind.

## 7. Требования к остальным экранам (кратко)

Для каждого экрана из приоритета 1 (п. 5) сделать минимум:

- Default state.
- Empty state (если применимо).
- Loading skeleton.
- Error state.
- Hover/active для основных интерактивных элементов.
- Light + Dark.
- Адаптив 1280px / 1440px / 1920px (одна ключевая ширина — 1440px; на 1280 — без потери функциональности; на 1920 — без растягивания контента, центровка с max-width).

Для форм (устройство, тег) обязательно:

- Состояния валидации (поле в фокусе / ошибка / disabled / readonly).
- Per-протокол вариации полей (для устройства — modbus_rtu vs modbus_tcp vs snmp_v1/v2c/v3); для тега — address per device type (Modbus: register_type + address; SNMP: oid).
- Размещение setpoints (LoLo/Lo/Hi/HiHi) и scaling (raw_min/raw_max/eng_min/eng_max ИЛИ factor/offset) в форме тега.

## 8. Интерактивные паттерны

- **Collapsible panels (Navbar / Panel / Aside)** — кнопка-стрелка в header слота; ширина → 0 при сворачивании, контент перетекает.
- **Autosave** — статус всегда виден в footer редактора. Без явной кнопки Save в редакторе мнемосхемы (Save = автосохранение, Publish — отдельное действие).
- **Keyboard shortcuts** — заявить хотя бы: `Ctrl+Z` / `Ctrl+Shift+Z` (undo/redo), `Delete` (remove figure), `Ctrl+D` (duplicate), `Esc` (deselect), `+/-` (zoom). Полный список — отдельный artifact.
- **Drag-and-drop** — figures из палитры на canvas; reorder слоёв в LayerList.
- **Контекстные меню** — правый клик по фигуре: Duplicate, Delete, Bring to front, Send to back, Bind to tag.

## 9. HMI-специфика

- В editor-ui тревоги не «горят» (нет realtime данных) — но превью значений тегов в Data Binding должно использовать те же токены семантики, что и viewer-ui (см. `Design-System.md` § «Цветовая семантика тревог»), чтобы инженер видел, как будет выглядеть в проде.
- Моноширинный шрифт — для всех координат в Properties, для значений тегов в превью, для технических идентификаторов (object_id, device_id) в служебных областях.
- Quality-бейдж (OK / Hi / HiHi / Uncertain / Bad / Offline / Comm Loss / Acknowledged) — должен быть готов как компонент в Figma, даже если в editor-ui его не показываем (пригодится для viewer-ui и для общей библиотеки).

## 10. Технические ограничения для дизайнера

Дизайн должен **переводиться в код** на:

- **Mantine 8** компонентах. Перед изобретением своего компонента проверить, есть ли подходящий в Mantine. Если есть — стилизовать через тему, не рисовать с нуля.
- **CSS-токенах** дизайн-системы (см. п. 4). В Figma — переменные, в коде — `theme.other` / CSS vars.
- **Konva** для canvas (это влияет только на редактор: фигуры — это Konva-объекты, transformer — стандартный Konva.Transformer; учитывать его внешний вид при проектировании).

Запрещено:

- Использовать в макетах фотореалистичные тени, gradients, glassmorphism — SCADA-эстетика плотная и flat.
- Применять анимации длительностью > 200ms; единственное исключение — мигание тревог в viewer (вне scope этого брифа).
- Произвольная типографика. Только Inter + JetBrains Mono.
- Произвольные радиусы. Только токены `defaultRadius="sm"` (4px) и `xs`/`md`/`lg` из Mantine при необходимости.

## 11. Deliverables (что агент сдаёт)

В Figma:

1. **Pages в основном файле** (см. `Design-System.md` § «Figma-файлы»):
   - `editor-ui / Foundations` — токены, типографика, иконки, доказательство соответствия `Design-System.md` (если что-то добавлено — обоснование).
   - `editor-ui / Components` — все кастомные компоненты (Layout, Header, Panel, Aside, Footer, ColorPicker, TagSelect, AlarmBadge, ShapePalette, LayerList, PropertiesForm, DataBindingForm) как Figma Components с variants.
   - `editor-ui / Screens` — все экраны из § 5 (приоритет 1 + системные состояния), light + dark, ключевая ширина 1440px.
   - `editor-ui / States` — отдельная страница с матрицей всех состояний редактора (empty/default/single/multi/saving/error/publishing).
   - `editor-ui / Prototype` — кликабельный прототип критического пути: Login → Объекты → Объект → Мнемосхемы → Редактор → Publish.
2. **Variables / Styles** — все токены `Design-System.md` заведены как Figma Variables с light/dark modes.
3. **Code Connect mappings** (опционально, но желательно) — связь Figma-компонентов с `editor-ui/src/shared/ui/components/*` и `editor-ui/src/widgets/*`.

В репозитории:

4. **Обновление `docs/06_Дизайн/`** — если по ходу работы появились решения, не покрытые `Design-System.md` или `Information-Architecture.md`, агент дописывает их в эти файлы (PR + ревью архитектора).
5. **`docs/06_Дизайн/Components.md`** (TBD в `Design-System.md`) — закрыть: каталог кастомных компонентов с описанием props, variants, states, accessibility-нот.
6. **`docs/06_Дизайн/HMI-Visual-Language.md`** (TBD) — описать визуальный язык фигур, привязок и состояний; нужно даже для editor-ui, чтобы было от чего отталкиваться при viewer-ui.

## 12. Критерии готовности (definition of done)

Дизайн считается принятым, когда:

- [ ] Все экраны § 5 (приоритет 1) есть в Figma в light + dark, на ширине 1440px, с состояниями default / empty / loading / error.
- [ ] Все состояния редактора из § 6 присутствуют как frames.
- [ ] Все компоненты § 11.1 — Figma Components с variants и properties (не статические frames).
- [ ] Все цвета/шрифты/spacing в макетах подключены через Figma Variables, ни одного «оторванного» hex/px.
- [ ] Light/Dark проверены на каждом экране — палитра тревог одинакова в обеих темах.
- [ ] Прототип критического пути из § 11.1.4 запускается и доводит до экрана Publish success.
- [ ] Accessibility: контраст текст/фон ≥ 4.5:1 (WCAG AA) для всех состояний; интерактивные элементы ≥ 32×32 px hit area; focus-стили нарисованы для всех инпутов и кнопок.
- [ ] Каждое отклонение от `Design-System.md` (если оно есть) задокументировано и согласовано — файл-список в `docs/06_Дизайн/Design-Decisions.md`.
- [ ] Архитектор (Claude Code) и фронтенд-разработчик отметили готовность в комментариях к Figma-файлу.

## 13. Open questions (агент должен закрыть до старта)

- Карточка мнемосхемы — оставить как отдельный экран или схлопнуть в список схем? (см. `Information-Architecture.md` § Open questions)
- Формы устройства/тега — drawer или отдельный route? Решение фиксируется в `Components.md`.
- Глобальный поиск — включить в MVP или отложить?
- Settings объекта — нужен ли в MVP отдельной вкладкой?

Все ответы агент собирает в чате с архитектором перед началом hi-fi этапа.

## 14. Связанные документы

- [`Design-System.md`](Design-System.md), [`Information-Architecture.md`](Information-Architecture.md), [`Figure-Types.md`](Figure-Types.md), [`README.md`](README.md).
- [`../09_План-работ/README.md`](../09_План-работ/README.md) — Phase 0 (дизайн) — общий статус.
- [`../../specs/0001.ready.editor-ui-layout-design-system.md`](../../specs/0001.ready.editor-ui-layout-design-system.md) — спека на код-реализацию дизайн-системы (после согласования макетов).
- [`../../specs/0011.done.editor-ui-editor-page.md`](../../specs/0011.done.editor-ui-editor-page.md) — текущая код-реализация редактора (что уже есть в `editor-ui/`).
