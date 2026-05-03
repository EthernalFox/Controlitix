import type { CardProps as MantineCardProps } from "@mantine/core";

export type CardVariant = "flat" | "glass";

export type CardProps = Omit<MantineCardProps, "variant"> & {
  variant?: CardVariant;
};

