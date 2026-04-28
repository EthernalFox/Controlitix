import { DateTimePicker } from "@mantine/dates";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import { useEffect, useMemo, useState } from "react";

import {
  RANGE_PRESET_OPTIONS,
  useTrendsStore,
  validateCustomRange
} from "@/features/trends";
import { Button, Group, Stack, Text } from "@/shared/ui/components";

dayjs.extend(utc);

const toInputDateTime = (iso: string): string => {
  return dayjs(iso).format("YYYY-MM-DDTHH:mm:ss");
};

export const TrendRangePicker = () => {
  const range = useTrendsStore((state) => state.range);
  const setRange = useTrendsStore((state) => state.setRange);

  const [customFrom, setCustomFrom] = useState<string | null>(
    toInputDateTime(dayjs().subtract(1, "hour").toISOString())
  );
  const [customTo, setCustomTo] = useState<string | null>(
    toInputDateTime(dayjs().toISOString())
  );

  useEffect(() => {
    if (range.kind !== "custom") {
      return;
    }

    setCustomFrom(toInputDateTime(range.from));
    setCustomTo(toInputDateTime(range.to));
  }, [range]);

  const selectedPreset = range.kind === "preset" ? range.preset : "custom";

  const validationMessage = useMemo(() => {
    if (!customFrom || !customTo) {
      return "Выберите корректные дату и время";
    }

    return validateCustomRange(
      dayjs(customFrom).utc().toISOString(),
      dayjs(customTo).utc().toISOString()
    );
  }, [customFrom, customTo]);

  const applyCustomRange = () => {
    if (!customFrom || !customTo) {
      return;
    }

    const fromIso = dayjs(customFrom).utc().toISOString();
    const toIso = dayjs(customTo).utc().toISOString();
    const validationError = validateCustomRange(fromIso, toIso);

    if (validationError) {
      return;
    }

    setRange({
      kind: "custom",
      from: fromIso,
      to: toIso
    });
  };

  return (
    <Stack gap="xs">
      <Group gap="xs" wrap="wrap">
        {RANGE_PRESET_OPTIONS.map((option) => (
          <Button
            key={option.value}
            variant={selectedPreset === option.value ? "primary" : "secondary"}
            size="xs"
            onClick={() => {
              if (option.value === "custom") {
                return;
              }

              setRange({
                kind: "preset",
                preset: option.value
              });
            }}
          >
            {option.label}
          </Button>
        ))}
      </Group>

      <Group gap="xs" align="flex-end" wrap="wrap">
        <DateTimePicker
          label="С"
          value={customFrom}
          onChange={setCustomFrom}
          valueFormat="DD.MM.YYYY HH:mm"
          size="xs"
          style={{ minWidth: 220 }}
        />
        <DateTimePicker
          label="По"
          value={customTo}
          onChange={setCustomTo}
          valueFormat="DD.MM.YYYY HH:mm"
          size="xs"
          style={{ minWidth: 220 }}
        />
        <Button
          size="xs"
          onClick={applyCustomRange}
          disabled={Boolean(validationMessage)}
        >
          Применить свой диапазон
        </Button>
      </Group>

      {validationMessage ? (
        <Text c="alarm-alarm" size="sm">
          {validationMessage}
        </Text>
      ) : null}
    </Stack>
  );
};
