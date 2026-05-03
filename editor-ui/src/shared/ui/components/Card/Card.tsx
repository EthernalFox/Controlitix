import { Card as MantineCard } from "@mantine/core";

import styles from "./Card.module.css";
import type { CardProps } from "./types";

const joinClassNames = (...classNames: Array<string | undefined>) =>
  classNames.filter(Boolean).join(" ");

const variantToClassName = {
  flat: styles.flat,
  glass: styles.glass
} as const;

export const Card = ({ variant = "flat", className, withBorder, ...props }: CardProps) => {
  return (
    <MantineCard
      {...props}
      withBorder={withBorder ?? variant === "flat"}
      className={joinClassNames(variantToClassName[variant], className)}
    />
  );
};

