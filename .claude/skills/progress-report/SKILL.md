---
name: progress-report
description: Generate an MVP readiness report by analysing spec files under `specs/` and update the roadmap with a new dated section.
---

# Progress Report

Review all spec files under `specs/` and compute:

1. **Overall MVP completion %** — weighted by phase (P0=10%, P1=10%, P2=12%, P3=18%, P4=15%, P5=15%, P6=15%, P7=5% — see `docs/09_План-работ/README.md`).
2. **Backend vs frontend breakdown** — per service (`ms-editor`, `ms-poll`, `ms-auth`, `ms-viewer`) and per UI (`editor-ui`, `viewer-ui`).
3. **User scenario coverage** — table with 🟢 / 🟡 / 🔴 per сценарий из `docs/01_Видение-и-требования/` и `docs/07_Сценарии/`, с колонками «Что работает» / «Что отсутствует».

## Method

- Count specs by status (`done`, `review`, `rework`, `wip`, `ready`, `draft`) — parse filenames `NNNN.STATUS.*.md`.
- For each phase, verify реальную реализацию в коде (не только статус спеки): `ls` по `cmd/`, `internal/domain/`, `src/pages/`, миграциям. Статус спеки ≠ готовый код.
- Flag gaps that are НЕ покрыты спеками (e.g. integration-тесты, OpenAPI, autosave).

## Output

Update `docs/09_План-работ/README.md`:
- Раздел «Статус разработки» — обновить дату и общий процент.
- Таблица «Прогресс по спецификациям» — синхронизировать статусы.
- Строка «Сводка по статусам» — пересчитать.
- Таблица «Прогресс по фазам» — обновить проценты.

Язык отчёта — русский (см. `docs/` convention). Даты — абсолютные (`апрель 2026`, не «сейчас»).
