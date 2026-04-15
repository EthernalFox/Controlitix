import { Stack as MantineStack } from "@mantine/core";

import type { StackProps } from "./types";

export const Stack = (props: StackProps) => {
  return <MantineStack {...props} />;
};

