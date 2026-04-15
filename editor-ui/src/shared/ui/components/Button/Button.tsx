import {
  Button as MantineButton
} from "@mantine/core";

import type { ButtonProps } from "./types";

const variantToMantineProps: Record<
  NonNullable<ButtonProps["variant"]>,
  { variant: "filled" | "outline" | "subtle"; color: string }
> = {
  primary: { variant: "filled", color: "deepBlue" },
  secondary: { variant: "outline", color: "deepBlue" },
  danger: { variant: "filled", color: "red" },
  ghost: { variant: "subtle", color: "deepBlue" }
};

export const Button = ({ variant = "primary", children, ...props }: ButtonProps) => {
  const mantineProps = variantToMantineProps[variant];
  return (
    <MantineButton {...props} variant={mantineProps.variant} color={mantineProps.color}>
      {children}
    </MantineButton>
  );
};
