import { NumberInput, Stack, Text } from "@shared/ui";

interface TagSetpointsFieldsValue {
  lolo: number | null;
  lo: number | null;
  hi: number | null;
  hihi: number | null;
}

interface TagSetpointsFieldsProps {
  values: TagSetpointsFieldsValue;
  errors?: Record<string, string>;
  onChange: (patch: Partial<TagSetpointsFieldsValue>) => void;
}

const toNullableNumber = (value: string | number): number | null => {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  return null;
};

export const TagSetpointsFields = ({
  errors,
  onChange,
  values
}: TagSetpointsFieldsProps) => {
  return (
    <Stack gap="sm">
      <Text size="xs" c="dimmed">
        Должно выполняться LoLo &lt; Lo &lt; Hi &lt; HiHi (для заполненных).
      </Text>

      <NumberInput
        label="LoLo"
        value={values.lolo ?? undefined}
        onChange={(value) => onChange({ lolo: toNullableNumber(value) })}
        error={errors?.lolo}
      />
      <NumberInput
        label="Lo"
        value={values.lo ?? undefined}
        onChange={(value) => onChange({ lo: toNullableNumber(value) })}
        error={errors?.lo}
      />
      <NumberInput
        label="Hi"
        value={values.hi ?? undefined}
        onChange={(value) => onChange({ hi: toNullableNumber(value) })}
        error={errors?.hi}
      />
      <NumberInput
        label="HiHi"
        value={values.hihi ?? undefined}
        onChange={(value) => onChange({ hihi: toNullableNumber(value) })}
        error={errors?.hihi}
      />
    </Stack>
  );
};
