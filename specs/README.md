# specs/

Task specifications for Claude → Codex workflow.

## Filename convention

```
NNNN.STATUS.short-description.md
```

- `NNNN` — zero-padded sequence number, e.g. `0001`
- `STATUS` — current state (see below)
- `short-description` — kebab-case, max 5 words

## Statuses

| Status      | Meaning                                      | Who sets it      |
|-------------|----------------------------------------------|------------------|
| `draft`     | Claude is writing the spec, not ready yet    | Claude           |
| `ready`     | Spec complete, waiting for Codex             | Claude           |
| `wip`       | Codex is working on it                       | Codex / human    |
| `review`    | Codex finished, waiting for Claude review    | Codex / human    |
| `done`      | Reviewed and accepted                        | Claude / human   |
| `cancelled` | Will not be implemented                      | Claude / human   |

## Lifecycle

```
draft → ready → wip → review → done
                             ↘ cancelled
```

To change status: rename the file (change the STATUS segment).

## Spec template

See `0000.done.spec-template.md` for the standard spec structure.
