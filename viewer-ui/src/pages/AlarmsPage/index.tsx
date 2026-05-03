import { useEffect, useMemo, useRef, useState } from "react";

import { useAlarmsStore } from "@/features/alarms";
import { fetchObjects } from "@/features/diagram";
import {
  Button,
  Card,
  Group,
  Select,
  Stack,
  Text,
  TextInput,
  Title,
  notify
} from "@/shared/ui/components";
import { AlarmRow } from "@/widgets/AlarmRow";
import { AlarmsBulkAckModal } from "@/widgets/AlarmsBulkAckModal";

interface ObjectOption {
  value: string;
  label: string;
}

const statusOptions = [
  { value: "active", label: "Активные" },
  { value: "acked", label: "Квитированные" },
  { value: "cleared", label: "Снятые" }
];

const severityOptions = [
  { value: "", label: "Все уровни" },
  { value: "warn", label: "Предупреждения (lo/hi)" },
  { value: "alarm", label: "Тревоги (lolo/hihi)" }
];

const isAckable = (item: { state: string; acked: boolean }): boolean => {
  return item.state !== "ok" && !item.acked;
};

export const AlarmsPage = () => {
  const query = useAlarmsStore((state) => state.query);
  const items = useAlarmsStore((state) => state.items);
  const total = useAlarmsStore((state) => state.total);
  const loading = useAlarmsStore((state) => state.loading);
  const error = useAlarmsStore((state) => state.error);
  const pendingAckTagIds = useAlarmsStore((state) => state.pendingAckTagIds);
  const load = useAlarmsStore((state) => state.load);
  const refresh = useAlarmsStore((state) => state.refresh);
  const acknowledge = useAlarmsStore((state) => state.acknowledge);
  const acknowledgeBulk = useAlarmsStore((state) => state.acknowledgeBulk);
  const startRealtime = useAlarmsStore((state) => state.startRealtime);
  const stopRealtime = useAlarmsStore((state) => state.stopRealtime);

  const [objectOptions, setObjectOptions] = useState<ObjectOption[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [bulkModalOpen, setBulkModalOpen] = useState(false);
  const [bulkSubmitting, setBulkSubmitting] = useState(false);

  const masterRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    if (items.length === 0) {
      void load();
    }
  }, [items.length, load]);

  useEffect(() => {
    startRealtime();

    return () => {
      stopRealtime();
    };
  }, [startRealtime, stopRealtime]);

  useEffect(() => {
    let cancelled = false;

    void fetchObjects(200, 0)
      .then((response) => {
        if (cancelled) {
          return;
        }

        setObjectOptions(
          response.items.map((item) => ({
            value: item.id,
            label: item.name
          }))
        );
      })
      .catch(() => {
        if (cancelled) {
          return;
        }

        setObjectOptions([]);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const ackableItems = useMemo(() => {
    return items.filter((item) => isAckable(item));
  }, [items]);

  useEffect(() => {
    const visibleIds = new Set(ackableItems.map((item) => item.tagId));

    setSelected((previous) => {
      const next = new Set<string>();
      previous.forEach((id) => {
        if (visibleIds.has(id)) {
          next.add(id);
        }
      });
      return next;
    });
  }, [ackableItems]);

  const selectedCount = selected.size;
  const allAckableSelected = ackableItems.length > 0 && selectedCount === ackableItems.length;
  const isIndeterminate = selectedCount > 0 && selectedCount < ackableItems.length;

  useEffect(() => {
    if (!masterRef.current) {
      return;
    }

    masterRef.current.indeterminate = isIndeterminate;
  }, [isIndeterminate]);

  const totalPages = useMemo(() => {
    if (query.limit <= 0) {
      return 1;
    }

    return Math.max(1, Math.ceil(total / query.limit));
  }, [query.limit, total]);

  const currentPage = useMemo(() => {
    if (query.limit <= 0) {
      return 1;
    }

    return Math.floor(query.offset / query.limit) + 1;
  }, [query.limit, query.offset]);

  const selectedItems = useMemo(() => {
    const selectedIds = selected;
    return items.filter((item) => selectedIds.has(item.tagId));
  }, [items, selected]);

  const toggleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelected(new Set(ackableItems.map((item) => item.tagId)));
      return;
    }

    setSelected(new Set());
  };

  const onSelectRow = (tagId: string, checked: boolean) => {
    setSelected((current) => {
      const next = new Set(current);
      if (checked) {
        next.add(tagId);
      } else {
        next.delete(tagId);
      }
      return next;
    });
  };

  const onBulkSubmit = async (note: string | null) => {
    setBulkSubmitting(true);
    try {
      const response = await acknowledgeBulk(Array.from(selected), note);

      if (response.failedN > 0) {
        notify.show({
          color: "alarm-warn",
          title: "Частичное квитирование",
          message: `${response.successN} квитировано, ${response.failedN} не удалось`,
          autoClose: false
        });
      } else {
        notify.show({
          color: "alarm-ok",
          title: "Квитирование выполнено",
          message: `Квитировано ${response.successN} тревог`,
          autoClose: 4000
        });
      }

      setBulkModalOpen(false);
      setSelected(new Set());
    } catch (submitError) {
      notify.show({
        color: "alarm-alarm",
        title: "Квитирование не выполнено",
        message: submitError instanceof Error ? submitError.message : "Сетевая ошибка"
      });
    } finally {
      setBulkSubmitting(false);
    }
  };

  return (
    <Stack gap="md">
      <Group justify="space-between" align="center">
        <Title order={2}>Тревоги</Title>
        <Button variant="ghost" onClick={() => void refresh()}>
          Обновить
        </Button>
      </Group>

      <Card withBorder p="md">
        <Group gap="sm" wrap="wrap">
          <Select
            label="Статус"
            data={statusOptions}
            value={query.status}
            onChange={(value) => {
              void load({
                status: (value as "active" | "acked" | "cleared") || "active",
                offset: 0
              });
            }}
            w={200}
          />

          <Select
            label="Уровень"
            data={severityOptions}
            value={query.severity}
            onChange={(value) => {
              void load({
                severity: (value as "" | "warn" | "alarm") || "",
                offset: 0
              });
            }}
            w={250}
          />

          <Select
            label="Объект"
            data={[{ value: "", label: "Все объекты" }, ...objectOptions]}
            value={query.objectId}
            onChange={(value) => {
              void load({
                objectId: value || "",
                offset: 0
              });
            }}
            searchable
            w={280}
          />

          {query.status === "cleared" ? (
            <>
              <TextInput
                label="С"
                value={query.from}
                onChange={(event) => {
                  void load({ from: event.currentTarget.value, offset: 0 });
                }}
                placeholder="2026-04-25T10:00:00Z"
                w={240}
              />
              <TextInput
                label="По"
                value={query.to}
                onChange={(event) => {
                  void load({ to: event.currentTarget.value, offset: 0 });
                }}
                placeholder="2026-04-25T12:00:00Z"
                w={240}
              />
            </>
          ) : null}
        </Group>
      </Card>

      {selectedCount > 0 ? (
        <Card withBorder p="sm">
          <Group justify="space-between" align="center">
            <Text>Выбрано: {selectedCount}</Text>
            <Group gap="xs">
              <Button variant="secondary" onClick={() => setBulkModalOpen(true)}>
                Квитировать выделенные
              </Button>
              <Button variant="ghost" onClick={() => setSelected(new Set())}>
                Снять выбор
              </Button>
            </Group>
          </Group>
        </Card>
      ) : null}

      <Card withBorder p={0}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th style={{ textAlign: "left", padding: "10px", width: 44 }}>
                  {ackableItems.length > 0 ? (
                    <input
                      ref={masterRef}
                      type="checkbox"
                      checked={allAckableSelected}
                      onChange={(event) => toggleSelectAll(event.currentTarget.checked)}
                      aria-label="Выбрать все"
                    />
                  ) : null}
                </th>
                <th style={{ textAlign: "left", padding: "10px" }}>Тег</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Объект</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Состояние</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Значение</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Начало</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Действие</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => {
                const canSelect = isAckable(item);

                return (
                  <AlarmRow
                    key={item.tagId}
                    record={item}
                    canSelect={canSelect}
                    selected={selected.has(item.tagId)}
                    pending={pendingAckTagIds.includes(item.tagId)}
                    onSelect={onSelectRow}
                    onAcknowledge={(tagId) => {
                      void acknowledge(tagId);
                    }}
                  />
                );
              })}
            </tbody>
          </table>
        </div>

        {loading ? (
          <Text size="sm" c="dimmed" p="sm">
            Загрузка тревог...
          </Text>
        ) : null}

        {!loading && items.length === 0 ? (
          <Text size="sm" c="dimmed" p="sm">
            Тревоги по выбранным фильтрам не найдены.
          </Text>
        ) : null}

        {error ? (
          <Text size="sm" c="alarm-alarm" p="sm">
            {error}
          </Text>
        ) : null}
      </Card>

      <Group justify="space-between" align="center">
        <Text size="sm" c="dimmed">
          Всего: {total}
        </Text>

        <Group gap="xs">
          <Button
            variant="ghost"
            disabled={currentPage <= 1}
            onClick={() => {
              void load({ offset: Math.max(0, query.offset - query.limit) });
            }}
          >
            Назад
          </Button>
          <Text size="sm" c="dimmed">
            Страница {currentPage} / {totalPages}
          </Text>
          <Button
            variant="ghost"
            disabled={currentPage >= totalPages}
            onClick={() => {
              void load({ offset: query.offset + query.limit });
            }}
          >
            Вперёд
          </Button>
        </Group>
      </Group>

      <AlarmsBulkAckModal
        opened={bulkModalOpen}
        items={selectedItems}
        loading={bulkSubmitting}
        onClose={() => setBulkModalOpen(false)}
        onSubmit={onBulkSubmit}
      />
    </Stack>
  );
};
