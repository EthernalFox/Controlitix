import { Button, useMantineColorScheme } from "@mantine/core";
import type { CSSProperties } from "react";

import { defaultColorScheme } from "@shared/libs/theme";

import type { HeaderProps } from "./types";

const slotStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  minWidth: 0
};

const ThemeToggle = () => {
  const { colorScheme, setColorScheme } = useMantineColorScheme();
  const resolvedScheme =
    colorScheme === "light" || colorScheme === "dark"
      ? colorScheme
      : defaultColorScheme;
  const nextScheme = resolvedScheme === "light" ? "dark" : "light";
  const label = resolvedScheme === "light" ? "Светлая" : "Тёмная";

  return (
    <Button
      size="xs"
      variant="subtle"
      color="primary"
      onClick={() => setColorScheme(nextScheme)}
      aria-label="Переключить тему"
    >
      Тема: {label}
    </Button>
  );
};

export const Header = ({ before, main, after }: HeaderProps) => {
  return (
    <div
      style={{
        height: "100%",
        display: "flex",
        alignItems: "center",
        gap: 12,
        paddingInline: 12
      }}
    >
      {before && <div style={{ ...slotStyle, flex: "0 0 auto" }}>{before}</div>}
      {main && <div style={{ ...slotStyle, flex: "1 1 auto" }}>{main}</div>}
      <div style={{ ...slotStyle, flex: "0 0 auto", gap: 8 }}>
        {after}
        <ThemeToggle />
      </div>
    </div>
  );
};
