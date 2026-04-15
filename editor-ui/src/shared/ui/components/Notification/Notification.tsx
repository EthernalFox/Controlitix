import { Notification as MantineNotification } from "@mantine/core";

import type { NotificationProps } from "./types";

export const Notification = (props: NotificationProps) => {
  return <MantineNotification {...props} />;
};
