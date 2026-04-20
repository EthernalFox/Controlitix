import { NumberInput, Stack, Text } from "@shared/ui";

interface TagScalingFieldsValue {
  rawMin: number | null;
  rawMax: number | null;
  engMin: number | null;
  engMax: number | null;
  factor: number | null;
  offset: number | null;
}

interface TagScalingFieldsProps {
  values: TagScalingFieldsValue;
  errors?: Record<string, string>;
  onChange: (patch: Partial<TagScalingFieldsValue>) => void;
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

export const TagScalingFields = ({
  errors,
  onChange,
  values
}: TagScalingFieldsProps) => {
  return (
    <Stack gap="sm">
      <Text size="xs" c="dimmed">
        Линейная интерполяция: укажите все 4 поля raw/eng. Либо factor/offset.
      </Text>

      <NumberInput
        label="Raw min"
        value={values.rawMin ?? undefined}
        onChange={(value) => onChange({ rawMin: toNullableNumber(value) })}
        error={errors?.raw_min}
      />
      <NumberInput
        label="Raw max"
        value={values.rawMax ?? undefined}
        onChange={(value) => onChange({ rawMax: toNullableNumber(value) })}
        error={errors?.raw_max}
      />
      <NumberInput
        label="Eng min"
        value={values.engMin ?? undefined}
        onChange={(value) => onChange({ engMin: toNullableNumber(value) })}
        error={errors?.eng_min}
      />
      <NumberInput
        label="Eng max"
        value={values.engMax ?? undefined}
        onChange={(value) => onChange({ engMax: toNullableNumber(value) })}
        error={errors?.eng_max}
      />
      <NumberInput
        label="Factor"
        value={values.factor ?? undefined}
        onChange={(value) => onChange({ factor: toNullableNumber(value) })}
        error={errors?.factor}
      />
      <NumberInput
        label="Offset"
        value={values.offset ?? undefined}
        onChange={(value) => onChange({ offset: toNullableNumber(value) })}
        error={errors?.offset}
      />
    </Stack>
  );
};
