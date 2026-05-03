import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router";

import {
  fetchObjectDiagrams,
  fetchObjects,
  useDiagramStore
} from "@/features/diagram";
import { APP_PATHS } from "@/shared/libs/router";
import {
  Card,
  Divider,
  ScrollArea,
  Stack,
  Text,
  Title,
  Button,
  Group
} from "@/shared/ui/components";

const navItems = [
  { label: "Объекты", href: APP_PATHS.DASHBOARD },
  { label: "Тревоги", href: APP_PATHS.ALARMS },
  { label: "Тренды", href: APP_PATHS.TRENDS }
];

const isDashboardRoute = (pathname: string): boolean => {
  return pathname === APP_PATHS.DASHBOARD || pathname === APP_PATHS.DASHBOARD_ALIAS;
};

export const AppNavbar = () => {
  const location = useLocation();

  const objects = useDiagramStore((state) => state.objects);
  const diagramsByObject = useDiagramStore((state) => state.diagramsByObject);
  const setObjects = useDiagramStore((state) => state.setObjects);
  const setObjectDiagrams = useDiagramStore((state) => state.setObjectDiagrams);

  const [expandedObjectId, setExpandedObjectId] = useState<string | null>(null);
  const [objectsLoading, setObjectsLoading] = useState(false);
  const [loadingObjectId, setLoadingObjectId] = useState<string | null>(null);

  useEffect(() => {
    if (objects.length > 0 || objectsLoading) {
      return;
    }

    let isCancelled = false;
    setObjectsLoading(true);

    void fetchObjects(100, 0)
      .then((response) => {
        if (isCancelled) {
          return;
        }

        setObjects(response.items);
      })
      .catch((error: unknown) => {
        console.warn("[diagram] failed to load objects", error);
      })
      .finally(() => {
        if (isCancelled) {
          return;
        }

        setObjectsLoading(false);
      });

    return () => {
      isCancelled = true;
    };
  }, [objects.length, objectsLoading, setObjects]);

  const handleObjectToggle = (objectId: string) => {
    const nextExpandedId = expandedObjectId === objectId ? null : objectId;
    setExpandedObjectId(nextExpandedId);

    if (!nextExpandedId || diagramsByObject[objectId] !== undefined) {
      return;
    }

    setLoadingObjectId(objectId);

    void fetchObjectDiagrams(objectId, 100, 0)
      .then((response) => {
        setObjectDiagrams(objectId, response.items);
      })
      .catch((error: unknown) => {
        console.warn("[diagram] failed to load object diagrams", error);
        setObjectDiagrams(objectId, []);
      })
      .finally(() => {
        setLoadingObjectId((current) => (current === objectId ? null : current));
      });
  };

  return (
    <Stack h="100%" gap="sm">
      <Title order={6}>Навигация</Title>
      <Stack gap={4}>
        {navItems.map((item) => {
          const isActive =
            item.href === APP_PATHS.DASHBOARD
              ? isDashboardRoute(location.pathname)
              : location.pathname.startsWith(item.href);

          return (
            <Card key={item.href} withBorder p="xs">
              <Link
                to={item.href}
                style={{ color: "inherit", textDecoration: "none" }}
              >
                <Text fw={isActive ? 600 : 400}>{item.label}</Text>
              </Link>
            </Card>
          );
        })}
      </Stack>

      <Divider />

      <Title order={6}>Объекты</Title>
      <ScrollArea style={{ flex: 1 }}>
        <Stack gap={6}>
          {objectsLoading ? <Text c="dimmed">Загрузка...</Text> : null}

          {!objectsLoading && objects.length === 0 ? (
            <Text size="sm" c="dimmed">
              Нет опубликованных мнемосхем
            </Text>
          ) : null}

          {objects.map((objectItem) => {
            const diagrams = diagramsByObject[objectItem.id] || [];
            const isExpanded = expandedObjectId === objectItem.id;
            const isObjectLoading = loadingObjectId === objectItem.id;

            return (
              <Card key={objectItem.id} withBorder p="xs">
                <Stack gap={6}>
                  <Group justify="space-between" align="center" wrap="nowrap">
                    <Text fw={600} size="sm">
                      {objectItem.name}
                    </Text>
                    <Button
                      variant="ghost"
                      size="compact-xs"
                      onClick={() => handleObjectToggle(objectItem.id)}
                    >
                      {isExpanded ? "Скрыть" : "Показать"}
                    </Button>
                  </Group>

                  {isExpanded ? (
                    <Stack gap={4}>
                      {isObjectLoading ? <Text size="xs">Загрузка...</Text> : null}

                      {!isObjectLoading && diagrams.length === 0 ? (
                        <Text size="xs" c="dimmed">
                          Нет опубликованных мнемосхем
                        </Text>
                      ) : null}

                      {diagrams.map((diagramItem) => {
                        const href = APP_PATHS.DIAGRAM(objectItem.id, diagramItem.id);
                        const isDiagramActive = location.pathname === href;

                        return (
                          <Link
                            key={diagramItem.id}
                            to={href}
                            style={{ color: "inherit", textDecoration: "none" }}
                          >
                            <Text size="xs" fw={isDiagramActive ? 700 : 400}>
                              {diagramItem.name}
                            </Text>
                          </Link>
                        );
                      })}
                    </Stack>
                  ) : null}
                </Stack>
              </Card>
            );
          })}
        </Stack>
      </ScrollArea>
    </Stack>
  );
};
