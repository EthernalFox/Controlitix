import { Text } from "@shared/ui";

export const DrawCanvas = () => {
  return (
    <div
      style={{
        height: "100%",
        minHeight: 360,
        borderRadius: 16,
        border: "1px dashed rgba(0, 0, 0, 0.18)",
        background: "rgba(0, 0, 0, 0.03)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center"
      }}
    >
      <Text c="dimmed">Canvas placeholder</Text>
    </div>
  );
};
