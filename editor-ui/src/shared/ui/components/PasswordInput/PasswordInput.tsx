import { PasswordInput as MantinePasswordInput } from "@mantine/core";

import type { PasswordInputProps } from "./types";

export const PasswordInput = (props: PasswordInputProps) => {
  return <MantinePasswordInput {...props} />;
};

