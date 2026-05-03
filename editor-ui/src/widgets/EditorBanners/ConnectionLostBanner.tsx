import { IconBolt } from "@tabler/icons-react";

import styles from "./ConnectionLostBanner.module.css";

export const ConnectionLostBanner = () => {
  return (
    <div className={styles.connectionLostBanner}>
      <IconBolt size={14} />
      <span>Соединение потеряно · автосохранение приостановлено</span>
    </div>
  );
};
