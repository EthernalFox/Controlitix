import { Button as MantineButton } from "@mantine/core";

import styles from "./Button.module.css";
import type { ButtonProps } from "./types";

const variantToMantineProps: Record<
  NonNullable<ButtonProps["variant"]>,
  { variant: "filled" | "outline" | "subtle"; color: string; className: string }
> = {
  primary: { variant: "filled", color: "deepBlue", className: styles.btnPrimary },
  secondary: { variant: "outline", color: "deepBlue", className: styles.btnSecondary },
  danger: { variant: "filled", color: "alarmCrit", className: styles.btnDanger },
  ghost: { variant: "subtle", color: "deepBlue", className: styles.btnGhost }
};

const joinClassNames = (...classNames: Array<string | undefined>) =>
  classNames.filter(Boolean).join(" ");

export const Button = ({ variant = "primary", children, className, ...props }: ButtonProps) => {
  const mantineProps = variantToMantineProps[variant];

  return (
    <MantineButton
      {...props}
      variant={mantineProps.variant}
      color={mantineProps.color}
      className={joinClassNames(styles.root, mantineProps.className, className)}
    >
      {children}
    </MantineButton>
  );
};

