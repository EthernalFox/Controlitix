import { ColorInput as MantineColorInput } from "@mantine/core";

import type { ColorInputProps } from "./types";

export const ColorInput = (props: ColorInputProps) => {
  return <MantineColorInput {...props} />;
};
