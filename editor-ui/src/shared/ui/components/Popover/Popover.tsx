import { Popover as MantinePopover } from "@mantine/core";

import type { PopoverProps } from "./types";
import modalStyles from "../Modal/Modal.module.css";

const PopoverRoot = (props: PopoverProps) => {
  return (
    <MantinePopover
      {...props}
      classNames={{
        dropdown: modalStyles.dropdown
      }}
    />
  );
};

export const Popover = Object.assign(PopoverRoot, {
  Target: MantinePopover.Target,
  Dropdown: MantinePopover.Dropdown
});
