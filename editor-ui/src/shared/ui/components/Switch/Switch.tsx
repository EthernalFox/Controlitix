import { Switch as MantineSwitch } from "@mantine/core";

import styles from "./Switch.module.css";
import type { SwitchProps } from "./types";

const variantToColor: Record<NonNullable<SwitchProps["variant"]>, string> = {
  primary: "deepBlue",
  danger: "alarmCrit"
};

export const Switch = ({ variant = "primary", onChange, ...props }: SwitchProps) => {
  return (
    <MantineSwitch
      {...props}
      color={variantToColor[variant]}
      size={props.size ?? "sm"}
      classNames={{
        root: styles.root,
        track: styles.track,
        thumb: styles.thumb
      }}
      onChange={(event) => onChange?.(event.currentTarget.checked)}
    />
  );
};
