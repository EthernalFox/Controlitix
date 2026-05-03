import { IconCopy, IconLayersIntersect2, IconTrash } from "@tabler/icons-react";
import type { CSSProperties } from "react";

import { Menu } from "@shared/ui";

import styles from "./EditorContextMenu.module.css";

interface EditorContextMenuProps {
  opened: boolean;
  x: number;
  y: number;
  onClose: () => void;
  onDuplicate: () => void;
  onBindTag: () => void;
  onBringToFront: () => void;
  onSendToBack: () => void;
  onDelete: () => void;
}

const hiddenTargetStyle: CSSProperties = {
  position: "fixed",
  width: 1,
  height: 1,
  opacity: 0,
  pointerEvents: "none"
};

const shortcut = (value: string) => <span className={styles.editorContextMenuShortcut}>{value}</span>;

export const EditorContextMenu = ({
  opened,
  x,
  y,
  onClose,
  onDuplicate,
  onBindTag,
  onBringToFront,
  onSendToBack,
  onDelete
}: EditorContextMenuProps) => {
  return (
    <Menu
      opened={opened}
      onChange={(nextOpened) => {
        if (!nextOpened) {
          onClose();
        }
      }}
      withinPortal
      closeOnItemClick
      classNames={{
        dropdown: styles.editorContextMenu,
        item: styles.editorContextMenuItem
      }}
    >
      <Menu.Target>
        <button
          type="button"
          aria-hidden
          tabIndex={-1}
          style={{ ...hiddenTargetStyle, left: x, top: y }}
        />
      </Menu.Target>

      <Menu.Dropdown>
        <Menu.Item
          leftSection={<IconCopy size={14} />}
          rightSection={shortcut("Cmd/Ctrl+D")}
          onClick={onDuplicate}
        >
          Дублировать
        </Menu.Item>
        <Menu.Item
          leftSection={<IconLayersIntersect2 size={14} />}
          rightSection={shortcut("B")}
          onClick={onBindTag}
        >
          Привязать к тегу
        </Menu.Item>
        <Menu.Item onClick={onBringToFront}>На передний план</Menu.Item>
        <Menu.Item onClick={onSendToBack}>На задний план</Menu.Item>

        <Menu.Divider />

        <Menu.Item
          color="red"
          leftSection={<IconTrash size={14} />}
          rightSection={shortcut("Del")}
          className={styles.editorContextMenuDelete}
          onClick={onDelete}
        >
          Удалить
        </Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
};
