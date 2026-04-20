import type { SnmpV1V2cSettings } from "@entities/devices";
import { NumberInput, Stack, TextInput } from "@shared/ui";

interface SnmpV1V2cFieldsProps {
  settings: SnmpV1V2cSettings;
  onChange: (settings: SnmpV1V2cSettings) => void;
  errors?: Record<string, string>;
}

const toNumber = (value: string | number, fallback: number) => {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  return fallback;
};

export const SnmpV1V2cFields = ({
  settings,
  onChange,
  errors
}: SnmpV1V2cFieldsProps) => {
  return (
    <Stack gap="sm">
      <TextInput
        label="Host"
        value={settings.host}
        onChange={(event) =>
          onChange({
            ...settings,
            host: event.currentTarget.value
          })
        }
        error={errors?.host}
        required
      />
      <NumberInput
        label="Port"
        value={settings.port}
        onChange={(value) =>
          onChange({
            ...settings,
            port: toNumber(value, settings.port)
          })
        }
        error={errors?.port}
        min={1}
        max={65535}
        required
      />
      <TextInput
        label="Community"
        value={settings.community}
        onChange={(event) =>
          onChange({
            ...settings,
            community: event.currentTarget.value
          })
        }
        error={errors?.community}
        required
      />
      <NumberInput
        label="Timeout, мс"
        value={settings.timeout_ms}
        onChange={(value) =>
          onChange({
            ...settings,
            timeout_ms: toNumber(value, settings.timeout_ms)
          })
        }
        error={errors?.timeout_ms}
        min={1}
        required
      />
    </Stack>
  );
};

