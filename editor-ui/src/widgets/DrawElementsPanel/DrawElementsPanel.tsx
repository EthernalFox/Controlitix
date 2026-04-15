import { Button, Card, Panel, Stack, Text, TextInput } from "@shared/ui";

const ELEMENTS = [
  { id: "layer-1", label: "Main group" },
  { id: "layer-2", label: "Background" },
  { id: "layer-3", label: "Indicators" },
  { id: "layer-4", label: "Labels" }
];

export const DrawElementsPanel = () => {
  return (
    <Panel
      top={
        <div style={{ padding: 12 }}>
          <Text fw={600}>Elements</Text>
          <TextInput placeholder="Search" size="xs" mt={8} />
        </div>
      }
      center={
        <div style={{ padding: 12 }}>
          <Stack gap={8}>
            {ELEMENTS.map((item) => (
              <Card key={item.id} withBorder p="sm">
                <Text size="sm" fw={500}>
                  {item.label}
                </Text>
              </Card>
            ))}
          </Stack>
        </div>
      }
      bottom={
        <div style={{ padding: 12 }}>
          <Button variant="secondary" size="xs" fullWidth>
            Add group
          </Button>
        </div>
      }
    />
  );
};
