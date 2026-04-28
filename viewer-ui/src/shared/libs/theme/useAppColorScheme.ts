import {
  useMantineColorScheme,
  type MantineColorScheme
} from "@mantine/core";

export const useAppColorScheme = () => {
  const { colorScheme, setColorScheme } = useMantineColorScheme();

  return {
    colorScheme,
    setColorScheme: (nextValue: MantineColorScheme) => {
      setColorScheme(nextValue);
    }
  };
};
