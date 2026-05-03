import { useNavigate } from "react-router";

import { routePaths } from "@shared/libs/router";
import { Button, Card, Group, Stack, Text } from "@shared/ui";

export default function SessionExpiredPage() {
  const navigate = useNavigate();

  return (
    <Group justify="center" align="center" style={{ minHeight: "100vh", padding: 16 }}>
      <Card withBorder p="lg" w="100%" maw={420}>
        <Stack gap="md">
          <Text size="xl" fw={700}>Сессия истекла</Text>
          <Text c="dimmed">Для продолжения работы войдите ещё раз.</Text>
          <Group justify="flex-end">
            <Button onClick={() => void navigate(routePaths.login, { replace: true })}>
              Войти ещё раз
            </Button>
          </Group>
        </Stack>
      </Card>
    </Group>
  );
}