import { ReactNode } from "react";

import { RequireAtLeastOne } from "@shared/libs/types/RequireAtLeastOne";

interface FooterBase {
  before: ReactNode;
  main: ReactNode;
  after: ReactNode;
}

export type FooterProps = RequireAtLeastOne<
  Partial<FooterBase>,
  keyof FooterBase
>;
