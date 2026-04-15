import { Switch as MantineSwitch } from "@mantine/core";

import type { SwitchProps } from "./types";

const variantToColor: Record<NonNullable<SwitchProps["variant"]>, string> = {
  primary: "deepBlue",
  danger: "red"
};

export const Switch = ({
  variant = "primary",
  onChange,
  ...props
}: SwitchProps) => {
  return (
    <MantineSwitch
      {...props}
      color={variantToColor[variant]}
      onChange={(event) => onChange?.(event.currentTarget.checked)}
    />
  );
};

