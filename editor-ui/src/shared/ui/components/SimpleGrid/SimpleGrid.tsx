import { SimpleGrid as MantineSimpleGrid } from "@mantine/core";

import type { SimpleGridProps } from "./types";

export const SimpleGrid = (props: SimpleGridProps) => {
  return <MantineSimpleGrid {...props} />;
};

