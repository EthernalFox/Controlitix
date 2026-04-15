import type { ButtonProps as MantineButtonBaseProps } from "@mantine/core";
import type { ComponentPropsWithoutRef, ReactNode } from "react";

export type ButtonVariant = "primary" | "secondary" | "danger" | "ghost";

type NativeButtonProps = ComponentPropsWithoutRef<"button">;

export interface ButtonProps
  extends Omit<MantineButtonBaseProps, "variant" | "color" | "children">,
    Omit<NativeButtonProps, "children" | "className" | "color" | "style"> {
  variant?: ButtonVariant;
  children?: ReactNode;
}
