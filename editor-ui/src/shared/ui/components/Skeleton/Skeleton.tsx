import { Skeleton as MantineSkeleton } from "@mantine/core";

import type { SkeletonProps } from "./types";

export const Skeleton = (props: SkeletonProps) => {
  return <MantineSkeleton {...props} />;
};
