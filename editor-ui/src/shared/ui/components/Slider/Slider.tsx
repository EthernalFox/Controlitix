import { Slider as MantineSlider } from "@mantine/core";

import type { SliderProps } from "./types";

export const Slider = (props: SliderProps) => {
  return <MantineSlider {...props} />;
};
