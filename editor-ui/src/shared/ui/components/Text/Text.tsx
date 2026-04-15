import { Text as MantineText } from "@mantine/core";

import type { TextProps } from "./types";

export const Text = (props: TextProps) => {
  return <MantineText {...props} />;
};

