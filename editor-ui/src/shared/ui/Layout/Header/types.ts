import { ReactNode } from "react";

import { RequireAtLeastOne } from "@shared/libs/types/RequireAtLeastOne";

interface HeaderBase {
  before: ReactNode;
  main: ReactNode;
  after: ReactNode;
}

export type HeaderProps = RequireAtLeastOne<
  Partial<HeaderBase>,
  keyof HeaderBase
>;
