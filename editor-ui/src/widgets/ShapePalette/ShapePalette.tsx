import {
  IconCircle,
  IconLine,
  IconPhoto,
  IconSquare,
  IconTypography
} from "@tabler/icons-react";

import { shapes } from "@entities/shapes";
import { useEditorStore } from "@features/editor";
import { ActionIcon, Stack, Text, Tooltip } from "@shared/ui";

import styles from "./ShapePalette.module.css";

const MVP_SHAPES = ["rect", "circle", "line", "text", "image"] as const;

type PaletteShapeType = (typeof MVP_SHAPES)[number];

const ICON_BY_TYPE: Record<PaletteShapeType, React.ComponentType<{ size?: number }>> = {
  rect: IconSquare,
  circle: IconCircle,
  line: IconLine,
  text: IconTypography,
  image: IconPhoto
};

const SHORTCUT_BY_TYPE: Record<PaletteShapeType, string> = {
  rect: "R",
  circle: "C",
  line: "L",
  text: "T",
  image: "I"
};

export const ShapePalette = () => {
  const { activeTool, setActiveTool } = useEditorStore();

  return (
    <Stack gap="sm" p="sm">
      <Text fw={600} size="sm">
        Фигуры
      </Text>
      <div className={styles.paletteGrid}>
        {shapes
          .filter((shape): shape is { type: PaletteShapeType; label: string } =>
            MVP_SHAPES.includes(shape.type as PaletteShapeType)
          )
          .map((shape) => {
            const Icon = ICON_BY_TYPE[shape.type];
            const isActive = activeTool === shape.type;

            return (
              <div key={shape.type} className={styles.paletteItem}>
                <Tooltip label={shape.label}>
                  <ActionIcon
                    variant={isActive ? "filled" : "subtle"}
                    onClick={() => setActiveTool(shape.type)}
                    aria-label={shape.label}
                  >
                    <Icon size={18} />
                  </ActionIcon>
                </Tooltip>
                <span className={styles.shortcut}>{SHORTCUT_BY_TYPE[shape.type]}</span>
              </div>
            );
          })}
      </div>
    </Stack>
  );
};
