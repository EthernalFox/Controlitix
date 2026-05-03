import { useMantineColorScheme } from "@mantine/core";
import { useElementSize } from "@mantine/hooks";
import type Konva from "konva";
import type { KonvaEventObject } from "konva/lib/Node";
import { useEffect, useMemo, useRef, useState } from "react";
import { Layer, Stage } from "react-konva";

import { DEFAULT_FIGURE_PARAMS, useFiguresStore } from "@entities/figures";
import {
  computeSnapTargets,
  fitToFrame,
  screenToCanvas,
  useEditorStore,
  useThemeColor,
  zoomAtPoint
} from "@features/editor";
import type { SnapGuide, SnapRect } from "@features/editor";
import { CanvasGrid, CursorCoords, SelectionFrame, SnapGuides } from "@widgets/EditorCanvas";
import { EditorContextMenu } from "@widgets/EditorContextMenu";
import { EditorEmptyState } from "@widgets/EditorEmptyState";

import { FigureRenderer } from "./FigureRenderer";
import { FrameGuide } from "./FrameGuide";

interface EditorCanvasProps {
  diagramId: string;
}

interface ContextMenuState {
  x: number;
  y: number;
}

const reorderBySelected = (
  ids: string[],
  selectedIds: string[],
  direction: "front" | "back"
) => {
  const selectedSet = new Set(selectedIds);
  const rest = ids.filter((id) => !selectedSet.has(id));

  if (direction === "front") {
    return [...rest, ...selectedIds.filter((id) => ids.includes(id))];
  }

  return [...selectedIds.filter((id) => ids.includes(id)), ...rest];
};

const collectRects = (stage: Konva.Stage, figureIds: string[]): SnapRect[] => {
  return figureIds
    .map((figureId) => {
      const node = stage.findOne(`#figure-${figureId}`);
      if (!node) {
        return null;
      }

      const rect = node.getClientRect({ skipShadow: true });
      return {
        id: figureId,
        x: rect.x,
        y: rect.y,
        width: rect.width,
        height: rect.height
      };
    })
    .filter((item): item is SnapRect => Boolean(item));
};

