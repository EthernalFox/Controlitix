import type { ObjectAlarmCounts } from "@/features/objects-summary/lib/aggregateAlarms";
import type { ObjectSummary } from "@/shared/api/objects";
import { Card, Skeleton, Stack } from "@/shared/ui/components";
import { ObjectCard } from "@/widgets/ObjectCard/ObjectCard";

import styles from "./ObjectGrid.module.css";

interface ObjectGridProps {
  objects: ObjectSummary[];
  getCounts: (objectId: string) => ObjectAlarmCounts;
  onOpen: (path: string) => void;
}

export const ObjectGrid = ({ objects, getCounts, onOpen }: ObjectGridProps) => {
  return (
    <div className={styles.grid}>
      {objects.map((item) => (
        <ObjectCard key={item.id} object={item} counts={getCounts(item.id)} onOpen={onOpen} />
      ))}
    </div>
  );
};

export const ObjectGridSkeleton = () => {
  return (
    <div className={styles.grid}>
      {Array.from({ length: 6 }).map((_, index) => (
        <Card key={index} withBorder p="md">
          <Stack gap="sm">
            <Skeleton height={20} radius="sm" />
            <Skeleton height={14} radius="sm" width="80%" />
            <Skeleton height={24} radius="sm" width="60%" />
          </Stack>
        </Card>
      ))}
    </div>
  );
};
