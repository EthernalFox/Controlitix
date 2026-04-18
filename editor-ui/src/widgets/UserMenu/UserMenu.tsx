import { useMantineColorScheme } from "@mantine/core";

import {
  ActionIcon,
  Avatar,
  Menu,
  Text
} from "@shared/ui";

export const UserMenu = () => {
  const { colorScheme, setColorScheme } = useMantineColorScheme();
  const resolvedColorScheme =
    colorScheme === "dark" || colorScheme === "light" ? colorScheme : "light";

  const onToggleTheme = () => {
    setColorScheme(resolvedColorScheme === "light" ? "dark" : "light");
  };

  return (
    <Menu shadow="md" width={220} position="bottom-end">
      <Menu.Target>
        <ActionIcon variant="subtle" color="gray" size="lg" aria-label="Open menu">
          <Avatar radius="xl" size="sm">
            CT
          </Avatar>
        </ActionIcon>
      </Menu.Target>

      <Menu.Dropdown>
        <Menu.Label>Профиль</Menu.Label>
        <Menu.Item onClick={onToggleTheme}>
          <Text size="sm">
            Тема: {resolvedColorScheme === "light" ? "Светлая" : "Тёмная"}
          </Text>
        </Menu.Item>
        <Menu.Item color="red">Выйти</Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
};
