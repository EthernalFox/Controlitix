import assert from "node:assert/strict";
import test from "node:test";

import type { DiagramSnapshotPoint } from "@/features/diagram";
import { applyDynamics } from "@/widgets/FigureRenderer/lib/applyDynamics";
import { parseParams } from "@/widgets/FigureRenderer/lib/parseParams";

const baseParams = parseParams({
  x: 10,
  y: 10,
  width: 120,
  height: 40,
  fill: "#112233",
  stroke: "#445566",
  text: "Static",
  visible: true
});

const makePoint = (
  overrides: Partial<DiagramSnapshotPoint>
): DiagramSnapshotPoint => {
  return {
    tagId: "tag-1",
    ts: "2026-04-25T10:00:00Z",
    v: 12.345,
    q: "ok",
    ...overrides
  };
};

test("applyDynamics returns base props when dynamics is absent", () => {
  const resolved = applyDynamics(baseParams, undefined, undefined, true);

  assert.equal(resolved.text, "Static");
  assert.equal(resolved.visible, true);
  assert.equal(resolved.fill, "#112233");
  assert.equal(resolved.blink, false);
});

test("applyDynamics uses bad quality and dash when value is missing", () => {
  const params = parseParams({
    ...baseParams,
    dynamics: {
      fillByQuality: {
        bad: "#990000"
      },
      textFromValue: {
        format: "%.1f"
      }
    }
  });

  const resolved = applyDynamics(params, undefined, undefined, true);

  assert.equal(resolved.fill, "#990000");
  assert.equal(resolved.text, "—");
});

test("applyDynamics keeps quality on null value and sets text to dash", () => {
  const params = parseParams({
    ...baseParams,
    dynamics: {
      fillByQuality: {
        comm_loss: "#334455"
      },
      textFromValue: {
        format: "%.2f"
      }
    }
  });

  const resolved = applyDynamics(
    params,
    makePoint({ q: "comm_loss", v: null }),
    undefined,
    true
  );

  assert.equal(resolved.fill, "#334455");
  assert.equal(resolved.text, "—");
});

test("applyDynamics falls back to params.fill when quality key is missing", () => {
  const params = parseParams({
    ...baseParams,
    dynamics: {
      fillByQuality: {
        ok: "#00ff00"
      }
    }
  });

  const resolved = applyDynamics(
    params,
    makePoint({ q: "hihi", v: 1 }),
    undefined,
    true
  );

  assert.equal(resolved.fill, "#112233");
});

test("applyDynamics hides figure when quality is not in visibility filter", () => {
  const params = parseParams({
    ...baseParams,
    dynamics: {
      visibilityByQuality: ["ok", "hi"]
    }
  });

  const resolved = applyDynamics(
    params,
    makePoint({ q: "comm_loss", v: 4 }),
    undefined,
    true
  );

  assert.equal(resolved.visible, false);
});

test("applyDynamics formats text and appends unit", () => {
  const params = parseParams({
    ...baseParams,
    dynamics: {
      textFromValue: {
        format: "%.1f",
        appendUnit: true
      }
    }
  });

  const resolved = applyDynamics(
    params,
    makePoint({ v: 71.28, q: "ok" }),
    "°C",
    true
  );

  assert.equal(resolved.text, "71.3 °C");
});
