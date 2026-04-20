import { REGISTER_TYPE_LABELS, type ModbusRegisterType } from "@entities/tags";
import { NumberInput, Select, Stack } from "@shared/ui";

interface ModbusAddressFieldsProps {
  registerType: ModbusRegisterType | null;
  address: number | null;
  errors?: Record<string, string>;
  onChange: (patch: {
    registerType?: ModbusRegisterType | null;
    address?: number | null;
  }) => void;
}

const REGISTER_TYPE_OPTIONS = (
  Object.entries(REGISTER_TYPE_LABELS) as Array<[ModbusRegisterType, string]>
).map(([value, label]) => ({ value, label }));

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

export const ModbusAddressFields = ({
  address,
  errors,
  onChange,
  registerType
}: ModbusAddressFieldsProps) => {
  return (
    <Stack gap="sm">
      <Select
        label="Тип регистра"
        data={REGISTER_TYPE_OPTIONS}
        value={registerType}
        onChange={(value) =>
          onChange({ registerType: (value as ModbusRegisterType | null) ?? null })
        }
        error={errors?.["address.register_type"]}
        required
      />
      <NumberInput
        label="Адрес регистра"
        min={0}
        max={65535}
        value={address ?? undefined}
        onChange={(value) => onChange({ address: toNullableNumber(value) })}
        error={errors?.["address.address"]}
        required
      />
    </Stack>
  );
};
