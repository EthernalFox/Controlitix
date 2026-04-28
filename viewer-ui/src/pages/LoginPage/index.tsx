import { Navigate } from "react-router";

import { useAuthStore } from "@/features/auth";
import { APP_PATHS } from "@/shared/libs/router";
import { Card, Stack, Text, Title } from "@/shared/ui/components";

import { LoginForm } from "./LoginForm";
import styles from "./LoginPage.module.css";

export const LoginPage = () => {
  const status = useAuthStore((state) => state.status);

  if (status === "authenticated") {
    return <Navigate to={APP_PATHS.DASHBOARD} replace />;
  }

  return (
    <div className={styles.page}>
      <Card withBorder shadow="md" p="lg" radius="md" className={styles.card}>
        <Stack gap="md">
          <Stack gap={2}>
            <Title order={2}>Вход</Title>
            <Text c="dimmed">Controlitix Viewer</Text>
          </Stack>
          <LoginForm />
        </Stack>
      </Card>
    </div>
  );
};
