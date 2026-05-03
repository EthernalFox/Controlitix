import { TextInput as MantineTextInput } from "@mantine/core";

import styles from "./TextInput.module.css";
import type { TextInputProps } from "./types";

const joinClassNames = (...classNames: Array<string | undefined>) =>
  classNames.filter(Boolean).join(" ");

export const TextInput = ({ mono = false, ...props }: TextInputProps) => {
  return (
    <MantineTextInput
      {...props}
      size={props.size ?? "sm"}
      classNames={{
        label: styles.label,
        input: joinClassNames(styles.input, mono ? styles.monoInput : undefined),
        description: styles.description,
        error: styles.error,
        section: styles.section
      }}
    />
  );
};
