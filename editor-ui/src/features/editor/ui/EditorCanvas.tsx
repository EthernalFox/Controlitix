import { useElementSize } from "@mantine/hooks";
import type Konva from "konva";
import type { KonvaEventObject } from "konva/lib/Node";
import { useEffect, useRef } from "react";
import { Layer, Stage, Text } from "react-konva";

import { DEFAULT_FIGURE_PARAMS, useFiguresStore } from "@entities/figures";
import { fitToFrame, screenToCanvas, zoomAtPoint } from "@features/editor/lib/canvasUtils";
import { useEditorStore } from "@features/editor/model/editorStore";

import { FigureRenderer } from "./FigureRenderer";
import { FrameGuide } from "./FrameGuide";
import { SelectionLayer } from "./SelectionLayer";

interface EditorCanvasProps {
  diagramId: string;
}

const isEditableTarget = (target: EventTarget | null) => {
  if (!(target instanceof HTMLElement)) {
    return false;
  }

  if (target.isContentEditable) {
    return true;
  }

  const tagName = target.tagName;
  return tagName === "INPUT" || tagName === "TEXTAREA" || tagName === "SELECT";
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

  const {
    activeTool,
    clearSelection,
    frame,
    panX,
    panY,
    selectedFigureIds,
    selectFigure,
    setActiveTool,
    setCursor,
    setPan,
    setSaveStatus,
    setZoom,
    zoom
  } = useEditorStore();

  const {
    addFigure,
    figures,
    removeFigure,
    updateFigure
  } = useFiguresStore();

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

      if (event.key !== "Delete" || isEditableTarget(event.target)) {
        return;
      }

      const selectedId = useEditorStore.getState().selectedFigureIds[0];
      if (!selectedId) {
        return;
      }

      event.preventDefault();
      useEditorStore.getState().setSaveStatus("saving");

      void useFiguresStore
        .getState()
        .removeFigure(selectedId)
        .then(() => {
          useEditorStore.getState().clearSelection();
          useEditorStore.getState().setSaveStatus("saved");
        })
        .catch(() => {
          useEditorStore.getState().setSaveStatus("error");
        });
    };

    const onKeyUp = (event: KeyboardEvent) => {
      if (event.code === "Space") {
        isSpacePressedRef.current = false;
      }
    };

    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("keyup", onKeyUp);

    return () => {
      window.removeEventListener("keydown", onKeyDown);
      window.removeEventListener("keyup", onKeyUp);
    };
  }, [removeFigure]);

  const persistParams = async (figureId: string, params: Record<string, unknown>) => {
    setSaveStatus("saving");
    try {
      await updateFigure(figureId, { params });
      setSaveStatus("saved");
    } catch {
      setSaveStatus("error");
    }
  };

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
      const nextViewport = zoomAtPoint(
        zoom,
        { x: panX, y: panY },
        pointer,
        event.evt.deltaY
      );
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

    if (
      event.evt.button === 1 ||
      (event.evt.button === 0 && isSpacePressedRef.current)
    ) {
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
      setPan(
        panStartRef.current.initialPanX + diffX,
        panStartRef.current.initialPanY + diffY
      );
      return;
    }

    const pointer = stage.getPointerPosition();
    if (!pointer) {
      return;
    }

    const canvasPoint = screenToCanvas(pointer.x, pointer.y, zoom, { x: panX, y: panY });
    setCursor(canvasPoint.x, canvasPoint.y);
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

  const canvasCursor = activeTool === "select" ? "default" : "crosshair";

  return (
    <div
      ref={containerRef}
      style={{
        width: "100%",
        height: "100%",
        background: "#F5F5F5",
        cursor: panStartRef.current ? "grabbing" : canvasCursor
      }}
    >
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
        onDblClick={handleDoubleClick}
      >
        <Layer listening={false}>
          <FrameGuide frame={frame} />
          {figures.length === 0 && (
            <Text
              x={32}
              y={32}
              text="Добавьте фигуру из палитры слева"
              fontSize={16}
              fill="#868E96"
            />
          )}
        </Layer>

        <Layer>
          {figures.map((figure) => (
            <FigureRenderer
              key={figure.id}
              figure={figure}
              isSelected={selectedFigureIds.includes(figure.id)}
              onSelect={selectFigure}
              onDragEnd={(id, x, y) => void persistParams(id, { x, y })}
              onTransformEnd={(id, params) => void persistParams(id, params)}
            />
          ))}
        </Layer>

        <Layer>
          <SelectionLayer stageRef={stageRef} selectedFigureIds={selectedFigureIds} />
        </Layer>
      </Stage>
    </div>
  );
};
