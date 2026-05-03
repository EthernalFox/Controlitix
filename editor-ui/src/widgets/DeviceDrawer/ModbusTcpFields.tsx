import type { ModbusTcpSettings } from "@entities/devices";
import { NumberInput, Stack, TextInput } from "@shared/ui";

interface ModbusTcpFieldsProps {
  settings: ModbusTcpSettings;
  onChange: (settings: ModbusTcpSettings) => void;
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

export const ModbusTcpFields = ({ settings, onChange, errors }: ModbusTcpFieldsProps) => {
  return (
    <Stack gap="sm">
      <TextInput
        label="Host"
        placeholder="192.168.1.100"
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
        mono
        required
      />
      <NumberInput
        label="Slave ID"
        value={settings.slave_id}
        onChange={(value) =>
          onChange({
            ...settings,
            slave_id: toNumber(value, settings.slave_id)
          })
        }
        error={errors?.slave_id}
        min={1}
        max={247}
        mono
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
        min={100}
        max={30000}
        mono
        required
      />
    </Stack>
  );
};
