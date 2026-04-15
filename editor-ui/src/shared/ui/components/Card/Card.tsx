import { Card as MantineCard } from "@mantine/core";

import type { CardProps } from "./types";

export const Card = (props: CardProps) => {
  return <MantineCard {...props} />;
};
