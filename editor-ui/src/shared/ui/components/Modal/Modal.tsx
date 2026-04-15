import { Modal as MantineModal } from "@mantine/core";

import type { ModalProps } from "./types";

export const Modal = (props: ModalProps) => {
  return <MantineModal {...props} />;
};
