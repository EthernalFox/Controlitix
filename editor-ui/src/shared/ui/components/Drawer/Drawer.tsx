import { Drawer as MantineDrawer } from "@mantine/core";

import type { DrawerProps } from "./types";

export const Drawer = (props: DrawerProps) => {
  return <MantineDrawer {...props} />;
};

