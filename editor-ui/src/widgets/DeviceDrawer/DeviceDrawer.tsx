import { FormEvent, useEffect, useMemo, useState } from "react";

import {
  cloneDefaultSettings,
  DEVICE_TYPE_LABELS,
  useDevicesStore,
  type DeviceSettings,
  type DeviceTypeName,
  type ModbusRtuSettings,
  type ModbusTcpSettings,
  type SnmpV1V2cSettings,
  type SnmpV3Settings
} from "@entities/devices";
import { ApiRequestError } from "@shared/api";
import {
  Button,
  Drawer,
  Select,
  Skeleton,
  Stack,
  Text,
  TextInput,
  Textarea
} from "@shared/ui";

import { DeviceSettingsForm } from "./DeviceSettingsForm";

interface DeviceDrawerProps {
  opened: boolean;
  mode: "create" | "edit";
  objectId: string;
  deviceId: string | null;
  onClose: () => void;
}

const isDeviceTypeName = (value: string): value is DeviceTypeName =>
  Object.hasOwn(DEVICE_TYPE_LABELS, value);

const toErrorMessage = (error: unknown) => {
  if (error instanceof ApiRequestError) {
    return error.payload.detail || error.payload.title;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Не удалось выполнить запрос";
};

const collectFieldErrors = (error: unknown) => {
  if (!(error instanceof ApiRequestError) || !error.payload.errors) {
    return {};
  }

  return error.payload.errors.reduce<Record<string, string>>((acc, fieldError) => {
    const field = fieldError.field.startsWith("settings.")
      ? fieldError.field.replace("settings.", "")
      : fieldError.field;
    acc[field] = fieldError.message;
    return acc;
  }, {});
};

const validateSettings = (
  typeName: DeviceTypeName,
  settings: DeviceSettings
): Record<string, string> => {
  const errors: Record<string, string> = {};

  const checkPort = (port: number) => {
    if (port < 1 || port > 65535) {
      errors.port = "Порт должен быть в диапазоне 1-65535";
    }
  };

  const checkTimeout = (timeout: number) => {
    if (timeout <= 0) {
      errors.timeout_ms = "Timeout должен быть больше 0";
    }
  };

  if (typeName === "modbus_tcp") {
    const typedSettings = settings as ModbusTcpSettings;
    if (!typedSettings.host.trim()) {
      errors.host = "Укажите host";
    }
    checkPort(typedSettings.port);
    if (typedSettings.slave_id < 1 || typedSettings.slave_id > 247) {
      errors.slave_id = "Slave ID должен быть в диапазоне 1-247";
    }
    checkTimeout(typedSettings.timeout_ms);
  }

  if (typeName === "modbus_rtu") {
    const typedSettings = settings as ModbusRtuSettings;
    if (!typedSettings.serial_port.trim()) {
      errors.serial_port = "Укажите serial port";
    }
    if (typedSettings.slave_id < 1 || typedSettings.slave_id > 247) {
      errors.slave_id = "Slave ID должен быть в диапазоне 1-247";
    }
    checkTimeout(typedSettings.timeout_ms);
  }

  if (typeName === "snmp_v1" || typeName === "snmp_v2c") {
    const typedSettings = settings as SnmpV1V2cSettings;
    if (!typedSettings.host.trim()) {
      errors.host = "Укажите host";
    }
    if (!typedSettings.community.trim()) {
      errors.community = "Укажите community";
    }
    checkPort(typedSettings.port);
    checkTimeout(typedSettings.timeout_ms);
  }

  if (typeName === "snmp_v3") {
    const typedSettings = settings as SnmpV3Settings;
    if (!typedSettings.host.trim()) {
      errors.host = "Укажите host";
    }
    if (!typedSettings.security_name.trim()) {
      errors.security_name = "Укажите security name";
    }
    if (!typedSettings.auth_password.trim()) {
      errors.auth_password = "Укажите auth password";
    }
    if (!typedSettings.priv_password.trim()) {
      errors.priv_password = "Укажите priv password";
    }
    checkPort(typedSettings.port);
    checkTimeout(typedSettings.timeout_ms);
  }

  return errors;
};

export const DeviceDrawer = ({
  opened,
  mode,
  objectId,
  deviceId,
  onClose
}: DeviceDrawerProps) => {
  const {
    createDevice,
    deviceTypes,
    getDevice,
    updateDevice
  } = useDevicesStore();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [typeId, setTypeId] = useState<string | null>(null);
  const [settings, setSettings] = useState<DeviceSettings | null>(null);
  const [isLoadingDevice, setIsLoadingDevice] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const selectedType = useMemo(
    () => deviceTypes.find((deviceType) => String(deviceType.id) === typeId) ?? null,
    [deviceTypes, typeId]
  );

  const selectedTypeName: DeviceTypeName | null = selectedType?.name ?? null;

  useEffect(() => {
    if (!opened) {
      return;
    }

    setFieldErrors({});
    setFormError(null);

    if (mode === "create") {
      setName("");
      setDescription("");
      setTypeId(null);
      setSettings(null);
      setIsLoadingDevice(false);
      return;
    }

    if (!deviceId) {
      setFormError("Не удалось определить устройство для редактирования");
      return;
    }

    let isCancelled = false;
    setIsLoadingDevice(true);
    void getDevice(deviceId)
      .then((device) => {
        if (isCancelled) {
          return;
        }

        setName(device.name);
        setDescription(device.description ?? "");
        setTypeId(String(device.typeId));
        setSettings(device.settings);
      })
      .catch((error) => {
        if (!isCancelled) {
          setFormError(toErrorMessage(error));
        }
      })
      .finally(() => {
        if (!isCancelled) {
          setIsLoadingDevice(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [deviceId, getDevice, mode, opened]);

  const drawerTitle =
    mode === "create" ? "Новое устройство" : `Редактирование: ${name || "устройство"}`;

  const onTypeChange = (nextTypeId: string | null) => {
    setTypeId(nextTypeId);
    setFieldErrors((currentErrors) => {
      const nextErrors = { ...currentErrors };
      delete nextErrors.type_id;
      return nextErrors;
    });

    if (!nextTypeId) {
      setSettings(null);
      return;
    }

    const nextType = deviceTypes.find(
      (deviceType) => String(deviceType.id) === nextTypeId
    );
    if (!nextType) {
      setSettings(null);
      return;
    }

    setSettings(cloneDefaultSettings(nextType.name));
  };

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedName = name.trim();
    const trimmedDescription = description.trim();
    const nextFieldErrors: Record<string, string> = {};

    if (!trimmedName) {
      nextFieldErrors.name = "Введите название устройства";
    }

    if (!typeId || !selectedTypeName) {
      nextFieldErrors.type_id = "Выберите тип устройства";
    }

    if (!settings) {
      nextFieldErrors.settings = "Заполните настройки устройства";
    }

    if (selectedTypeName && settings) {
      Object.assign(nextFieldErrors, validateSettings(selectedTypeName, settings));
    }

    if (Object.keys(nextFieldErrors).length > 0) {
      setFieldErrors(nextFieldErrors);
      return;
    }

    setIsSubmitting(true);
    setFormError(null);
    setFieldErrors({});
    const validatedSettings = settings as DeviceSettings;

    try {
      if (mode === "create") {
        await createDevice(objectId, {
          name: trimmedName,
          description: trimmedDescription || undefined,
          type_id: Number(typeId),
          settings: validatedSettings
        });
      } else {
        if (!deviceId) {
          setFormError("Не удалось определить устройство для редактирования");
          setIsSubmitting(false);
          return;
        }

        await updateDevice(deviceId, {
          name: trimmedName,
          description: trimmedDescription || undefined,
          settings: validatedSettings
        });
      }

      onClose();
    } catch (error) {
      setFieldErrors(collectFieldErrors(error));
      setFormError(toErrorMessage(error));
      setIsSubmitting(false);
    }
  };

  return (
    <Drawer
      opened={opened}
      onClose={onClose}
      title={drawerTitle}
      position="right"
      size={420}
    >
      {isLoadingDevice ? (
        <Stack>
          <Skeleton h={36} radius="sm" />
          <Skeleton h={36} radius="sm" />
          <Skeleton h={36} radius="sm" />
          <Skeleton h={36} radius="sm" />
          <Skeleton h={36} radius="sm" />
        </Stack>
      ) : (
        <form onSubmit={onSubmit}>
          <Stack>
            <TextInput
              label="Название"
              value={name}
              onChange={(event) => setName(event.currentTarget.value)}
              error={fieldErrors.name}
              required
            />
            <Textarea
              label="Описание"
              value={description}
              onChange={(event) => setDescription(event.currentTarget.value)}
              error={fieldErrors.description}
              minRows={3}
            />
            <Select
              label="Тип устройства"
              value={typeId}
              data={deviceTypes.map((deviceType) => ({
                value: String(deviceType.id),
                label: DEVICE_TYPE_LABELS[deviceType.name]
              }))}
              onChange={onTypeChange}
              error={fieldErrors.type_id}
              disabled={mode === "edit"}
              required
            />

            {fieldErrors.settings && (
              <Text c="red" size="sm">
                {fieldErrors.settings}
              </Text>
            )}

            <DeviceSettingsForm
              typeName={selectedTypeName}
              settings={settings}
              onChange={setSettings}
              errors={fieldErrors}
            />

            {typeId && selectedType && (
              <Text size="xs" c="dimmed">
                Тип: {isDeviceTypeName(selectedType.name) ? DEVICE_TYPE_LABELS[selectedType.name] : selectedType.name}
              </Text>
            )}

            {formError && (
              <Text c="red" size="sm">
                {formError}
              </Text>
            )}

            <Button type="submit" loading={isSubmitting}>
              Сохранить
            </Button>
            <Button
              type="button"
              variant="secondary"
              onClick={onClose}
              disabled={isSubmitting}
            >
              Отмена
            </Button>
          </Stack>
        </form>
      )}
    </Drawer>
  );
};
