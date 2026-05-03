import { ReactNode } from "react";

import { RequireAtLeastOne } from "@shared/libs/types/RequireAtLeastOne";

interface PanelBase {
  top: ReactNode;
  center: ReactNode;
  bottom: ReactNode;
}

export type PanelProps = RequireAtLeastOne<Partial<PanelBase>, keyof PanelBase> & {
  enabled?: boolean;
};
