import { Table as MantineTable } from "@mantine/core";

import type { TableProps } from "./types";

export const Table = (props: TableProps) => {
  return <MantineTable {...props} />;
};
