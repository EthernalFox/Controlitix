import { Badge as MantineBadge } from "@mantine/core";

import styles from "./Badge.module.css";
import type { BadgeProps } from "./types";

const joinClassNames = (...classNames: Array<string | undefined>) =>
  classNames.filter(Boolean).join(" ");

export const Badge = ({ className, variant, ...props }: BadgeProps) => {
  return (
    <MantineBadge
      {...props}
      variant={variant ?? "light"}
      className={joinClassNames(styles.root, className)}
    />
  );
};

