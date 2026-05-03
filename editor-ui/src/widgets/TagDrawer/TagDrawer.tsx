import { FormEvent, useEffect, useMemo, useState } from "react";

import type { DataType } from "@entities/data-types";
import { useDataTypesStore } from "@entities/data-types";
import type { Device } from "@entities/devices";
import { useDevicesStore } from "@entities/devices";
import {
  OID_PATTERN,
  useTagsStore,
  type DeviceProtocol,
  type ModbusAddress,
  type ModbusAddressPayload,
  type ModbusRegisterType,
  type SnmpAddress,
  type SnmpAddressPayload,
  type Tag,
  type TagAddress,
  type TagAddressPayload,
  type TagEditDiff,
  type TagScalingPayload,
  type TagSetpoints,
  type UpdateTagParamsPayload
} from "@entities/tags";
import type { Unit } from "@entities/units";
import { useUnitsStore } from "@entities/units";
import { ApiRequestError } from "@shared/api";
import { Button, Divider, Drawer, Group, Select, Skeleton, Stack, Switch, Text, TextInput, Textarea } from "@shared/ui";

import { TagAddressForm } from "./TagAddressForm";
import styles from "./TagDrawer.module.css";
import { TagScalingFields } from "./TagScalingFields";
import { TagSetpointsFields } from "./TagSetpointsFields";

interface TagDrawerProps {
  opened: boolean;
  mode: "create" | "edit";
  objectId: string;
  tagId: string | null;
  onClose: () => void;
}

interface TagFormState {
  deviceId: string | null;
  name: string;
  description: string;
  dataTypeId: string | null;
  unitId: string | null;
  address: {
    registerType: ModbusRegisterType | null;
    address: number | null;
    oid: string;
  };
  useScaling: boolean;
  scaling: {
    rawMin: number | null;
    rawMax: number | null;
    engMin: number | null;
    engMax: number | null;
    factor: number | null;
    offset: number | null;
  };
  useSetpoints: boolean;
  setpoints: {
    lolo: number | null;
    lo: number | null;
    hi: number | null;
    hihi: number | null;
  };
}

const INITIAL_FORM_STATE: TagFormState = {
  deviceId: null,
  name: "",
  description: "",
  dataTypeId: null,
  unitId: null,
  address: {
    registerType: null,
    address: null,
    oid: ""
  },
  useScaling: false,
  scaling: {
    rawMin: null,
    rawMax: null,
    engMin: null,
    engMax: null,
    factor: null,
    offset: null
  },
  useSetpoints: false,
  setpoints: {
    lolo: null,
    lo: null,
    hi: null,
    hihi: null
  }
};

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
    const normalizedField = fieldError.field.startsWith("params.")
      ? fieldError.field.replace("params.", "")
      : fieldError.field;
    acc[normalizedField] = fieldError.message;
    return acc;
  }, {});
};

const isModbusAddress = (address: TagAddress): address is ModbusAddress => "registerType" in address;

const isSnmpAddress = (address: TagAddress): address is SnmpAddress => "oid" in address;

const toAddressPayloadFromTagAddress = (address: TagAddress): TagAddressPayload => {
  if (isModbusAddress(address)) {
    return {
      register_type: address.registerType,
      address: address.address
    };
  }

  return {
    oid: address.oid
  };
};

const toScalingPayload = (values: TagFormState["scaling"]): TagScalingPayload => ({
  raw_min: values.rawMin,
  raw_max: values.rawMax,
  eng_min: values.engMin,
  eng_max: values.engMax,
  factor: values.factor,
  offset: values.offset
});

const toSetpointsPayload = (values: TagFormState["setpoints"]): TagSetpoints => ({
  lolo: values.lolo,
  lo: values.lo,
  hi: values.hi,
  hihi: values.hihi
});

const hasPayloadChanges = (left: unknown, right: unknown) => JSON.stringify(left) !== JSON.stringify(right);

