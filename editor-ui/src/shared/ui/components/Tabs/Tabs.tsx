import { Tabs as MantineTabs } from "@mantine/core";

import styles from "./Tabs.module.css";
import type { TabsProps } from "./types";

const TabsRoot = ({ variant = "pill", ...props }: TabsProps) => {
  const resolvedClassNames =
    variant === "pill"
      ? {
          list: styles.tabsList,
          tab: styles.tab
        }
      : undefined;

  return <MantineTabs {...props} classNames={resolvedClassNames} />;
};

export const Tabs = Object.assign(TabsRoot, {
  List: MantineTabs.List,
  Tab: MantineTabs.Tab,
  Panel: MantineTabs.Panel
});
