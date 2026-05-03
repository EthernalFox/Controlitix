import type { CSSProperties, ReactNode } from "react";

import styles from "./CanvasGrid.module.css";

interface CanvasGridProps {
  children: ReactNode;
  enabled: boolean;
  zoom: number;
  cursor: string;
}

const joinClasses = (...classes: Array<string | undefined | false>) =>
  classes.filter(Boolean).join(" ");

export const CanvasGrid = ({ children, enabled, zoom, cursor }: CanvasGridProps) => {
  return (
    <div
      className={joinClasses(styles.canvasContainer, !enabled && styles.gridOff)}
      style={{
        "--grid-size": `${Math.max(8, 20 * zoom)}px`,
        cursor
      } as CSSProperties}
    >
      {children}
    </div>
  );
};
