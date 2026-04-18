import { Avatar as MantineAvatar } from "@mantine/core";

import type { AvatarProps } from "./types";

export const Avatar = (props: AvatarProps) => {
  return <MantineAvatar {...props} />;
};
