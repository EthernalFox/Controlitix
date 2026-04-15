import { AppShell } from "@shared/ui/components";

import type { PanelProps } from "./types";

export const Panel = ({ top, center, bottom }: PanelProps) => {
  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      {top && <AppShell.Section>{top}</AppShell.Section>}
      {center && <AppShell.Section grow>{center}</AppShell.Section>}
      {bottom && <AppShell.Section>{bottom}</AppShell.Section>}
    </div>
  );
};

