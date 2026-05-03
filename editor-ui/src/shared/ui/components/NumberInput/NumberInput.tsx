import { NumberInput as MantineNumberInput } from "@mantine/core";

import type { NumberInputProps } from "./types";
import styles from "../TextInput/TextInput.module.css";

const joinClassNames = (...classNames: Array<string | undefined>) =>
  classNames.filter(Boolean).join(" ");

export const NumberInput = ({ mono = false, ...props }: NumberInputProps) => {
  return (
    <MantineNumberInput
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
