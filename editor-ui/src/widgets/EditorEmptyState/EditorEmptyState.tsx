import { IconPlus } from "@tabler/icons-react";

import styles from "./EditorEmptyState.module.css";

export const EditorEmptyState = () => {
  return (
    <div className={styles.emptyState}>
      <div className={styles.dashedSquare}>
        <IconPlus size={28} color="var(--mantine-color-deepBlue-6)" />
      </div>
      <p className={styles.title}>Перетащите фигуру из палитры</p>
      <div className={styles.hints}>
        <kbd>R</kbd> Rect <kbd>C</kbd> Circle <kbd>L</kbd> Line <kbd>T</kbd> Text
      </div>
    </div>
  );
};
