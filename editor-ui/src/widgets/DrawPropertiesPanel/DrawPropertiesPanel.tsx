import { Aside, Card, Group, Stack, Text, TextInput } from "@shared/ui";

export const DrawPropertiesPanel = () => {
  return (
    <Aside
      top={
        <div style={{ padding: 12 }}>
          <Text fw={600}>Properties</Text>
          <Text size="xs" c="dimmed">
            Select an element to edit its params.
          </Text>
        </div>
      }
      center={
        <div style={{ padding: 12 }}>
          <Stack gap={12}>
            <Card withBorder p="md">
              <Text fw={600} mb={8}>
                Position
              </Text>
              <Group grow gap="sm">
                <TextInput label="X" placeholder="0" size="xs" />
                <TextInput label="Y" placeholder="0" size="xs" />
              </Group>
            </Card>
            <Card withBorder p="md">
              <Text fw={600} mb={8}>
                Size
              </Text>
              <Group grow gap="sm">
                <TextInput label="W" placeholder="0" size="xs" />
                <TextInput label="H" placeholder="0" size="xs" />
              </Group>
            </Card>
            <Card withBorder p="md">
              <Text fw={600} mb={8}>
                Appearance
              </Text>
              <TextInput label="Fill" placeholder="#000000" size="xs" />
              <TextInput
                label="Stroke"
                placeholder="#FFFFFF"
                size="xs"
                mt={8}
              />
            </Card>
          </Stack>
        </div>
      }
    />
  );
};
