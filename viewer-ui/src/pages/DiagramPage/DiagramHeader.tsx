import { Group, Text, Title, Button } from "@/shared/ui/components";

import styles from "./DiagramPage.module.css";

interface DiagramHeaderProps {
  objectName: string;
  diagramName: string;
  scale: number;
  onBack: () => void;
  onFit: () => void;
  onReset: () => void;
  onZoomIn: () => void;
  onZoomOut: () => void;
}

export const DiagramHeader = ({
  objectName,
  diagramName,
  scale,
  onBack,
  onFit,
  onReset,
  onZoomIn,
  onZoomOut
}: DiagramHeaderProps) => {
  return (
    <div className={styles.header}>
      <Group justify="space-between" align="center" wrap="wrap" gap="sm">
        <div>
          <Title order={3}>{diagramName}</Title>
          <Text size="sm" c="dimmed">
            {objectName}
          </Text>
        </div>

        <Group gap="xs" wrap="wrap">
          <Button variant="secondary" onClick={onBack}>
            Назад
          </Button>
          <Button variant="secondary" onClick={onFit}>
            Вписать
          </Button>
          <Button variant="secondary" onClick={onReset}>
            100%
          </Button>
          <Button variant="secondary" onClick={onZoomOut}>
            -
          </Button>
          <Button variant="secondary" onClick={onZoomIn}>
            +
          </Button>
          <Text size="sm" c="dimmed" className={styles.scaleLabel}>
            {Math.round(scale * 100)}%
          </Text>
        </Group>
      </Group>
    </div>
  );
};
