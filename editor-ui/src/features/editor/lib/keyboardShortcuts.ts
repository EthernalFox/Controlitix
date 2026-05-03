import { useEffect } from "react";

import { useEditorStore } from "../model/editorStore";
import type { EditorTool } from "../model/types";

interface UseEditorShortcutsParams {
  onUndo: () => void;
  onRedo: () => void;
  onDelete: () => void;
  onDuplicate: () => void;
  onZoomIn: () => void;
  onZoomOut: () => void;
  onFitToScreen: () => void;
  onPublish: () => void;
  onBindTag: () => void;
}

const TOOL_SHORTCUTS: Array<{ key: string; tool: EditorTool }> = [
  { key: "r", tool: "rect" },
  { key: "c", tool: "circle" },
  { key: "l", tool: "line" },
  { key: "t", tool: "text" }
];

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

const isOverlayOpen = () =>
  Boolean(document.querySelector("[role='dialog'], .mantine-Popover-dropdown"));

export const useEditorShortcuts = ({
  onUndo,
  onRedo,
  onDelete,
  onDuplicate,
  onZoomIn,
  onZoomOut,
  onFitToScreen,
  onPublish,
  onBindTag
}: UseEditorShortcutsParams) => {
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (isEditableTarget(event.target)) {
        return;
      }

      const key = event.key.toLowerCase();
      const hasMod = event.metaKey || event.ctrlKey;
      const selectionSize = useEditorStore.getState().selection.ids.length;

      if (hasMod && key === "z" && event.shiftKey) {
        event.preventDefault();
        onRedo();
        return;
      }

      if (hasMod && key === "z") {
        event.preventDefault();
        onUndo();
        return;
      }

      if (hasMod && key === "d") {
        if (selectionSize === 0) {
          return;
        }

        event.preventDefault();
        onDuplicate();
        return;
      }

      if (hasMod && key === "0") {
        event.preventDefault();
        onFitToScreen();
        return;
      }

      if (hasMod && event.key === "ArrowUp") {
        event.preventDefault();
        onPublish();
        return;
      }

      if ((event.key === "Delete" || event.key === "Backspace") && selectionSize > 0) {
        event.preventDefault();
        onDelete();
        return;
      }

      if (event.key === "Escape") {
        if (isOverlayOpen()) {
          return;
        }

        useEditorStore.getState().clearSelection();
        return;
      }

      if (key === "+" || key === "=") {
        onZoomIn();
        return;
      }

      if (key === "-") {
        onZoomOut();
        return;
      }

      if (key === "b" && selectionSize === 1) {
        onBindTag();
        return;
      }

      const toolShortcut = TOOL_SHORTCUTS.find((item) => item.key === key);
      if (toolShortcut) {
        useEditorStore.getState().setActiveTool(toolShortcut.tool);
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [
    onBindTag,
    onDelete,
    onDuplicate,
    onFitToScreen,
    onPublish,
    onRedo,
    onUndo,
    onZoomIn,
    onZoomOut
  ]);
};
