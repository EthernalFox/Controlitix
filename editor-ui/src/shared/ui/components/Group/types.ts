import type { GroupProps as MantineGroupProps } from "@mantine/core";
import type { ReactNode } from "react";

export interface GroupProps extends MantineGroupProps {
  children?: ReactNode;
}
