import type { ReactNode } from "react";

import { RequireAtLeastOne } from "@shared/libs/types/RequireAtLeastOne";

interface NavbarBase {
  top: ReactNode;
  center: ReactNode;
  bottom: ReactNode;
}

export type NavbarProps = RequireAtLeastOne<
  Partial<NavbarBase>,
  keyof NavbarBase
>;