const mapTagToFormState = (tag: Tag): TagFormState => {
  const nextState: TagFormState = {
    deviceId: tag.deviceId,
    name: tag.name,
    description: tag.description ?? "",
    dataTypeId: tag.params ? String(tag.params.dataTypeId) : null,
    unitId: tag.params?.unitId ? String(tag.params.unitId) : null,
    address: {
      registerType: null,
      address: null,
      oid: ""
    },
    useScaling: Boolean(tag.scaling),
    scaling: {
      rawMin: tag.scaling?.rawMin ?? null,
      rawMax: tag.scaling?.rawMax ?? null,
      engMin: tag.scaling?.engMin ?? null,
      engMax: tag.scaling?.engMax ?? null,
      factor: tag.scaling?.factor ?? null,
      offset: tag.scaling?.offset ?? null
    },
    useSetpoints: Boolean(tag.setpoints),
    setpoints: {
      lolo: tag.setpoints?.lolo ?? null,
      lo: tag.setpoints?.lo ?? null,
      hi: tag.setpoints?.hi ?? null,
      hihi: tag.setpoints?.hihi ?? null
    }
  };

  if (tag.params?.address && isModbusAddress(tag.params.address)) {
    nextState.address.registerType = tag.params.address.registerType;
    nextState.address.address = tag.params.address.address;
  }

  if (tag.params?.address && isSnmpAddress(tag.params.address)) {
    nextState.address.oid = tag.params.address.oid;
  }

  return nextState;
};

const buildUnitOptions = (units: Unit[]) => {
  const groupedUnits = units.reduce<Record<string, Unit[]>>((acc, unit) => {
    if (!acc[unit.category]) {
      acc[unit.category] = [];
    }

    acc[unit.category].push(unit);
    return acc;
  }, {});

  return Object.entries(groupedUnits).map(([category, categoryUnits]) => ({
    group: category,
    items: categoryUnits.map((unit) => ({
      value: String(unit.id),
      label: `${unit.symbol} (${unit.name})`
    }))
  }));
};

const toProtocol = (device: Device | undefined): DeviceProtocol | null =>
  (device?.typeName as DeviceProtocol | undefined) ?? null;

const validateForm = (formState: TagFormState, protocol: DeviceProtocol | null): Record<string, string> => {
  const errors: Record<string, string> = {};

  if (!formState.name.trim()) {
    errors.name = "Введите название тега";
  }

  if (!formState.deviceId) {
    errors.device_id = "Выберите устройство";
  }

  if (!formState.dataTypeId) {
    errors.data_type_id = "Выберите тип данных";
  }

  if (!protocol) {
    errors.address = "Выберите устройство, чтобы заполнить адрес";
  }

  if (protocol === "modbus_rtu" || protocol === "modbus_tcp") {
    if (!formState.address.registerType) {
      errors["address.register_type"] = "Выберите тип регистра";
    }

    if (formState.address.address === null || formState.address.address < 0 || formState.address.address > 65535) {
      errors["address.address"] = "Адрес должен быть в диапазоне 0-65535";
    }
  }

  if (protocol && protocol.startsWith("snmp")) {
    if (!OID_PATTERN.test(formState.address.oid.trim())) {
      errors["address.oid"] = "OID должен соответствовать формату 1.3.6.1...";
    }
  }

  if (formState.useSetpoints) {
    const { hihi, hi, lo, lolo } = formState.setpoints;

    if (lolo !== null && lo !== null && lolo >= lo) {
      errors.lolo = "LoLo должно быть меньше Lo";
    }
    if (lo !== null && hi !== null && lo >= hi) {
      errors.lo = "Lo должно быть меньше Hi";
    }
    if (hi !== null && hihi !== null && hi >= hihi) {
      errors.hi = "Hi должно быть меньше HiHi";
    }
  }

  return errors;
};

const buildParamsPayload = (formState: TagFormState, protocol: DeviceProtocol): UpdateTagParamsPayload => {
  const dataTypeID = Number(formState.dataTypeId);
  const unitID = formState.unitId ? Number(formState.unitId) : null;

  if (protocol === "modbus_rtu" || protocol === "modbus_tcp") {
    const addressPayload: ModbusAddressPayload = {
      register_type: formState.address.registerType ?? "holding",
      address: formState.address.address ?? 0
    };

    return {
      data_type_id: dataTypeID,
      unit_id: unitID,
      address: addressPayload
    };
  }

  const addressPayload: SnmpAddressPayload = {
    oid: formState.address.oid.trim()
  };

  return {
    data_type_id: dataTypeID,
    unit_id: unitID,
    address: addressPayload
  };
};

