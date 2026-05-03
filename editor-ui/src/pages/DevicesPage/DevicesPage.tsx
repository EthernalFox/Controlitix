import { IconPencil, IconPlus, IconTrash } from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";
import { useParams } from "react-router";

import { DEVICE_TYPE_LABELS, useDevicesStore, type Device } from "@entities/devices";
import { ApiRequestError } from "@shared/api";
import { ActionIcon, Button, Card, Group, Modal, Select, Stack, Table, Text, TextInput, Tooltip } from "@shared/ui";
import { DeviceDrawer } from "@widgets/DeviceDrawer";
import { EmptyState } from "@widgets/EmptyState";
import { ListSkeleton } from "@widgets/ListSkeleton";
import { PageToolbar } from "@widgets/PageToolbar";

import styles from "./DevicesPage.module.css";

const toErrorMessage = (error: unknown) => {
  if (error instanceof ApiRequestError) {
    return error.payload.detail || error.payload.title;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Не удалось выполнить запрос";
};

const getTypeLabel = (typeName: string) =>
  DEVICE_TYPE_LABELS[typeName as keyof typeof DEVICE_TYPE_LABELS] ?? typeName;

const formatDate = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "short",
    timeStyle: "short"
  }).format(new Date(value));

export default function DevicesPage() {
  const { objectId } = useParams<{ objectId: string }>();
  const { deleteDevice, deviceTypes, devices, error, fetchDevices, fetchDeviceTypes, isLoading, reset } =
    useDevicesStore();

  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState<string | null>(null);
  const [drawerMode, setDrawerMode] = useState<"closed" | "create" | "edit">("closed");
  const [editDeviceId, setEditDeviceId] = useState<string | null>(null);
  const [deviceToDelete, setDeviceToDelete] = useState<Device | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [isDeleteSubmitting, setIsDeleteSubmitting] = useState(false);

  useEffect(() => {
    if (!objectId) {
      return;
    }

    reset();
    void Promise.all([fetchDeviceTypes(), fetchDevices(objectId)]).catch(() => undefined);

    return () => {
      reset();
    };
  }, [fetchDevices, fetchDeviceTypes, objectId, reset]);

  const filteredDevices = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();

    return devices.filter((device) => {
      if (normalizedSearch && !device.name.toLowerCase().includes(normalizedSearch)) {
        return false;
      }

      if (typeFilter && device.typeName !== typeFilter) {
        return false;
      }

      return true;
    });
  }, [devices, search, typeFilter]);

  const closeDrawer = () => {
    setDrawerMode("closed");
    setEditDeviceId(null);
  };

  const closeDeleteModal = () => {
    if (isDeleteSubmitting) {
      return;
    }

    setDeviceToDelete(null);
    setDeleteError(null);
  };

  const onDeleteConfirm = async () => {
    if (!deviceToDelete) {
      return;
    }

    setIsDeleteSubmitting(true);
    setDeleteError(null);

    try {
      await deleteDevice(deviceToDelete.id);
      closeDeleteModal();
    } catch (errorValue) {
      setDeleteError(toErrorMessage(errorValue));
      setIsDeleteSubmitting(false);
    }
  };

  useEffect(() => {
    if (!deviceToDelete) {
      setIsDeleteSubmitting(false);
    }
  }, [deviceToDelete]);

  if (!objectId) {
    return (
      <Card withBorder p="md">
        <Text c="red">Не удалось определить объект для загрузки устройств.</Text>
      </Card>
    );
  }

  return (
    <>
      <Stack gap="md">
        <PageToolbar
          title="Устройства"
          search={
            <TextInput placeholder="Поиск..." value={search} onChange={(event) => setSearch(event.currentTarget.value)} />
          }
          tabs={
            <Select
              placeholder="Все типы"
              value={typeFilter}
              onChange={setTypeFilter}
              clearable
              data={deviceTypes.map((deviceType) => ({
                value: deviceType.name,
                label: getTypeLabel(deviceType.name)
              }))}
              w={220}
            />
          }
          primaryAction={
            <Button onClick={() => setDrawerMode("create")}>
              <IconPlus size={14} />
              Добавить устройство
            </Button>
          }
        />

        {error && !isLoading && (
          <Card withBorder p="md">
            <Stack gap="xs">
              <Text c="red">{error}</Text>
              <Button variant="secondary" onClick={() => void fetchDevices(objectId)}>
                Повторить
              </Button>
            </Stack>
          </Card>
        )}

        {isLoading && <ListSkeleton cols={3} />}

        {!isLoading && !error && devices.length === 0 && (
          <Card variant="flat" withBorder p="lg">
            <EmptyState
              icon={<Text size="xl">⚙</Text>}
              title="Нет устройств"
              description="Добавьте первое устройство — Modbus или SNMP"
              action={
                <Button onClick={() => setDrawerMode("create")}>
                  <IconPlus size={14} />
                  Добавить устройство
                </Button>
              }
            />
          </Card>
        )}

        {!isLoading && !error && devices.length > 0 && filteredDevices.length === 0 && (
          <Card variant="flat" withBorder p="md">
            <Text>По заданным фильтрам устройства не найдены.</Text>
          </Card>
        )}

        {!isLoading && filteredDevices.length > 0 && (
          <Card variant="glass" p={0} className={styles.tableCard}>
            <Table
              className={styles.table}
              data={{
                head: ["Статус", "Имя", "Тип", "Адрес", "Последний опрос", ""],
                body: filteredDevices.map((device) => [
                  <span key={`${device.id}-dot`} className={styles.statusDot} />,
                  device.name,
                  getTypeLabel(device.typeName),
                  <span key={`${device.id}-addr`} className={styles.mono}>
                    —
                  </span>,
                  formatDate(device.updatedAt),
                  <Group key={`${device.id}-actions`} gap="xs" justify="center">
                    <Tooltip label="Редактировать">
                      <ActionIcon
                        aria-label="Edit device"
                        variant="light"
                        onClick={() => {
                          setEditDeviceId(device.id);
                          setDrawerMode("edit");
                        }}
                      >
                        <IconPencil size={14} />
                      </ActionIcon>
                    </Tooltip>
                    <Tooltip label="Удалить">
                      <ActionIcon
                        aria-label="Delete device"
                        variant="light"
                        color="red"
                        onClick={() => setDeviceToDelete(device)}
                      >
                        <IconTrash size={14} />
                      </ActionIcon>
                    </Tooltip>
                  </Group>
                ])
              }}
            />
          </Card>
        )}
      </Stack>

      <DeviceDrawer
        opened={drawerMode !== "closed"}
        mode={drawerMode === "edit" ? "edit" : "create"}
        objectId={objectId}
        deviceId={editDeviceId}
        onClose={closeDrawer}
      />

      <Modal
        opened={deviceToDelete !== null}
        onClose={closeDeleteModal}
        title={`Удалить устройство "${deviceToDelete?.name ?? ""}"?`}
      >
        <Stack>
          <Text>
            Все теги этого устройства
            {typeof deviceToDelete?.tagsCount === "number" ? ` (${deviceToDelete.tagsCount} шт.)` : ""} будут
            удалены.
          </Text>
          {deleteError && (
            <Text c="red" size="sm">
              {deleteError}
            </Text>
          )}
          <Button variant="danger" onClick={() => void onDeleteConfirm()} loading={isDeleteSubmitting}>
            Удалить
          </Button>
          <Button variant="secondary" onClick={closeDeleteModal} disabled={isDeleteSubmitting}>
            Отмена
          </Button>
        </Stack>
      </Modal>
    </>
  );
}
