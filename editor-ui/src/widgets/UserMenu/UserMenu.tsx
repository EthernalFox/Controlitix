import { useNavigate } from "react-router";

import { useAuthStore } from "@features/auth";
import { routePaths } from "@shared/libs/router";
import { ActionIcon, Avatar, Menu, Text } from "@shared/ui";

export const UserMenu = () => {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);

  const onLogout = async () => {
    await logout();
    void navigate(routePaths.login, { replace: true });
  };

  return (
    <Menu shadow="md" width={240} position="bottom-end">
      <Menu.Target>
        <ActionIcon variant="subtle" color="gray" size="lg" aria-label="Открыть меню пользователя">
          <Avatar radius="xl" size="sm">
            {user?.username?.slice(0, 2).toUpperCase() || "CT"}
          </Avatar>
        </ActionIcon>
      </Menu.Target>

      <Menu.Dropdown>
        <Menu.Label>Профиль</Menu.Label>
        <Menu.Item>
          <Text size="sm">{user?.username || "Пользователь"}</Text>
        </Menu.Item>
        <Menu.Item color="red" onClick={() => void onLogout()}>
          Выйти
        </Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
};
