import { useNavigate } from "react-router";

import { useAuthStore } from "@features/auth";
import { routePaths } from "@shared/libs/router";
import { Button, Card, Group, Stack, Text } from "@shared/ui";

export default function AccessDeniedPage() {
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);

  const viewerURL = (import.meta.env.VITE_VIEWER_UI_URL ?? "").trim();

  return (
    <Group justify="center" align="center" style={{ minHeight: "100vh", padding: 16 }}>
      <Card withBorder p="lg" w="100%" maw={460}>
        <Stack gap="md">
          <Text size="xl" fw={700}>Доступ ограничен</Text>
          <Text>
            Эта страница доступна инженерам и администраторам. Для оператора — viewer-ui.
          </Text>
          <Group justify="flex-end" gap="xs">
            {viewerURL ? (
              <Button
                variant="secondary"
                onClick={() => {
                  window.location.assign(viewerURL);
                }}
              >
                Перейти в viewer-ui
              </Button>
            ) : null}
            <Button
              variant="ghost"
              onClick={async () => {
                await logout();
                void navigate(routePaths.login, { replace: true });
              }}
            >
              Выйти
            </Button>
          </Group>
        </Stack>
      </Card>
    </Group>
  );
}