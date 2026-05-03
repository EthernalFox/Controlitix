import type { SnmpV3Settings } from "@entities/devices";
import { NumberInput, PasswordInput, Select, Stack, TextInput } from "@shared/ui";

interface SnmpV3FieldsProps {
  settings: SnmpV3Settings;
  onChange: (settings: SnmpV3Settings) => void;
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

export const SnmpV3Fields = ({ settings, onChange, errors }: SnmpV3FieldsProps) => {
  const showAuth = settings.security_level === "authNoPriv" || settings.security_level === "authPriv";
  const showPriv = settings.security_level === "authPriv";

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
        mono
        required
      />
      <TextInput
        label="Security Name"
        value={settings.security_name}
        onChange={(event) =>
          onChange({
            ...settings,
            security_name: event.currentTarget.value
          })
        }
        error={errors?.security_name}
        required
      />
      <Select
        label="Security Level"
        value={settings.security_level}
        data={[
          { value: "noAuthNoPriv", label: "noAuthNoPriv" },
          { value: "authNoPriv", label: "authNoPriv" },
          { value: "authPriv", label: "authPriv" }
        ]}
        onChange={(value) => {
          if (!value) {
            return;
          }

          onChange({
            ...settings,
            security_level:
              value === "noAuthNoPriv" || value === "authNoPriv" || value === "authPriv"
                ? value
                : "authPriv"
          });
        }}
      />

      {showAuth && (
        <>
          <Select
            label="Auth Protocol"
            value={settings.auth_protocol}
            data={[
              { value: "MD5", label: "MD5" },
              { value: "SHA", label: "SHA" }
            ]}
            onChange={(value) => {
              if (!value) {
                return;
              }

              onChange({
                ...settings,
                auth_protocol: value === "MD5" ? "MD5" : "SHA"
              });
            }}
            error={errors?.auth_protocol}
            required
          />
          <PasswordInput
            label="Auth Password"
            value={settings.auth_password}
            onChange={(event) =>
              onChange({
                ...settings,
                auth_password: event.currentTarget.value
              })
            }
            error={errors?.auth_password}
          />
        </>
      )}

      {showPriv && (
        <>
          <Select
            label="Priv Protocol"
            value={settings.priv_protocol}
            data={[
              { value: "DES", label: "DES" },
              { value: "AES", label: "AES" }
            ]}
            onChange={(value) => {
              if (!value) {
                return;
              }

              onChange({
                ...settings,
                priv_protocol: value === "DES" ? "DES" : "AES"
              });
            }}
            error={errors?.priv_protocol}
            required
          />
          <PasswordInput
            label="Priv Password"
            value={settings.priv_password}
            onChange={(event) =>
              onChange({
                ...settings,
                priv_password: event.currentTarget.value
              })
            }
            error={errors?.priv_password}
          />
        </>
      )}

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
        mono
        required
      />
      <NumberInput
        label="Retry count"
        value={settings.retry_count}
        onChange={(value) =>
          onChange({
            ...settings,
            retry_count: toNumber(value, settings.retry_count)
          })
        }
        error={errors?.retry_count}
        min={0}
        max={10}
        mono
        required
      />
    </Stack>
  );
};
