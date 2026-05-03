import { useDisclosure } from "@mantine/hooks";
import { createElement } from "react";
import {
  Navigate,
  Outlet,
  RouterProvider as ReactRouterProvider,
  createBrowserRouter,
  useNavigate
} from "react-router";

import { AlarmsPage } from "@/pages/AlarmsPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { DiagramPage } from "@/pages/DiagramPage";
import { LoginPage } from "@/pages/LoginPage";
import { TrendsPage } from "@/pages/TrendsPage";
import { APP_PATHS, ProtectedRoute } from "@/shared/libs/router";
import { AppShell, Burger, Button, Group, Stack, Text, Title } from "@/shared/ui/components";
import { AppHeader } from "@/widgets/AppHeader";
import { AppNavbar } from "@/widgets/AppNavbar";

const MinimalLayout = ({ title }: { title: string }) => {
  return (
    <Group justify="center" align="center" style={{ minHeight: "100vh" }}>
      <Stack gap="xs" align="center">
        <Title order={2}>{title}</Title>
        <Text c="dimmed">Coming soon</Text>
      </Stack>
    </Group>
  );
};

const SessionExpiredLayout = () => {
  const navigate = useNavigate();

  return (
    <Group justify="center" align="center" style={{ minHeight: "100vh" }}>
      <Stack gap="sm" align="center">
        <Title order={2}>Сессия истекла</Title>
        <Text c="dimmed">Для продолжения войдите снова.</Text>
        <Button onClick={() => void navigate(APP_PATHS.LOGIN, { replace: true })}>
          Войти ещё раз
        </Button>
      </Stack>
    </Group>
  );
};

const AppShellLayout = () => {
  const [opened, { toggle }] = useDisclosure(true);

  return (
    <AppShell
      header={{ height: { base: 56, sm: 64 } }}
      navbar={{
        width: { base: 240, md: 280 },
        breakpoint: "sm",
        collapsed: { mobile: !opened, desktop: !opened }
      }}
      padding="md"
    >
      <AppShell.Header>
        <Group h="100%" px="sm" wrap="nowrap" gap="sm">
          <Burger
            opened={opened}
            onClick={toggle}
            hiddenFrom="sm"
            size="sm"
            aria-label="Toggle navigation"
          />
          <Burger
            opened={opened}
            onClick={toggle}
            visibleFrom="sm"
            size="sm"
            aria-label="Toggle navigation"
          />
          <div style={{ flex: 1, minWidth: 0 }}>
            <AppHeader />
          </div>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="sm">
        <AppNavbar />
      </AppShell.Navbar>

      <AppShell.Main>
        <Outlet />
      </AppShell.Main>
    </AppShell>
  );
};

export const router = createBrowserRouter([
  {
    path: APP_PATHS.LOGIN,
    element: createElement(LoginPage)
  },
  {
    path: APP_PATHS.DASHBOARD,
    element: createElement(
      ProtectedRoute,
      null,
      createElement(AppShellLayout)
    ),
    children: [
      { index: true, element: createElement(DashboardPage) },
      {
        path: APP_PATHS.DASHBOARD_ALIAS,
        element: createElement(Navigate, { to: APP_PATHS.DASHBOARD, replace: true })
      },
      {
        path: "/diagrams/:diagramId",
        element: createElement(DiagramPage)
      },
      {
        path: "/objects/:objectId/diagrams/:diagramId",
        element: createElement(DiagramPage)
      },
      {
        path: APP_PATHS.ALARMS,
        element: createElement(AlarmsPage)
      },
      {
        path: APP_PATHS.TRENDS,
        element: createElement(TrendsPage)
      },
      {
        path: "/trends/:tagId",
        element: createElement(TrendsPage)
      }
    ]
  },
  {
    path: APP_PATHS.FORBIDDEN,
    element: createElement(MinimalLayout, { title: "Forbidden" })
  },
  {
    path: APP_PATHS.NOT_FOUND,
    element: createElement(MinimalLayout, { title: "Not found" })
  },
  {
    path: APP_PATHS.SERVER_ERROR,
    element: createElement(MinimalLayout, { title: "Server error" })
  },
  {
    path: APP_PATHS.SESSION_EXPIRED,
    element: createElement(SessionExpiredLayout)
  },
  {
    path: "*",
    element: createElement(Navigate, { to: APP_PATHS.NOT_FOUND, replace: true })
  }
]);

export const RouterProvider = () => {
  return <ReactRouterProvider router={router} />;
};
