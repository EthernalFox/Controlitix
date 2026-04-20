import { IconChevronDown, IconMinus, IconPlus } from "@tabler/icons-react";
import { useMemo } from "react";


import { FRAME_PRESETS, useEditorStore } from "@features/editor";
import { ActionIcon, Button, Footer, Group, Popover, Slider, Stack, Text } from "@shared/ui";

interface EditorFooterProps {
  diagramId: string;
}

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

const saveStatusMeta: Record<
  "saved" | "saving" | "error",
  { color: string; label: string }
> = {
  saved: { color: "#2F9E44", label: "Сохранено" },
  saving: { color: "#FAB005", label: "Сохранение..." },
  error: { color: "#E03131", label: "Ошибка сохранения" }
};

export const EditorFooter = ({ diagramId }: EditorFooterProps) => {
  const {
    cursorX,
    cursorY,
    frame,
    saveStatus,
    setFrame,
    setZoom,
    zoom
  } = useEditorStore();

  const zoomPercent = useMemo(() => Math.round(zoom * 100), [zoom]);

  const setZoomPercent = (value: number) => {
    setZoom(clamp(value / 100, 0.1, 5));
  };

  const onFrameSelect = (label: string) => {
    const nextPreset = FRAME_PRESETS.find((preset) => preset.label === label);
    if (!nextPreset) {
      return;
    }

    setFrame(nextPreset);
    localStorage.setItem(`editor:frame:${diagramId}`, JSON.stringify(nextPreset));
  };

  return (
    <Footer
      before={
        <Group gap="xs" wrap="nowrap" style={{ minWidth: 280 }}>
          <ActionIcon
            variant="subtle"
            onClick={() => setZoomPercent(zoomPercent - 10)}
            aria-label="Zoom out"
          >
            <IconMinus size={14} />
          </ActionIcon>
          <Slider
            min={10}
            max={500}
            value={zoomPercent}
            onChange={setZoomPercent}
            style={{ width: 180 }}
          />
          <ActionIcon
            variant="subtle"
            onClick={() => setZoomPercent(zoomPercent + 10)}
            aria-label="Zoom in"
          >
            <IconPlus size={14} />
          </ActionIcon>
          <Text size="sm" style={{ minWidth: 52 }}>
            {zoomPercent}%
          </Text>
        </Group>
      }
      main={
        <Popover width={180} position="top" shadow="md">
          <Popover.Target>
            <Button variant="secondary" rightSection={<IconChevronDown size={14} />}>
              Frame: {frame.label}
            </Button>
          </Popover.Target>
          <Popover.Dropdown>
            <Stack gap={6}>
              {FRAME_PRESETS.map((preset) => (
                <Button
                  key={preset.label}
                  variant={preset.label === frame.label ? "primary" : "secondary"}
                  onClick={() => onFrameSelect(preset.label)}
                >
                  {preset.label} ({preset.width} × {preset.height})
                </Button>
              ))}
            </Stack>
          </Popover.Dropdown>
        </Popover>
      }
      after={
        <Group gap="md" wrap="nowrap">
          <Group gap={6} wrap="nowrap">
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                display: "inline-block",
                background: saveStatusMeta[saveStatus].color
              }}
            />
            <Text size="sm">{saveStatusMeta[saveStatus].label}</Text>
          </Group>
          <Text size="sm" c="dimmed">
            X: {Math.round(cursorX)} Y: {Math.round(cursorY)}
          </Text>
        </Group>
      }
    />
  );
};
