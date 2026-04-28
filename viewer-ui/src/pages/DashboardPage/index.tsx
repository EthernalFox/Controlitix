import { notifications } from "@mantine/notifications";
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router";

import { fetchObjects, useDiagramStore } from "@/features/diagram";
import { APP_PATHS } from "@/shared/libs/router";
import { useTopicSubscription } from "@/shared/modules/realtime";
import { Button, Card, Group, Stack, Text, Title } from "@/shared/ui/components";

const OBJECT_LIMIT_NOTIFICATION_ID = "dashboard-object-limit";

export const DashboardPage = () => {
  const objects = useDiagramStore((state) => state.objects);
  const setObjects = useDiagramStore((state) => state.setObjects);

  const [isLoading, setIsLoading] = useState(false);
  const [expandedObjectIDs, setExpandedObjectIDs] = useState<string[]>([]);

  useEffect(() => {
    if (objects.length > 0 || isLoading) {
      return;
    }

    let isCancelled = false;
    setIsLoading(true);

    void fetchObjects(100, 0)
      .then((response) => {
        if (isCancelled) {
          return;
        }

        setObjects(response.items);
      })
      .catch((error: unknown) => {
        console.warn("[dashboard] failed to load objects", error);
      })
      .finally(() => {
        if (isCancelled) {
          return;
        }

        setIsLoading(false);
      });

    return () => {
      isCancelled = true;
    };
  }, [isLoading, objects.length, setObjects]);

  const expandedTopics = useMemo(() => {
    return expandedObjectIDs.map((objectID) => `object:${objectID}`);
  }, [expandedObjectIDs]);

  useTopicSubscription(
    expandedTopics,
    () => {
      // Dashboard keeps object-level subscription active for future aggregate widgets.
    },
    {
      onLimitExceeded: () => {
        notifications.show({
          id: OBJECT_LIMIT_NOTIFICATION_ID,
          title: "Ограничение подписки",
          message: "Слишком крупный объект, отображаются не все теги",
          color: "alarm-warn",
          autoClose: 10000
        });
      }
    }
  );

  const toggleExpanded = (objectID: string) => {
    setExpandedObjectIDs((current) => {
      if (current.includes(objectID)) {
        return current.filter((id) => id !== objectID);
      }

      return [...current, objectID];
    });
  };

  return (
    <Stack gap="md">
      <Title order={2}>Объекты</Title>

      {isLoading ? <Text c="dimmed">Загрузка...</Text> : null}

      {!isLoading && objects.length === 0 ? (
        <Text c="dimmed">Нет объектов с опубликованными мнемосхемами</Text>
      ) : null}

      <Group align="stretch" gap="sm" wrap="wrap">
        {objects.map((objectItem) => {
          const firstDiagramId = objectItem.firstPublishedDiagramId;
          const canOpen = firstDiagramId.trim().length > 0;
          const diagramPath = canOpen
            ? APP_PATHS.DIAGRAM(objectItem.id, firstDiagramId)
            : APP_PATHS.DASHBOARD;
          const isExpanded = expandedObjectIDs.includes(objectItem.id);

          return (
            <Card key={objectItem.id} withBorder p="md" style={{ flex: "1 1 280px" }}>
              <Stack gap="xs" h="100%" justify="space-between">
                <Stack gap={2}>
                  <Title order={4}>{objectItem.name}</Title>
                  <Text size="sm" c="dimmed">
                    {objectItem.description || "Без описания"}
                  </Text>
                  <Text size="xs" c="dimmed">
                    Опубликованных мнемосхем: {objectItem.publishedDiagramCount}
                  </Text>
                  {isExpanded ? (
                    <Text size="xs" c="dimmed">
                      Подписка на object:{objectItem.id} активна
                    </Text>
                  ) : null}
                </Stack>

                <Stack gap="xs">
                  {canOpen ? (
                    <Link to={diagramPath} style={{ textDecoration: "none" }}>
                      <Button variant="secondary" fullWidth>
                        Открыть первую схему
                      </Button>
                    </Link>
                  ) : (
                    <Button variant="secondary" fullWidth disabled>
                      Нет схем
                    </Button>
                  )}

                  <Button variant="secondary" fullWidth onClick={() => toggleExpanded(objectItem.id)}>
                    {isExpanded ? "Свернуть" : "Раскрыть"}
                  </Button>
                </Stack>
              </Stack>
            </Card>
          );
        })}
      </Group>
    </Stack>
  );
};
