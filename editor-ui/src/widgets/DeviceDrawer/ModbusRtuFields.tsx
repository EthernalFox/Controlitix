import type { ModbusRtuSettings } from "@entities/devices";
import { NumberInput, Select, Stack, TextInput } from "@shared/ui";

interface ModbusRtuFieldsProps {
  settings: ModbusRtuSettings;
  onChange: (settings: ModbusRtuSettings) => void;
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

export const ModbusRtuFields = ({
  settings,
  onChange,
  errors
}: ModbusRtuFieldsProps) => {
  return (
    <Stack gap="sm">
      <TextInput
        label="Serial Port"
        placeholder="/dev/ttyUSB0"
        value={settings.serial_port}
        onChange={(event) =>
          onChange({
            ...settings,
            serial_port: event.currentTarget.value
          })
        }
        error={errors?.serial_port}
        required
      />
      <Select
        label="Baud Rate"
        value={String(settings.baud_rate)}
        data={[
          { value: "9600", label: "9600" },
          { value: "19200", label: "19200" },
          { value: "38400", label: "38400" },
          { value: "57600", label: "57600" },
          { value: "115200", label: "115200" }
        ]}
        onChange={(value) => {
          if (!value) {
            return;
          }

          onChange({
            ...settings,
            baud_rate: toNumber(value, settings.baud_rate)
          });
        }}
        error={errors?.baud_rate}
        required
      />
      <Select
        label="Data Bits"
        value={String(settings.data_bits)}
        data={[
          { value: "7", label: "7" },
          { value: "8", label: "8" }
        ]}
        onChange={(value) => {
          if (!value) {
            return;
          }

          onChange({
            ...settings,
            data_bits: toNumber(value, settings.data_bits)
          });
        }}
        error={errors?.data_bits}
        required
      />
      <Select
        label="Stop Bits"
        value={String(settings.stop_bits)}
        data={[
          { value: "1", label: "1" },
          { value: "2", label: "2" }
        ]}
        onChange={(value) => {
          if (!value) {
            return;
          }

          onChange({
            ...settings,
            stop_bits: toNumber(value, settings.stop_bits)
          });
        }}
        error={errors?.stop_bits}
        required
      />
      <Select
        label="Parity"
        value={settings.parity}
        data={[
          { value: "none", label: "none" },
          { value: "even", label: "even" },
          { value: "odd", label: "odd" }
        ]}
        onChange={(value) => {
          if (!value) {
            return;
          }

          onChange({
            ...settings,
            parity:
              value === "even" || value === "odd"
                ? value
                : "none"
          });
        }}
        error={errors?.parity}
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

