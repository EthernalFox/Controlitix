import { useState } from "react";
import { useNavigate } from "react-router";

import { useAuthStore } from "@/features/auth";
import { APP_PATHS } from "@/shared/libs/router";
import { Button } from "@/shared/ui/components";

export const UserMenu = () => {
  const navigate = useNavigate();

  const user = useAuthStore((state) => state.user);
  const logout = useAuthStore((state) => state.logout);

  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const onLogout = async () => {
    setIsLoggingOut(true);

    try {
      await logout();
    } finally {
      void navigate(APP_PATHS.LOGIN, { replace: true });
      setIsLoggingOut(false);
    }
  };

  return (
    <Button variant="ghost" onClick={onLogout} loading={isLoggingOut}>
      {user?.username ?? "Пользователь"} · Выйти
    </Button>
  );
};
