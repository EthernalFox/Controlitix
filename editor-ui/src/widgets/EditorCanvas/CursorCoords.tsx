import type { Point } from "@features/editor";

import styles from "./CursorCoords.module.css";

interface CursorCoordsProps {
  cursor: Point | null;
}

export const CursorCoords = ({ cursor }: CursorCoordsProps) => {
  if (!cursor) {
    return null;
  }

  return (
    <div className={styles.cursorChip}>
      cursor:
      <span className={styles.mono}>{cursor.x.toFixed(1)}</span>,
      <span className={styles.mono}>{cursor.y.toFixed(1)}</span>
    </div>
  );
};
