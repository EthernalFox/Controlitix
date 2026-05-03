import { IconPencil, IconTrash } from "@tabler/icons-react";

import type { DataType } from "@entities/data-types";
import type { Device } from "@entities/devices";
import { REGISTER_TYPE_LABELS, type Tag } from "@entities/tags";
import type { Unit } from "@entities/units";
import { ActionIcon, Group, Table, Text, Tooltip } from "@shared/ui";

import styles from "./TagsTable.module.css";

interface TagsTableProps {
  tags: Tag[];
  devices: Device[];
  dataTypes: DataType[];
  units: Unit[];
  onEdit: (id: string) => void;
  onDelete: (tag: Tag) => void;
}

const toAddressLabel = (tag: Tag) => {
  const address = tag.params?.address;
  if (!address) {
    return "—";
  }

  if ("registerType" in address) {
    return `${address.registerType}/${address.address}`;
  }

  if ("oid" in address) {
    return address.oid;
  }

  return "—";
};

const toUnitLabel = (unit: Unit | undefined) => unit?.symbol ?? "—";

const toRegisterBadge = (tag: Tag) => {
  const address = tag.params?.address;
  if (!address || !("registerType" in address)) {
    return null;
  }

  return REGISTER_TYPE_LABELS[address.registerType] ?? address.registerType;
};

export const TagsTable = ({ dataTypes, devices, onDelete, onEdit, tags, units }: TagsTableProps) => {
  return (
    <Table
      className={styles.table}
      data={{
        head: ["Статус", "Название", "Устройство", "Тип данных", "Единица", "Адрес", ""],
        body: tags.map((tag) => {
          const device = devices.find((item) => item.id === tag.deviceId);
          const dataType = dataTypes.find((item) => item.id === tag.params?.dataTypeId);
          const unit = units.find((item) => item.id === tag.params?.unitId);
          const registerBadge = toRegisterBadge(tag);

          return [
            <span key={`${tag.id}-dot`} className={styles.statusDot} />,
            tag.name,
            device?.name ?? "—",
            dataType?.name ?? "—",
            toUnitLabel(unit),
            <Group key={`${tag.id}-address`} gap="xs" wrap="nowrap">
              <Text size="sm" c="dimmed" truncate className={styles.mono}>
                {toAddressLabel(tag)}
              </Text>
              {registerBadge && (
                <Text size="xs" c="dimmed">
                  ({registerBadge})
                </Text>
              )}
            </Group>,
            <Group key={`${tag.id}-actions`} gap="xs" justify="center" wrap="nowrap">
              <Tooltip label="Редактировать">
                <ActionIcon variant="light" aria-label="Edit tag" onClick={() => onEdit(tag.id)}>
                  <IconPencil size={14} />
                </ActionIcon>
              </Tooltip>
              <Tooltip label="Удалить">
                <ActionIcon variant="light" color="red" aria-label="Delete tag" onClick={() => onDelete(tag)}>
                  <IconTrash size={14} />
                </ActionIcon>
              </Tooltip>
            </Group>
          ];
        })
      }}
    />
  );
};
