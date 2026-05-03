import type { ReactNode } from "react";

import styles from "./PageToolbar.module.css";

interface PageToolbarProps {
  title: string;
  search?: ReactNode;
  tabs?: ReactNode;
  primaryAction?: ReactNode;
}

export const PageToolbar = ({ title, search, tabs, primaryAction }: PageToolbarProps) => {
  return (
    <div className={styles.root}>
      <div className={styles.top}>
        <h2 className={styles.title}>{title}</h2>
        {primaryAction}
      </div>

      {(search || tabs) && (
        <div className={styles.controls}>
          {search}
          {tabs}
        </div>
      )}
    </div>
  );
};
