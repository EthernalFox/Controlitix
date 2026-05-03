import { Drawer as MantineDrawer } from "@mantine/core";

import type { DrawerProps } from "./types";
import modalStyles from "../Modal/Modal.module.css";

export const Drawer = ({ overlayProps, ...props }: DrawerProps) => {
  return (
    <MantineDrawer
      {...props}
      classNames={{
        content: modalStyles.content,
        body: modalStyles.body,
        header: modalStyles.header,
        title: modalStyles.title
      }}
      overlayProps={{
        ...overlayProps,
        className: overlayProps?.className
          ? `${modalStyles.overlay} ${overlayProps.className}`
          : modalStyles.overlay
      }}
    />
  );
};
