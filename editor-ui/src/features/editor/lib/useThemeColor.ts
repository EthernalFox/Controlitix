import { useEffect, useState } from "react";

const readColor = (varName: string) =>
  getComputedStyle(document.documentElement).getPropertyValue(varName).trim() || "#228BE6";

export const useThemeColor = (varName: string) => {
  const [color, setColor] = useState(() => readColor(varName));

  useEffect(() => {
    const updateColor = () => {
      setColor(readColor(varName));
    };

    updateColor();

    const observer = new MutationObserver(updateColor);
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["data-mantine-color-scheme"]
    });

    return () => {
      observer.disconnect();
    };
  }, [varName]);

  return color;
};
