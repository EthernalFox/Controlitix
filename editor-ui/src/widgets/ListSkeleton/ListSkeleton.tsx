import { SimpleGrid } from "@shared/ui";

import styles from "./ListSkeleton.module.css";

interface ListSkeletonProps {
  cards?: number;
  cols?: number;
}

export const ListSkeleton = ({ cards = 6, cols = 3 }: ListSkeletonProps) => {
  return (
    <SimpleGrid cols={cols} spacing="md" className={styles.grid}>
      {Array.from({ length: cards }).map((_, index) => (
        <div key={index} className={styles.skeletonCard} />
      ))}
    </SimpleGrid>
  );
};
