import { useMemo } from "react";
import { Link, useLocation, useParams } from "react-router";

import { APP_PATHS } from "@/shared/libs/router";
import { Divider, Group, Stack, Text, Title } from "@/shared/ui/components";
import { AlarmTicker } from "@/widgets/AlarmTicker";
import { RealtimeStatus } from "@/widgets/RealtimeStatus";
import { ThemeSelect } from "@/widgets/ThemeSelect";
import { UserMenu } from "@/widgets/UserMenu";

const Logo = () => {
  return (
    <Text fw={700} size="sm">
      Controlitix Viewer
    </Text>
  );
};

export const AppHeader = () => {
  const location = useLocation();
  const { objectId, diagramId } = useParams();

  const breadcrumbs = useMemo(() => {
    if (objectId && diagramId) {
      return `Object ${objectId} / Diagram ${diagramId}`;
    }

    if (location.pathname === APP_PATHS.ALARMS) {
      return "Alarms";
    }

    if (location.pathname.startsWith("/trends")) {
      return "Trends";
    }

    return "Dashboard";
  }, [diagramId, location.pathname, objectId]);

  return (
    <Group justify="space-between" align="center" wrap="nowrap" h="100%" gap="sm">
      <Group gap="sm" wrap="nowrap" style={{ minWidth: 0 }}>
        <Logo />
        <Divider orientation="vertical" />
        <Stack gap={0} style={{ minWidth: 0 }}>
          <Title order={6} c="dimmed">
            <Link
              to={APP_PATHS.DASHBOARD}
              style={{ color: "inherit", textDecoration: "none" }}
            >
              {breadcrumbs}
            </Link>
          </Title>
        </Stack>
      </Group>

      <Group gap="xs" wrap="nowrap">
        <RealtimeStatus />
        <AlarmTicker />
        <ThemeSelect />
        <UserMenu />
      </Group>
    </Group>
  );
};
