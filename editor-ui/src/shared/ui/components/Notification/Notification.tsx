import { Notification as MantineNotification } from "@mantine/core";

import styles from "./Notification.module.css";
import type { NotificationProps } from "./types";

export const Notification = (props: NotificationProps) => {
  return (
    <MantineNotification
      {...props}
      classNames={{
        root: styles.root,
        title: styles.title,
        description: styles.description
      }}
    />
  );
};
