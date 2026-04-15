import { Group as MantineGroup } from "@mantine/core";

import type { GroupProps } from "./types";

export const Group = (props: GroupProps) => {
  return <MantineGroup {...props} />;
};

