import { Modal as MantineModal } from "@mantine/core";

import styles from "./Modal.module.css";
import type { ModalProps } from "./types";

export const Modal = ({ overlayProps, ...props }: ModalProps) => {
  return (
    <MantineModal
      {...props}
      classNames={{
        content: styles.content,
        body: styles.body,
        header: styles.header,
        title: styles.title
      }}
      overlayProps={{
        ...overlayProps,
        className: overlayProps?.className
          ? `${styles.overlay} ${overlayProps.className}`
          : styles.overlay
      }}
    />
  );
};
