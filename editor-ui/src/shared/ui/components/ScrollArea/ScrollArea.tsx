import { ScrollArea as MantineScrollArea } from "@mantine/core";

import type { ScrollAreaProps } from "./types";

export const ScrollArea = (props: ScrollAreaProps) => {
  return <MantineScrollArea {...props} />;
};
