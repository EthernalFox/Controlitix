import { useEffect } from "react";
import { useNavigate } from "react-router";

import { useAuthStore } from "@/features/auth";
import {
  getObjectAlarmCounts,
  useObjectsSummaryStore
} from "@/features/objects-summary";
import {
  realtimeClient,
  useRealtimeStore,
  useTopicSubscription
} from "@/shared/modules/realtime";
import type { AlarmMessage } from "@/shared/modules/realtime/types";
import { Button, Card, Group, Stack, Text, Title } from "@/shared/ui/components";
import { ObjectGrid, ObjectGridSkeleton } from "@/widgets/ObjectGrid/ObjectGrid";

import styles from "./DashboardPage.module.css";

const mapAlarmMessage = (message: Omit<AlarmMessage, "t">) => {
  return {
    eventType: message.event_type,
    tagId: message.tag_id,
    stateFrom: message.state_from,
    stateTo: message.state_to,
    value: message.value,
    quality: message.quality,
    ts: message.ts,
    actorId: message.actor_id,
    note: message.note,
    objectId: message.object_id ?? null
  };
};

export const DashboardPage = () => {
  const navigate = useNavigate();
  const status = useObjectsSummaryStore((state) => state.status);
  const objects = useObjectsSummaryStore((state) => state.objects);
  const alarmCounts = useObjectsSummaryStore((state) => state.alarmCounts);
  const error = useObjectsSummaryStore((state) => state.lastError);
  const load = useObjectsSummaryStore((state) => state.load);
  const refresh = useObjectsSummaryStore((state) => state.refresh);
  const applyAlarmEvent = useObjectsSummaryStore((state) => state.applyAlarmEvent);

  const realtimeStatus = useRealtimeStore((state) => state.status);
  const user = useAuthStore((state) => state.user);

  useTopicSubscription(["alarms"], () => {
    // Subscription is required to receive alarm topic frames.
  });

  useEffect(() => {
    const unsubAlarm = realtimeClient.on("alarm", (message) => {
      applyAlarmEvent(mapAlarmMessage(message));
    });

    const unsubBatch = realtimeClient.on("alarms_batch", (message) => {
      message.events.forEach((event) => {
        applyAlarmEvent(mapAlarmMessage(event));
      });
    });

    return () => {
      unsubAlarm();
      unsubBatch();
    };
  }, [applyAlarmEvent]);

  useEffect(() => {
    if (status === "idle") {
      void load();
    }
  }, [load, status]);

  const openPath = (path: string) => {
    void navigate(path);
  };

  const goToEditor = () => {
    const editorUrl = (import.meta.env.VITE_EDITOR_UI_URL ?? "").trim();
    if (!editorUrl) {
      return;
    }

    window.location.assign(editorUrl);
  };

  return (
    <Stack gap="md" className={styles.page}>
      <Group justify="space-between" align="center">
        <Title order={2}>Объекты</Title>
        <Button variant="ghost" onClick={() => void refresh()}>
          Обновить
        </Button>
      </Group>

      {realtimeStatus === "closed" ? (
        <Card withBorder p="sm" className={styles.realtimeWarning}>
          <Group justify="space-between" align="center">
            <Text size="sm">Realtime недоступен. Данные могут устаревать</Text>
            <Button size="xs" variant="secondary" onClick={() => void refresh()}>
              Обновить
            </Button>
          </Group>
        </Card>
      ) : null}

      {status === "loading" ? <ObjectGridSkeleton /> : null}

      {status === "error" ? (
        <Card withBorder p="lg" className={styles.stateCard}>
          <Stack align="center" gap="sm">
            <Text>{error || "Не удалось загрузить данные"}</Text>
            <Button variant="secondary" onClick={() => void refresh()}>
              Повторить
            </Button>
          </Stack>
        </Card>
      ) : null}

      {status === "ready" && objects.length === 0 ? (
        <Card withBorder p="lg" className={styles.stateCard}>
          <Stack align="center" gap="sm">
            <Text size="xl">📭</Text>
            <Text>Нет объектов</Text>
            {user?.roles?.includes("admin") ? (
              <Button variant="secondary" onClick={goToEditor}>
                Перейти в editor-ui
              </Button>
            ) : null}
          </Stack>
        </Card>
      ) : null}

      {status === "ready" && objects.length > 0 ? (
        <ObjectGrid
          objects={objects}
          getCounts={(objectId) => getObjectAlarmCounts(alarmCounts, objectId)}
          onOpen={openPath}
        />
      ) : null}
    </Stack>
  );
};
