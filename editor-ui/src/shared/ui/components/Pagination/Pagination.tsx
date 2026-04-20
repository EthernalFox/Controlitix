import { Pagination as MantinePagination } from "@mantine/core";

import type { PaginationProps } from "./types";

export const Pagination = (props: PaginationProps) => {
  return <MantinePagination {...props} />;
};
