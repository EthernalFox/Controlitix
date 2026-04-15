import { Select as MantineSelect } from "@mantine/core";

import type { SelectProps } from "./types";

export const Select = (props: SelectProps) => {
  return <MantineSelect {...props} />;
};
