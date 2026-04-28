import { useEffect, useMemo, useState } from "react";

import { fetchTrendTags, useTrendsStore } from "@/features/trends";
import {
  Button,
  Card,
  Group,
  Loader,
  ScrollArea,
  Stack,
  Text,
  TextInput,
  Title,
  Tooltip
} from "@/shared/ui/components";

const MAX_SELECTED_TAGS = 6;

export const TrendTagPicker = () => {
  const selectedTagIds = useTrendsStore((state) => state.selectedTagIds);
  const addTag = useTrendsStore((state) => state.addTag);
  const removeTag = useTrendsStore((state) => state.removeTag);

  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(false);
  const [options, setOptions] = useState<Array<{ id: string; name: string }>>([]);

  const maxSelectedReached = selectedTagIds.length >= MAX_SELECTED_TAGS;

  useEffect(() => {
    let isCancelled = false;

    const timeoutId = window.setTimeout(async () => {
      setLoading(true);

      try {
        const nextOptions = await fetchTrendTags(search, 20);

        if (!isCancelled) {
          setOptions(nextOptions);
        }
      } catch {
        if (!isCancelled) {
          setOptions([]);
        }
      } finally {
        if (!isCancelled) {
          setLoading(false);
        }
      }
    }, 250);

    return () => {
      isCancelled = true;
      window.clearTimeout(timeoutId);
    };
  }, [search]);

  const selectedSet = useMemo(() => {
    return new Set(selectedTagIds);
  }, [selectedTagIds]);

  return (
    <Stack gap="sm">
      <Title order={5}>Теги</Title>
      <TextInput
        placeholder="Поиск тега"
        value={search}
        onChange={(event) => setSearch(event.currentTarget.value)}
      />

      <ScrollArea h={300}>
        <Stack gap="xs">
          {loading ? (
            <Group justify="center" py="md">
              <Loader />
            </Group>
          ) : null}

          {!loading && options.length === 0 ? (
            <Text size="sm" c="dimmed">
              Теги не найдены
            </Text>
          ) : null}

          {options.map((option) => {
            const selected = selectedSet.has(option.id);
            const disabled = !selected && maxSelectedReached;

            return (
              <Card key={option.id} withBorder p="xs">
                <Group justify="space-between" align="center" wrap="nowrap">
                  <Stack gap={0} style={{ minWidth: 0 }}>
                    <Text size="sm" truncate>
                      {option.name}
                    </Text>
                    <Text size="xs" c="dimmed" truncate>
                      {option.id}
                    </Text>
                  </Stack>

                  <Tooltip
                    disabled={!disabled}
                    label="Не более 6 тегов одновременно"
                    withArrow
                  >
                    <div>
                      <Button
                        variant={selected ? "secondary" : "primary"}
                        size="xs"
                        disabled={disabled}
                        onClick={() => {
                          if (selected) {
                            removeTag(option.id);
                            return;
                          }

                          addTag(option.id);
                        }}
                      >
                        {selected ? "Убрать" : "Добавить"}
                      </Button>
                    </div>
                  </Tooltip>
                </Group>
              </Card>
            );
          })}
        </Stack>
      </ScrollArea>
    </Stack>
  );
};