const buildEditDiff = (originalTag: Tag, formState: TagFormState, protocol: DeviceProtocol): TagEditDiff => {
  const diff: TagEditDiff = {};

  const nextName = formState.name.trim();
  const nextDescription = formState.description.trim();
  const originalDescription = originalTag.description ?? "";

  if (nextName !== originalTag.name || nextDescription !== originalDescription) {
    diff.meta = {
      name: nextName,
      description: nextDescription
    };
  }

  const nextParams = buildParamsPayload(formState, protocol);
  const originalAddress = originalTag.params?.address;
  const originalAddressPayload = originalAddress ? toAddressPayloadFromTagAddress(originalAddress) : null;
  const originalParamsPayload: UpdateTagParamsPayload | null = originalTag.params
    ? {
        data_type_id: originalTag.params.dataTypeId,
        unit_id: originalTag.params.unitId,
        address: originalAddressPayload ?? nextParams.address
      }
    : null;

  if (!originalParamsPayload || hasPayloadChanges(nextParams, originalParamsPayload)) {
    diff.params = nextParams;
  }

  const nextSetpoints = toSetpointsPayload(formState.setpoints);
  if (!formState.useSetpoints && originalTag.setpoints) {
    diff.setpoints = "delete";
  }
  if (formState.useSetpoints && (!originalTag.setpoints || hasPayloadChanges(nextSetpoints, originalTag.setpoints))) {
    diff.setpoints = nextSetpoints;
  }

  const nextScaling = toScalingPayload(formState.scaling);
  const originalScalingPayload: TagScalingPayload | null = originalTag.scaling
    ? {
        raw_min: originalTag.scaling.rawMin,
        raw_max: originalTag.scaling.rawMax,
        eng_min: originalTag.scaling.engMin,
        eng_max: originalTag.scaling.engMax,
        factor: originalTag.scaling.factor,
        offset: originalTag.scaling.offset
      }
    : null;

  if (!formState.useScaling && originalTag.scaling) {
    diff.scaling = "delete";
  }
  if (formState.useScaling && (!originalScalingPayload || hasPayloadChanges(nextScaling, originalScalingPayload))) {
    diff.scaling = nextScaling;
  }

  return diff;
};

const hasDiff = (diff: TagEditDiff) => Boolean(diff.meta || diff.params || diff.setpoints || diff.scaling);

