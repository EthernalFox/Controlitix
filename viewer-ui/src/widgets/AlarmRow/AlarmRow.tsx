import type { AlarmRecord } from "@/features/alarms";
import { Badge, Button, Group, Text } from "@/shared/ui/components";

interface AlarmRowProps {
  record: AlarmRecord;
  canSelect: boolean;
  selected: boolean;
  pending: boolean;
  onSelect: (tagId: string, checked: boolean) => void;
  onAcknowledge: (tagId: string) => void;
}

const resolveColor = (state: AlarmRecord["state"]): string => {
  switch (state) {
    case "hihi":
    case "lolo":
      return "alarm-alarm";
    case "hi":
    case "lo":
      return "alarm-warn";
    case "uncertain":
      return "alarm-uncertain";
    case "bad":
      return "alarm-bad";
    case "comm_loss":
      return "alarm-comm";
    case "offline":
      return "alarm-offline";
    default:
      return "alarm-ok";
  }
};

const formatDate = (value: string): string => {
  const timestamp = Date.parse(value);
  if (!Number.isFinite(timestamp)) {
    return value;
  }

  return new Date(timestamp).toLocaleString();
};

export const AlarmRow = ({
  record,
  canSelect,
  selected,
  pending,
  onSelect,
  onAcknowledge
}: AlarmRowProps) => {
  const shouldBlink = !record.acked && (record.state === "hi" || record.state === "hihi");

  return (
    <tr style={{ opacity: pending ? 0.55 : 1 }}>
      <td style={{ padding: "8px", borderBottom: "1px solid var(--mantine-color-dark-4)", width: 44 }}>
        {canSelect ? (
          <input
            type="checkbox"
            checked={selected}
            onChange={(event) => onSelect(record.tagId, event.currentTarget.checked)}
            aria-label={`Выбрать ${record.tagName || record.tagId}`}
          />
        ) : null}
      </td>
      <td style={{ padding: "8px", borderBottom: "1px solid var(--mantine-color-dark-4)" }}>
        <Group gap={6} wrap="nowrap">
          <span
            style={{
              width: 8,
              height: 8,
              borderRadius: "50%",
              backgroundColor: "var(--mantine-color-alarm-alarm-5)",
              opacity: shouldBlink ? 1 : 0.3,
              animation: shouldBlink ? "alarmRowBlink 1s linear infinite" : "none"
            }}
          />
          <Text size="sm">{record.tagName || record.tagId}</Text>
        </Group>
      </td>
      <td style={{ padding: "8px", borderBottom: "1px solid var(--mantine-color-dark-4)" }}>
        <Text size="sm">{record.objectName || "—"}</Text>
      </td>
      <td style={{ padding: "8px", borderBottom: "1px solid var(--mantine-color-dark-4)" }}>
        <Badge color={resolveColor(record.state)}>{record.state}</Badge>
      </td>
      <td style={{ padding: "8px", borderBottom: "1px solid var(--mantine-color-dark-4)" }}>
        <Text size="sm">{record.value === null ? "—" : record.value}</Text>
      </td>
      <td style={{ padding: "8px", borderBottom: "1px solid var(--mantine-color-dark-4)" }}>
        <Text size="sm">{formatDate(record.enteredAt)}</Text>
      </td>
      <td style={{ padding: "8px", borderBottom: "1px solid var(--mantine-color-dark-4)" }}>
        {record.acked ? (
          <Badge color="alarm-ack">ACK</Badge>
        ) : (
          <Button size="xs" variant="ghost" onClick={() => onAcknowledge(record.tagId)}>
            Квитировать
          </Button>
        )}
      </td>
    </tr>
  );
};
