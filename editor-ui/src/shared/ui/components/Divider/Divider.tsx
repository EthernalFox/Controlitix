import { Divider as MantineDivider } from "@mantine/core";

import type { DividerProps } from "./types";

export const Divider = (props: DividerProps) => {
  return <MantineDivider {...props} />;
};
