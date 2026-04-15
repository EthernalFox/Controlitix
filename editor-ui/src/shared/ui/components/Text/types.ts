import type { TextProps as MantineTextProps } from "@mantine/core";
import type { ReactNode } from "react";

export interface TextProps extends MantineTextProps {
  children?: ReactNode;
}
