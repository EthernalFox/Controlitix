import { TextInput as MantineTextInput } from "@mantine/core";

import type { TextInputProps } from "./types";

export const TextInput = (props: TextInputProps) => {
  return <MantineTextInput {...props} />;
};
