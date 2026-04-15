import { useMemo, useState } from "react";

import { monitoringObjectsMock, type MonitoringObjectModel } from "@entities/objects";
import { Logo } from "@shared/ui";
import { Header, Layout, Navbar, Aside } from "@shared/ui";
import { Card, Stack, Text } from "@shared/ui";
import { AppNavbar } from "@widgets/AppNavbar";

const formatDateTime = (value: string) => new Date(value).toLocaleString();

const EMPTY_TEXT = "—";

const getDescription = (value: string | null) => value ?? EMPTY_TEXT;

export default function MonitoringObjectsPage() {
  const objects = useMemo(() => monitoringObjectsMock, []);
  const [activeId, setActiveId] = useState<MonitoringObjectModel["id"] | null>(
    objects[0]?.id ?? null
  );

  const activeObject = objects.find((object) => object.id === activeId) ?? null;

  return (
    <Layout
      header={<Header main={<Logo />} />}
      navbar={<Navbar center={<AppNavbar />} />}
      aside={
        <Aside
          center={
            <div style={{ padding: 12 }}>
              <Card p="md" withBorder>
                <Text fw={600} mb={8}>
                  Детали
                </Text>
                {activeObject ? (
                  <div style={{ display: "grid", gap: 8 }}>
                    <Text>
                      <b>ID:</b> {activeObject.id}
                    </Text>
                    <Text>
                      <b>Название:</b> {activeObject.name}
                    </Text>
                    <Text>
                      <b>Описание:</b> {getDescription(activeObject.description)}
                    </Text>
                    <Text>
                      <b>Создано:</b> {formatDateTime(activeObject.createdAt)}
                    </Text>
                    <Text>
                      <b>Обновлено:</b> {formatDateTime(activeObject.updatedAt)}
                    </Text>
                  </div>
                ) : (
                  <Text c="dimmed">Нет данных</Text>
                )}
              </Card>
            </div>
          }
        />
      }
    >
      <div style={{ padding: 16 }}>
        <Text size="xl" fw={600} mb={12}>
          Объекты мониторинга
        </Text>

        <Stack gap={10}>
          {objects.map((object) => {
            const selected = object.id === activeId;
            return (
              <div
                key={object.id}
                role="button"
                tabIndex={0}
                onClick={() => setActiveId(object.id)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    setActiveId(object.id);
                  }
                }}
                style={{ cursor: "pointer" }}
              >
                <Card
                  withBorder
                  p="md"
                  style={{
                    background: selected ? "rgba(84, 116, 180, 0.14)" : undefined
                  }}
                >
                  <div style={{ display: "grid", gap: 4 }}>
                    <Text fw={600}>{object.name}</Text>
                    <Text c="dimmed" size="sm">
                      {getDescription(object.description)}
                    </Text>
                    <Text c="dimmed" size="xs">
                      Обновлено: {formatDateTime(object.updatedAt)}
                    </Text>
                  </div>
                </Card>
              </div>
            );
          })}
        </Stack>
      </div>
    </Layout>
  );
}
