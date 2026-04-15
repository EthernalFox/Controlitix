import type { SwitchProps as MantineSwitchProps } from "@mantine/core";

export type SwitchVariant = "primary" | "danger";

export interface SwitchProps
  extends Omit<MantineSwitchProps, "color" | "onChange"> {
  variant?: SwitchVariant;
  onChange?: (checked: boolean) => void;
}

