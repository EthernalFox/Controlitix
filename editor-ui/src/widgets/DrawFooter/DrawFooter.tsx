import { shapes } from "@entities/shapes";
import { Button, Footer, Group, Text } from "@shared/ui";

export const DrawFooter = () => {
  return (
    <Footer
      main={
        <div
          style={{
            width: "100%",
            display: "flex",
            alignItems: "center",
            gap: 12
          }}
        >
          <Text size="sm" fw={600}>
            Shapes
          </Text>
          <div style={{ flex: "1 1 auto", overflowX: "auto" }}>
            <Group gap={8} wrap="nowrap">
              {shapes.map((shape) => (
                <Button key={shape.type} variant="ghost" size="xs">
                  {shape.label}
                </Button>
              ))}
            </Group>
          </div>
        </div>
      }
    />
  );
};
