import { AppShell } from "@shared/ui/components";

import type { AsideProps } from "./types";

export const Aside = ({ top, center, bottom }: AsideProps) => {
  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      {top && <AppShell.Section>{top}</AppShell.Section>}
      {center && <AppShell.Section grow>{center}</AppShell.Section>}
      {bottom && <AppShell.Section>{bottom}</AppShell.Section>}
    </div>
  );
};

