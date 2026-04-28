import {
  MantineProvider,
  localStorageColorSchemeManager,
  type MantineColorScheme
} from "@mantine/core";
import { Notifications } from "@mantine/notifications";
import type { PropsWithChildren } from "react";

import {
  COLOR_SCHEME_STORAGE_KEY,
  defaultColorScheme,
  mantineTheme
} from "@/shared/libs/theme";

const colorSchemeManager = localStorageColorSchemeManager({
  key: COLOR_SCHEME_STORAGE_KEY
});

export const ThemeProvider = ({ children }: PropsWithChildren) => {
  return (
    <MantineProvider
      defaultColorScheme={defaultColorScheme as MantineColorScheme}
      colorSchemeManager={colorSchemeManager}
      theme={mantineTheme}
    >
      <Notifications />
      {children}
    </MantineProvider>
  );
};
