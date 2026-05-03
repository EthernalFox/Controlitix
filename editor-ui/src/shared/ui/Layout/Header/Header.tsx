import type { CSSProperties } from "react";

import type { HeaderProps } from "./types";

const slotStyle: CSSProperties = {
  display: "flex",
  alignItems: "center",
  minWidth: 0
};

export const Header = ({ before, main, after }: HeaderProps) => {
  return (
    <div
      style={{
        height: "100%",
        display: "flex",
        alignItems: "center",
        gap: 12,
        paddingInline: "var(--mantine-spacing-sm)"
      }}
    >
      {before && <div style={{ ...slotStyle, flex: "0 0 auto" }}>{before}</div>}
      {main && <div style={{ ...slotStyle, flex: "1 1 auto" }}>{main}</div>}
      {after && <div style={{ ...slotStyle, flex: "0 0 auto", gap: 8 }}>{after}</div>}
    </div>
  );
};
