import { Textarea as MantineTextarea } from "@mantine/core";

import type { TextareaProps } from "./types";

export const Textarea = (props: TextareaProps) => {
  return <MantineTextarea {...props} />;
};

