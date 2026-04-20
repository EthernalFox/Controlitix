import { NumberInput as MantineNumberInput } from "@mantine/core";

import type { NumberInputProps } from "./types";

export const NumberInput = (props: NumberInputProps) => {
  return <MantineNumberInput {...props} />;
};

