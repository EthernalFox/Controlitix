import {
  IconCircle,
  IconLine,
  IconPhoto,
  IconSquare,
  IconTypography
} from "@tabler/icons-react";

import { shapes } from "@entities/shapes";
import { useEditorStore } from "@features/editor";
import { ActionIcon, Group, Stack, Text, Tooltip } from "@shared/ui";

const MVP_SHAPES = ["rect", "circle", "line", "text", "image"] as const;

type PaletteShapeType = (typeof MVP_SHAPES)[number];

const ICON_BY_TYPE: Record<PaletteShapeType, React.ComponentType<{ size?: number }>> = {
  rect: IconSquare,
  circle: IconCircle,
  line: IconLine,
  text: IconTypography,
  image: IconPhoto
};

export const ShapePalette = () => {
  const { activeTool, setActiveTool } = useEditorStore();

  return (
    <Stack gap="sm" p="sm">
      <Text fw={600} size="sm">
        Фигуры
      </Text>
      <Group gap="xs" wrap="wrap">
        {shapes
          .filter((shape): shape is { type: PaletteShapeType; label: string } =>
            MVP_SHAPES.includes(shape.type as PaletteShapeType)
          )
          .map((shape) => {
            const Icon = ICON_BY_TYPE[shape.type];
            const isActive = activeTool === shape.type;

            return (
              <Tooltip key={shape.type} label={shape.label}>
                <ActionIcon
                  variant={isActive ? "filled" : "subtle"}
                  onClick={() => setActiveTool(shape.type)}
                  aria-label={shape.label}
                >
                  <Icon size={18} />
                </ActionIcon>
              </Tooltip>
            );
          })}
      </Group>
    </Stack>
  );
};
