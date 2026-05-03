import { PasswordInput as MantinePasswordInput } from "@mantine/core";

import type { PasswordInputProps } from "./types";
import styles from "../TextInput/TextInput.module.css";

const joinClassNames = (...classNames: Array<string | undefined>) =>
  classNames.filter(Boolean).join(" ");

export const PasswordInput = ({ mono = false, ...props }: PasswordInputProps) => {
  return (
    <MantinePasswordInput
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
