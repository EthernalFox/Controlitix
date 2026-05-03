import { useMemo, useState } from "react";

import type { AlarmRecord } from "@/features/alarms";
import {
  Button,
  Group,
  Modal,
  Stack,
  Text,
  Textarea
} from "@/shared/ui/components";

import styles from "./AlarmsBulkAckModal.module.css";

interface AlarmsBulkAckModalProps {
  opened: boolean;
  items: AlarmRecord[];
  loading: boolean;
  onClose: () => void;
  onSubmit: (note: string | null) => Promise<void>;
}

export const AlarmsBulkAckModal = ({
  opened,
  items,
  loading,
  onClose,
  onSubmit
}: AlarmsBulkAckModalProps) => {
  const [note, setNote] = useState("");
  const [showList, setShowList] = useState(false);

  const title = useMemo(() => {
    if (items.length === 1) {
      return "Квитировать 1 тревогу";
    }

    return `Квитировать ${items.length} тревог`;
  }, [items.length]);

  const submit = async () => {
    const normalizedNote = note.trim();
    await onSubmit(normalizedNote ? normalizedNote : null);
    setNote("");
    setShowList(false);
  };

  return (
    <Modal
      opened={opened}
      onClose={loading ? () => undefined : onClose}
      title={title}
      closeOnClickOutside={!loading}
      closeOnEscape={!loading}
    >
      <Stack gap="sm">
        <Textarea
          label="Комментарий"
          placeholder="Опционально"
          value={note}
          maxLength={500}
          onChange={(event) => setNote(event.currentTarget.value)}
          autosize
          minRows={3}
        />

        <Button variant="ghost" onClick={() => setShowList((current) => !current)}>
          {showList ? "Скрыть список тегов" : `Показать ${items.length} тегов`}
        </Button>

        {showList ? (
          <div className={styles.tagsList}>
            {items.map((item) => (
              <Text key={item.tagId} size="sm">
                {item.tagName || item.tagId} • {item.deviceName || "Устройство"}
              </Text>
            ))}
          </div>
        ) : null}

        <Group justify="flex-end" gap="xs">
          <Button variant="ghost" onClick={onClose} disabled={loading}>
            Отмена
          </Button>
          <Button onClick={() => void submit()} loading={loading}>
            Квитировать
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