export const EditorCanvas = ({ diagramId }: EditorCanvasProps) => {
  const stageRef = useRef<Konva.Stage | null>(null);
  const { ref: containerRef, width: containerWidth, height: containerHeight } = useElementSize();
  const panStartRef = useRef<{
    startX: number;
    startY: number;
    initialPanX: number;
    initialPanY: number;
  } | null>(null);
  const isSpacePressedRef = useRef(false);
  const didFitRef = useRef(false);
  const cursorThrottleRef = useRef<number>(0);

  const [snapGuides, setSnapGuides] = useState<SnapGuide[]>([]);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  const accentBlue = useThemeColor("--mantine-color-deepBlue-6");
  const multiFill =
    colorScheme === "dark"
      ? "rgba(76, 154, 255, 0.08)"
      : "rgba(34, 139, 230, 0.05)";

  const {
    activeTool,
    clearSelection,
    cursor,
    dragging,
    frame,
    gridEnabled,
    panX,
    panY,
    selection,
    selectFigure,
    setActiveTool,
    setCursor,
    setDragging,
    setPan,
    setSaveStatus,
    setSelection,
    setZoom,
    snapEnabled,
    zoom
  } = useEditorStore();

  const { addFigure, figures, removeFigure, reorderFigures, updateFigure } = useFiguresStore();

  const selectionVersion = useMemo(
    () => figures.map((figure) => `${figure.id}:${figure.updatedAt}`).join("|"),
    [figures]
  );

  useEffect(() => {
    didFitRef.current = false;
  }, [diagramId]);

  useEffect(() => {
    if (!containerWidth || !containerHeight || didFitRef.current) {
      return;
    }

    const fitted = fitToFrame(containerWidth, containerHeight, frame);
    setZoom(fitted.zoom);
    setPan(fitted.panX, fitted.panY);
    didFitRef.current = true;
  }, [containerHeight, containerWidth, frame, setPan, setZoom]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.code === "Space") {
        isSpacePressedRef.current = true;
      }
    };

    const onKeyUp = (event: KeyboardEvent) => {
      if (event.code === "Space") {
        isSpacePressedRef.current = false;
      }
    };

    const onFitToScreen = () => {
      if (!containerWidth || !containerHeight) {
        return;
      }

      const fitted = fitToFrame(containerWidth, containerHeight, frame);
      setZoom(fitted.zoom);
      setPan(fitted.panX, fitted.panY);
    };

    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("keyup", onKeyUp);
    window.addEventListener("editor:fit-to-screen", onFitToScreen as EventListener);

    return () => {
      window.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("keyup", onKeyUp);
      window.removeEventListener("editor:fit-to-screen", onFitToScreen as EventListener);
    };
  }, [containerHeight, containerWidth, frame, setPan, setZoom]);

  const persistParams = async (figureId: string, params: Record<string, unknown>) => {
    setSaveStatus("saving");
    try {
      await updateFigure(figureId, { params });
      setSaveStatus("saved");
    } catch {
      setSaveStatus("error");
    }
  };

  const removeSelected = async () => {
    if (selection.ids.length === 0) {
      return;
    }

    setSaveStatus("saving");
    try {
      await Promise.all(selection.ids.map((id) => removeFigure(id)));
      clearSelection();
      setSaveStatus("saved");
    } catch {
      setSaveStatus("error");
    }
  };

  const duplicateSelected = async () => {
    const selectedFigures = figures.filter((figure) => selection.ids.includes(figure.id));
    if (selectedFigures.length === 0) {
      return;
    }

    setSaveStatus("saving");
    try {
      const created = await Promise.all(
        selectedFigures.map((figure) => {
          const x = typeof figure.params.x === "number" ? figure.params.x : 0;
          const y = typeof figure.params.y === "number" ? figure.params.y : 0;

          return addFigure(diagramId, {
            type: figure.type,
            tag_id: figure.tagId ?? undefined,
            params: {
              ...DEFAULT_FIGURE_PARAMS[figure.type],
              ...figure.params,
              x: x + 16,
              y: y + 16
            }
          });
        })
      );

      setSelection(created.map((figure) => figure.id));
      setSaveStatus("saved");
    } catch {
      setSaveStatus("error");
    }
  };

  const moveSelected = (direction: "front" | "back") => {
    if (selection.ids.length === 0) {
      return;
    }

    const orderedIds = reorderBySelected(
      figures.map((figure) => figure.id),
      selection.ids,
      direction
    );
    reorderFigures(orderedIds);
  };

  useEffect(() => {
    const onDelete = () => {
      void removeSelected();
    };

    const onDuplicate = () => {
      void duplicateSelected();
    };

    const onRetrySave = () => {
      setSaveStatus("saving");
      window.setTimeout(() => {
        setSaveStatus("saved");
      }, 350);
    };

    const onUndo = () => {
      setSaveStatus("idle");
    };

    const onRedo = () => {
      setSaveStatus("idle");
    };

    window.addEventListener("editor:delete-selection", onDelete as EventListener);
    window.addEventListener("editor:duplicate-selection", onDuplicate as EventListener);
    window.addEventListener("editor:retry-save", onRetrySave as EventListener);
    window.addEventListener("editor:undo", onUndo as EventListener);
    window.addEventListener("editor:redo", onRedo as EventListener);

    return () => {
      window.removeEventListener("editor:delete-selection", onDelete as EventListener);
      window.removeEventListener("editor:duplicate-selection", onDuplicate as EventListener);
      window.removeEventListener("editor:retry-save", onRetrySave as EventListener);
      window.removeEventListener("editor:undo", onUndo as EventListener);
      window.removeEventListener("editor:redo", onRedo as EventListener);
    };
  }, [duplicateSelected, removeSelected, setSaveStatus]);

  const handleWheel = (event: KonvaEventObject<WheelEvent>) => {
    event.evt.preventDefault();

    const stage = stageRef.current;
    if (!stage) {
      return;
    }

    const pointer = stage.getPointerPosition();
    if (!pointer) {
      return;
    }

    if (event.evt.ctrlKey || event.evt.metaKey) {
      const nextViewport = zoomAtPoint(zoom, { x: panX, y: panY }, pointer, event.evt.deltaY);
      setZoom(nextViewport.zoom);
      setPan(nextViewport.panX, nextViewport.panY);
      return;
    }

    if (event.evt.shiftKey) {
      setPan(panX - event.evt.deltaY, panY);
      return;
    }

    setPan(panX, panY - event.evt.deltaY);
  };

  const handleMouseDown = (event: KonvaEventObject<MouseEvent>) => {
    const stage = stageRef.current;
    if (!stage) {
      return;
    }

    setContextMenu(null);

    if (event.evt.button === 2) {
      return;
    }

    if (event.evt.button === 1 || (event.evt.button === 0 && isSpacePressedRef.current)) {
      panStartRef.current = {
        startX: event.evt.clientX,
        startY: event.evt.clientY,
        initialPanX: panX,
        initialPanY: panY
      };
      return;
    }

    if (activeTool === "select") {
      if (event.target === stage) {
        clearSelection();
      }
      return;
    }

    if (event.target !== stage) {
      return;
    }

    const pointer = stage.getPointerPosition();
    if (!pointer) {
      return;
    }

    const canvasPoint = screenToCanvas(pointer.x, pointer.y, zoom, { x: panX, y: panY });
    const payload = {
      type: activeTool,
      params: {
        ...DEFAULT_FIGURE_PARAMS[activeTool],
        x: canvasPoint.x,
        y: canvasPoint.y
      }
    };

    setSaveStatus("saving");
    void addFigure(diagramId, payload)
      .then((figure) => {
        selectFigure(figure.id);
        setActiveTool("select");
        setSaveStatus("saved");
      })
      .catch(() => {
        setSaveStatus("error");
      });
  };

  const handleMouseMove = (event: KonvaEventObject<MouseEvent>) => {
    const stage = stageRef.current;
    if (!stage) {
      return;
    }

    if (panStartRef.current) {
      const diffX = event.evt.clientX - panStartRef.current.startX;
      const diffY = event.evt.clientY - panStartRef.current.startY;
      setPan(panStartRef.current.initialPanX + diffX, panStartRef.current.initialPanY + diffY);
      return;
    }

    const pointer = stage.getPointerPosition();
    if (!pointer) {
      return;
    }

    const now = performance.now();
    if (now - cursorThrottleRef.current < 33) {
      return;
    }

    cursorThrottleRef.current = now;
    const canvasPoint = screenToCanvas(pointer.x, pointer.y, zoom, { x: panX, y: panY });
    setCursor(canvasPoint);
  };

  const handleMouseUp = () => {
    panStartRef.current = null;
  };

  const handleDoubleClick = () => {
    if (!containerWidth || !containerHeight) {
      return;
    }

    const fitted = fitToFrame(containerWidth, containerHeight, frame);
    setZoom(fitted.zoom);
    setPan(fitted.panX, fitted.panY);
  };

  const handleFigureDragStart = (id: string, event: KonvaEventObject<DragEvent>) => {
    if (!selection.ids.includes(id)) {
      setSelection([id]);
    }

    setDragging({
      ids: [id],
      from: { x: event.target.x(), y: event.target.y() }
    });
  };

  const handleFigureDragMove = (id: string, event: KonvaEventObject<DragEvent>) => {
    if (!snapEnabled) {
      setSnapGuides([]);
      return;
    }

    const stage = stageRef.current;
    if (!stage) {
      return;
    }

    const rects = collectRects(
      stage,
      figures.map((figure) => figure.id)
    );
    const result = computeSnapTargets(rects, id, 4, 200);

    if (result.deltaX !== 0 || result.deltaY !== 0) {
      event.target.x(event.target.x() + result.deltaX);
      event.target.y(event.target.y() + result.deltaY);
    }

    setSnapGuides(result.guides);
  };

  const handleFigureDragEnd = (id: string, x: number, y: number) => {
    setSnapGuides([]);
    setDragging(null);
    void persistParams(id, { x, y });
  };

  const handleFigureContextMenu = (id: string, event: KonvaEventObject<PointerEvent>) => {
    event.evt.preventDefault();

    if (!selection.ids.includes(id)) {
      setSelection([id]);
    }

    setContextMenu({ x: event.evt.clientX, y: event.evt.clientY });
  };

  const canvasCursor = panStartRef.current
    ? "grabbing"
    : activeTool === "select"
      ? "default"
      : "crosshair";

  return (
    <div ref={containerRef} style={{ width: "100%", height: "100%", position: "relative" }}>
      <CanvasGrid enabled={gridEnabled} zoom={zoom} cursor={canvasCursor}>
        <CursorCoords cursor={cursor} />
        {figures.length === 0 && <EditorEmptyState />}

        <Stage
          ref={stageRef}
          width={Math.max(1, containerWidth)}
          height={Math.max(1, containerHeight)}
          x={panX}
          y={panY}
          scaleX={zoom}
          scaleY={zoom}
          onWheel={handleWheel}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={() => setCursor(null)}
          onDblClick={handleDoubleClick}
          onContextMenu={(event) => {
            if (event.target === event.target.getStage()) {
              event.evt.preventDefault();
              setContextMenu(null);
            }
          }}
        >
          <Layer listening={false}>
            <FrameGuide frame={frame} />
          </Layer>

          <Layer>
            {figures.map((figure) => (
              <FigureRenderer
                key={figure.id}
                figure={figure}
                isSelected={selection.ids.includes(figure.id)}
                isDragging={dragging?.ids.includes(figure.id) ?? false}
                onSelect={selectFigure}
                onContextMenu={handleFigureContextMenu}
                onDragStart={handleFigureDragStart}
                onDragMove={handleFigureDragMove}
                onDragEnd={handleFigureDragEnd}
                onTransformEnd={(id, params) => void persistParams(id, params)}
              />
            ))}
          </Layer>

          <Layer listening={false}>
            <SelectionFrame
              stageRef={stageRef}
              selectedFigureIds={selection.ids}
              borderColor={accentBlue}
              multiFill={multiFill}
              version={selectionVersion}
            />
          </Layer>

          <SnapGuides guides={snapGuides} stroke={accentBlue} />
        </Stage>
      </CanvasGrid>

      <EditorContextMenu
        opened={Boolean(contextMenu)}
        x={contextMenu?.x ?? 0}
        y={contextMenu?.y ?? 0}
        onClose={() => setContextMenu(null)}
        onDuplicate={() => void duplicateSelected()}
        onBindTag={() => {
          setContextMenu(null);
          window.dispatchEvent(new Event("editor:open-binding"));
        }}
        onBringToFront={() => {
          setContextMenu(null);
          moveSelected("front");
        }}
        onSendToBack={() => {
          setContextMenu(null);
          moveSelected("back");
        }}
        onDelete={() => {
          setContextMenu(null);
          void removeSelected();
        }}
      />
    </div>
  );
};
  const { colorScheme } = useMantineColorScheme();
