import type { TabsProps as MantineTabsProps } from "@mantine/core";

export type TabsVariant = "pill" | "panel";

export type TabsProps = Omit<MantineTabsProps, "variant"> & {
  variant?: TabsVariant;
};

