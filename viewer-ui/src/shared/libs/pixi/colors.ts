const normalizeHex = (hex: string): string => {
  const value = hex.trim();
  if (!value) {
    return "#000000";
  }

  if (value.startsWith("#")) {
    if (value.length === 4) {
      const r = value[1];
      const g = value[2];
      const b = value[3];
      return `#${r}${r}${g}${g}${b}${b}`;
    }

    return value;
  }

  return `#${value}`;
};

export const hexToPixiColor = (hex: string): number => {
  const normalized = normalizeHex(hex);
  const parsed = Number.parseInt(normalized.slice(1), 16);

  if (Number.isNaN(parsed)) {
    return 0x000000;
  }

  return parsed;
};