export const TagDrawer = ({ mode, objectId, onClose, opened, tagId }: TagDrawerProps) => {
  const { items: dataTypes } = useDataTypesStore();
  const { devices } = useDevicesStore();
  const { items: units } = useUnitsStore();
  const { createTag, getTag, saveTagEdits } = useTagsStore();

  const [formState, setFormState] = useState<TagFormState>(INITIAL_FORM_STATE);
  const [initialTag, setInitialTag] = useState<Tag | null>(null);
  const [isLoadingTag, setIsLoadingTag] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const objectDevices = useMemo(() => devices.filter((device) => device.objectId === objectId), [devices, objectId]);

  const selectedDevice = useMemo(() => objectDevices.find((device) => device.id === formState.deviceId), [formState.deviceId, objectDevices]);

  const protocol = toProtocol(selectedDevice);

  useEffect(() => {
    if (!opened) {
      return;
    }

    setFormError(null);
    setFieldErrors({});

    if (mode === "create") {
      setInitialTag(null);
      setFormState(INITIAL_FORM_STATE);
      setIsLoadingTag(false);
      return;
    }

    if (!tagId) {
      setFormError("Не удалось определить тег для редактирования");
      return;
    }

    let isCancelled = false;
    setIsLoadingTag(true);

    void getTag(tagId)
      .then((tag) => {
        if (isCancelled) {
          return;
        }

        setInitialTag(tag);
        setFormState(mapTagToFormState(tag));
      })
      .catch((error) => {
        if (!isCancelled) {
          setFormError(toErrorMessage(error));
        }
      })
      .finally(() => {
        if (!isCancelled) {
          setIsLoadingTag(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [getTag, mode, opened, tagId]);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const validationErrors = validateForm(formState, protocol);
    setFieldErrors(validationErrors);
    setFormError(null);

    if (Object.keys(validationErrors).length > 0) {
      return;
    }

    if (!protocol || !formState.deviceId) {
      setFormError("Выберите устройство");
      return;
    }

    setIsSubmitting(true);

    try {
      if (mode === "create") {
        const paramsPayload = buildParamsPayload(formState, protocol);

        await createTag(formState.deviceId, {
          name: formState.name.trim(),
          description: formState.description.trim() || undefined,
          params: paramsPayload,
          setpoints: formState.useSetpoints ? toSetpointsPayload(formState.setpoints) : undefined,
          scaling: formState.useScaling ? toScalingPayload(formState.scaling) : undefined
        });
      }

      if (mode === "edit" && initialTag && tagId) {
        const diff = buildEditDiff(initialTag, formState, protocol);
        if (hasDiff(diff)) {
          await saveTagEdits(tagId, diff);
        }
      }

      onClose();
    } catch (error) {
      setFieldErrors(collectFieldErrors(error));
      setFormError(toErrorMessage(error));
      setIsSubmitting(false);
      return;
    }

    setIsSubmitting(false);
  };

  const drawerTitle = mode === "create" ? "Новый тег" : `Редактирование: ${initialTag?.name ?? ""}`;

  const dataTypeOptions = dataTypes.map((dataType: DataType) => ({
    value: String(dataType.id),
    label: dataType.name
  }));

  const unitOptions = buildUnitOptions(units);
  const deviceOptions = objectDevices.map((device) => ({
    value: device.id,
    label: device.name
  }));

  return (
    <Drawer opened={opened} onClose={onClose} title={drawerTitle} position="right" size={380}>
      {isLoadingTag ? (
        <Stack>
          <Skeleton h={36} radius="sm" />
          <Skeleton h={36} radius="sm" />
          <Skeleton h={120} radius="sm" />
        </Stack>
      ) : (
        <form onSubmit={onSubmit}>
          <Stack>
            <Text className={styles.sectionTitle}>Основное</Text>
            <TextInput
              label="Имя тега"
              value={formState.name}
              onChange={(event) => setFormState((state) => ({ ...state, name: event.currentTarget.value }))}
              error={fieldErrors.name}
              required
            />
            <Select
              label="Устройство"
              data={deviceOptions}
              value={formState.deviceId}
              onChange={(value) => setFormState((state) => ({ ...state, deviceId: value }))}
              error={fieldErrors.device_id}
              disabled={mode === "edit"}
              required
            />
            <Select
              label="Тип данных"
              data={dataTypeOptions}
              value={formState.dataTypeId}
              onChange={(value) => setFormState((state) => ({ ...state, dataTypeId: value }))}
              error={fieldErrors.data_type_id}
              required
            />
            <Select
              label="Единица"
              data={unitOptions}
              searchable
              clearable
              value={formState.unitId}
              onChange={(value) => setFormState((state) => ({ ...state, unitId: value }))}
              error={fieldErrors.unit_id}
            />
            <Textarea
              label="Описание"
              value={formState.description}
              onChange={(event) => setFormState((state) => ({ ...state, description: event.currentTarget.value }))}
              error={fieldErrors.description}
              minRows={2}
            />

            <Divider />

            <Text className={styles.sectionTitle}>Адрес</Text>
            <TagAddressForm
              protocol={protocol}
              address={formState.address}
              errors={fieldErrors}
              onChange={(patch) =>
                setFormState((state) => ({
                  ...state,
                  address: {
                    ...state.address,
                    ...patch
                  }
                }))
              }
            />

            <Divider />

            <Text className={styles.sectionTitle}>Scaling</Text>
            <Switch
              label="Использовать масштабирование"
              checked={formState.useScaling}
              onChange={(checked) => setFormState((state) => ({ ...state, useScaling: checked }))}
            />

            {formState.useScaling && (
              <TagScalingFields
                values={formState.scaling}
                errors={fieldErrors}
                onChange={(patch) =>
                  setFormState((state) => ({
                    ...state,
                    scaling: {
                      ...state.scaling,
                      ...patch
                    }
                  }))
                }
              />
            )}

            <Divider />

            <Text className={styles.sectionTitle}>Setpoints</Text>
            <Switch
              label="Использовать уставки"
              checked={formState.useSetpoints}
              onChange={(checked) => setFormState((state) => ({ ...state, useSetpoints: checked }))}
            />

            {formState.useSetpoints && (
              <TagSetpointsFields
                values={formState.setpoints}
                errors={fieldErrors}
                onChange={(patch) =>
                  setFormState((state) => ({
                    ...state,
                    setpoints: {
                      ...state.setpoints,
                      ...patch
                    }
                  }))
                }
              />
            )}

            {formError && (
              <Text c="red" size="sm">
                {formError}
              </Text>
            )}

            <Group className={styles.footerActions}>
              <Button type="button" variant="ghost" onClick={onClose} disabled={isSubmitting}>
                Отмена
              </Button>
              <Button type="submit" loading={isSubmitting}>
                Сохранить
              </Button>
            </Group>
          </Stack>
        </form>
      )}
    </Drawer>
  );
};
