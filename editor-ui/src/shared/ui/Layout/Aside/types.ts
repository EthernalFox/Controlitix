import type { ReactNode } from "react";

import { RequireAtLeastOne } from "@shared/libs/types/RequireAtLeastOne";

interface AsideBase {
  top: ReactNode;
  center: ReactNode;
  bottom: ReactNode;
}

export type AsideProps = RequireAtLeastOne<Partial<AsideBase>, keyof AsideBase> & {
  enabled?: boolean;
};
