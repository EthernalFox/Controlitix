import { Navigate } from "react-router";

import { hasEditorRole, useAuthStore } from "@features/auth";
import { routePaths } from "@shared/libs/router";
import { Card, Stack, Text } from "@shared/ui";

import { LoginForm } from "./LoginForm";
import styles from "./LoginPage.module.css";

export default function LoginPage() {
  const status = useAuthStore((state) => state.status);
  const user = useAuthStore((state) => state.user);

  if (status === "authenticated") {
    return (
      <Navigate
        to={hasEditorRole(user) ? routePaths.objects : routePaths.accessDenied}
        replace
      />
    );
  }

  return (
    <main className={styles.page}>
      <Card withBorder shadow="md" p="lg" radius="md" className={styles.card}>
        <Stack gap="md">
          <Stack gap={2}>
            <Text size="xl" fw={700}>Вход</Text>
            <Text c="dimmed">Controlitix Editor</Text>
          </Stack>
          <LoginForm />
        </Stack>
      </Card>
    </main>
  );
}