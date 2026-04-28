import { useMemo } from "react";

import { useAppColorScheme } from "@/shared/libs/theme";
import { Select } from "@/shared/ui/components";

const options = [
  { value: "dark", label: "Dark" },
  { value: "light", label: "Light" },
  { value: "auto", label: "System" }
] as const;

export const ThemeSelect = () => {
  const { colorScheme, setColorScheme } = useAppColorScheme();

  const value = useMemo(() => {
    if (colorScheme === "light" || colorScheme === "dark") {
      return colorScheme;
    }

    return "auto";
  }, [colorScheme]);

  return (
    <Select
      aria-label="Color scheme"
      size="xs"
      w={110}
      data={options.map((option) => ({ value: option.value, label: option.label }))}
      value={value}
      onChange={(nextValue) => {
        if (!nextValue) {
          return;
        }

        if (nextValue === "auto") {
          setColorScheme("auto");
          return;
        }

        if (nextValue === "light" || nextValue === "dark") {
          setColorScheme(nextValue);
        }
      }}
      allowDeselect={false}
    />
  );
};
