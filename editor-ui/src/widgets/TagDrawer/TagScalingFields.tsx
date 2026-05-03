import { NumberInput, SimpleGrid, Stack, Text } from "@shared/ui";

import styles from "./TagScalingFields.module.css";

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

export const TagScalingFields = ({ errors, onChange, values }: TagScalingFieldsProps) => {
  return (
    <Stack gap="sm">
      <Text size="xs" c="dimmed">
        Линейная интерполяция raw → engineering
      </Text>

      <SimpleGrid cols={2} spacing="sm" className={styles.grid}>
        <NumberInput
          label="raw_min"
          value={values.rawMin ?? undefined}
          onChange={(value) => onChange({ rawMin: toNullableNumber(value) })}
          error={errors?.raw_min}
          mono
        />
        <NumberInput
          label="raw_max"
          value={values.rawMax ?? undefined}
          onChange={(value) => onChange({ rawMax: toNullableNumber(value) })}
          error={errors?.raw_max}
          mono
        />
        <NumberInput
          label="eng_min"
          value={values.engMin ?? undefined}
          onChange={(value) => onChange({ engMin: toNullableNumber(value) })}
          error={errors?.eng_min}
          mono
        />
        <NumberInput
          label="eng_max"
          value={values.engMax ?? undefined}
          onChange={(value) => onChange({ engMax: toNullableNumber(value) })}
          error={errors?.eng_max}
          mono
        />
      </SimpleGrid>
    </Stack>
  );
};
