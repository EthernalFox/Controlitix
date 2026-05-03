import type { ReactNode } from "react";

import { Badge } from "@shared/ui/components/Badge";

import styles from "./AlarmBadge.module.css";
import {
  CheckCircleIcon,
  CircleIcon,
  DiamondIcon,
  HatchedIcon,
  RingBrokenIcon,
  RingHollowIcon,
  SquareDimmedIcon,
  TriangleIcon
} from "./icons";

export type AlarmBadgeState =
  | "ok"
  | "warn"
  | "crit"
  | "uncertain"
  | "bad"
  | "comm"
  | "offline"
  | "ack";

type AlarmBadgeConfig = {
  icon: ReactNode;
  label: string;
};

const configByState: Record<AlarmBadgeState, AlarmBadgeConfig> = {
  ok: { icon: <CircleIcon />, label: "OK" },
  warn: { icon: <TriangleIcon />, label: "HI / LO" },
  crit: { icon: <DiamondIcon />, label: "HI HI / LO LO" },
  uncertain: { icon: <RingHollowIcon />, label: "UNC" },
  bad: { icon: <HatchedIcon />, label: "BAD" },
  comm: { icon: <RingBrokenIcon />, label: "COMM" },
  offline: { icon: <SquareDimmedIcon />, label: "OFFLINE" },
  ack: { icon: <CheckCircleIcon />, label: "ACK" }
};

const joinClassNames = (...classNames: Array<string | undefined | false>) =>
  classNames.filter(Boolean).join(" ");

export type AlarmBadgeProps = {
  state: AlarmBadgeState;
  label?: string;
  blink?: boolean;
};

export const AlarmBadge = ({ state, label, blink = false }: AlarmBadgeProps) => {
  const config = configByState[state];
  const shouldBlink = blink && (state === "crit" || state === "warn");

  return (
    <Badge
      className={joinClassNames(styles.root, styles[state], shouldBlink && styles.blink)}
      variant="light"
    >
      <span className={styles.icon} aria-hidden>
        {config.icon}
      </span>
      {label ?? config.label}
    </Badge>
  );
};

