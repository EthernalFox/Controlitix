import { AppShell } from "@shared/ui/components";

import type { PanelProps } from "./types";

export const Panel = ({ top, center, bottom, enabled = true }: PanelProps) => {
  if (!enabled) {
    return null;
  }

  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column", padding: "var(--mantine-spacing-sm)" }}>
      {top && <AppShell.Section>{top}</AppShell.Section>}
      {center && <AppShell.Section grow>{center}</AppShell.Section>}
      {bottom && <AppShell.Section>{bottom}</AppShell.Section>}
    </div>
  );
};
