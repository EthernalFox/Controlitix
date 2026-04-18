import { Tooltip as MantineTooltip } from "@mantine/core";

import type { TooltipProps } from "./types";

export const Tooltip = (props: TooltipProps) => {
  return <MantineTooltip {...props} />;
};
