import { Menu as MantineMenu } from "@mantine/core";

import type { MenuProps } from "./types";
import modalStyles from "../Modal/Modal.module.css";

const MenuRoot = ({ classNames, ...props }: MenuProps) => {
  const resolvedClassNames =
    typeof classNames === "function"
      ? classNames
      : {
          dropdown: classNames?.dropdown ?? modalStyles.dropdown,
          item: classNames?.item ?? modalStyles.item,
          label: classNames?.label ?? modalStyles.label,
          ...classNames
        };

  return <MantineMenu {...props} classNames={resolvedClassNames} />;
};

export const Menu = Object.assign(MenuRoot, {
  Target: MantineMenu.Target,
  Dropdown: MantineMenu.Dropdown,
  Item: MantineMenu.Item,
  Label: MantineMenu.Label,
  Divider: MantineMenu.Divider
});
