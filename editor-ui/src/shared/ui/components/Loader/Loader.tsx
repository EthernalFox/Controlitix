import { Loader as MantineLoader } from "@mantine/core";

import type { LoaderProps } from "./types";

export const Loader = (props: LoaderProps) => {
  return <MantineLoader {...props} />;
};
