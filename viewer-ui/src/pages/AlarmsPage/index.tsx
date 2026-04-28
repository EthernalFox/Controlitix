import { useEffect, useMemo, useState } from "react";

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
  Title
} from "@/shared/ui/components";
import { AlarmRow } from "@/widgets/AlarmRow";

interface ObjectOption {
  value: string;
  label: string;
}

const statusOptions = [
  { value: "active", label: "Active" },
  { value: "acked", label: "Acked" },
  { value: "cleared", label: "Cleared" }
];

const severityOptions = [
  { value: "", label: "All severities" },
  { value: "warn", label: "Warn (lo/hi)" },
  { value: "alarm", label: "Alarm (lolo/hihi)" }
];

export const AlarmsPage = () => {
  const query = useAlarmsStore((state) => state.query);
  const items = useAlarmsStore((state) => state.items);
  const total = useAlarmsStore((state) => state.total);
  const loading = useAlarmsStore((state) => state.loading);
  const error = useAlarmsStore((state) => state.error);
  const load = useAlarmsStore((state) => state.load);
  const refresh = useAlarmsStore((state) => state.refresh);
  const acknowledge = useAlarmsStore((state) => state.acknowledge);

  const [objectOptions, setObjectOptions] = useState<ObjectOption[]>([]);

  useEffect(() => {
    if (items.length === 0) {
      void load();
    }
  }, [items.length, load]);

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

  return (
    <Stack gap="md">
      <Group justify="space-between" align="center">
        <Title order={2}>Alarms</Title>
        <Button variant="ghost" onClick={() => void refresh()}>
          Refresh
        </Button>
      </Group>

      <Card withBorder p="md">
        <Group gap="sm" wrap="wrap">
          <Select
            label="Status"
            data={statusOptions}
            value={query.status}
            onChange={(value) => {
              void load({
                status: (value as "active" | "acked" | "cleared") || "active",
                offset: 0
              });
            }}
            w={180}
          />

          <Select
            label="Severity"
            data={severityOptions}
            value={query.severity}
            onChange={(value) => {
              void load({
                severity: (value as "" | "warn" | "alarm") || "",
                offset: 0
              });
            }}
            w={220}
          />

          <Select
            label="Object"
            data={[{ value: "", label: "All objects" }, ...objectOptions]}
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
                label="From (RFC3339)"
                value={query.from}
                onChange={(event) => {
                  void load({ from: event.currentTarget.value, offset: 0 });
                }}
                placeholder="2026-04-25T10:00:00Z"
                w={240}
              />
              <TextInput
                label="To (RFC3339)"
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

      <Card withBorder p={0}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr>
                <th style={{ textAlign: "left", padding: "10px" }}>Tag</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Object</th>
                <th style={{ textAlign: "left", padding: "10px" }}>State</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Value</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Entered</th>
                <th style={{ textAlign: "left", padding: "10px" }}>Action</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <AlarmRow
                  key={item.tagId}
                  record={item}
                  onAcknowledge={(tagId) => {
                    void acknowledge(tagId);
                  }}
                />
              ))}
            </tbody>
          </table>
        </div>

        {loading ? (
          <Text size="sm" c="dimmed" p="sm">
            Loading alarms...
          </Text>
        ) : null}

        {!loading && items.length === 0 ? (
          <Text size="sm" c="dimmed" p="sm">
            No alarms found for selected filters.
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
          Total: {total}
        </Text>

        <Group gap="xs">
          <Button
            variant="ghost"
            disabled={currentPage <= 1}
            onClick={() => {
              void load({ offset: Math.max(0, query.offset - query.limit) });
            }}
          >
            Prev
          </Button>
          <Text size="sm" c="dimmed">
            Page {currentPage} / {totalPages}
          </Text>
          <Button
            variant="ghost"
            disabled={currentPage >= totalPages}
            onClick={() => {
              void load({ offset: query.offset + query.limit });
            }}
          >
            Next
          </Button>
        </Group>
      </Group>
    </Stack>
  );
};

