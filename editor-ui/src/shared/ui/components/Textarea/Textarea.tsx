import { Textarea as MantineTextarea } from "@mantine/core";

import type { TextareaProps } from "./types";
import styles from "../TextInput/TextInput.module.css";

const joinClassNames = (...classNames: Array<string | undefined>) =>
  classNames.filter(Boolean).join(" ");

export const Textarea = ({ mono = false, ...props }: TextareaProps) => {
  return (
    <MantineTextarea
      {...props}
      size={props.size ?? "sm"}
      classNames={{
        label: styles.label,
        input: joinClassNames(styles.input, mono ? styles.monoInput : undefined),
        description: styles.description,
        error: styles.error
      }}
    />
  );
};
