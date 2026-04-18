import { Badge as MantineBadge } from "@mantine/core";

import type { BadgeProps } from "./types";

export const Badge = (props: BadgeProps) => {
  return <MantineBadge {...props} />;
};
