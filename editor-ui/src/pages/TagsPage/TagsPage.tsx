import { IconPlus } from "@tabler/icons-react";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router";

import { useDataTypesStore } from "@entities/data-types";
import { useDevicesStore, type Device } from "@entities/devices";
import { useTagsStore, type Tag } from "@entities/tags";
import { useUnitsStore, type Unit } from "@entities/units";
import { buildDevicesPath } from "@shared/libs/router";
import { Button, Card, Group, Modal, Pagination, Select, Stack, Text, TextInput } from "@shared/ui";
import { EmptyState } from "@widgets/EmptyState";
import { ListSkeleton } from "@widgets/ListSkeleton";
import { PageToolbar } from "@widgets/PageToolbar";
import { TagDrawer } from "@widgets/TagDrawer";
import { TagsTable } from "@widgets/TagsTable";

const buildUnitFilterOptions = (units: Unit[]) => {
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

export default function TagsPage() {
  const navigate = useNavigate();
  const { objectId } = useParams<{ objectId: string }>();

  const { deleteTag, error, fetchTags, filters, isLoading, reset, setFilters, tags, total } = useTagsStore();
  const { devices, fetchDevices, reset: resetDevices } = useDevicesStore();
  const { fetch: fetchDataTypes, items: dataTypes } = useDataTypesStore();
  const { fetch: fetchUnits, items: units } = useUnitsStore();

  const [drawerMode, setDrawerMode] = useState<"closed" | "create" | "edit">("closed");
  const [editTagId, setEditTagId] = useState<string | null>(null);
  const [tagToDelete, setTagToDelete] = useState<Tag | null>(null);
  const [isDeleteSubmitting, setIsDeleteSubmitting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  useEffect(() => {
    if (!objectId) {
      return;
    }

    reset();
    resetDevices();

    setFilters({
      object_id: objectId,
      device_id: undefined,
      data_type_id: undefined,
      unit_id: undefined,
      search: undefined,
      offset: 0,
      limit: 50
    });

    void Promise.all([fetchDevices(objectId), fetchDataTypes(), fetchUnits()]).catch(() => undefined);

    return () => {
      reset();
      resetDevices();
    };
  }, [fetchDataTypes, fetchDevices, fetchUnits, objectId, reset, resetDevices, setFilters]);

  useEffect(() => {
    if (!objectId || filters.object_id !== objectId) {
      return;
    }

    void fetchTags().catch(() => undefined);
  }, [fetchTags, filters, objectId]);

  const limit = filters.limit ?? 50;
  const currentPage = Math.floor((filters.offset ?? 0) / limit) + 1;
  const pagesTotal = Math.max(1, Math.ceil(total / limit));

  const deviceOptions = useMemo(
    () =>
      devices.map((device: Device) => ({
        value: device.id,
        label: device.name
      })),
    [devices]
  );

  const dataTypeOptions = useMemo(
    () =>
      dataTypes.map((dataType) => ({
        value: String(dataType.id),
        label: dataType.name
      })),
    [dataTypes]
  );

  const unitOptions = useMemo(() => buildUnitFilterOptions(units), [units]);

  const closeDeleteModal = () => {
    if (isDeleteSubmitting) {
      return;
    }

    setTagToDelete(null);
    setDeleteError(null);
  };

  const onDeleteConfirm = async () => {
    if (!tagToDelete) {
      return;
    }

    setDeleteError(null);
    setIsDeleteSubmitting(true);

    try {
      await deleteTag(tagToDelete.id);
      closeDeleteModal();
    } catch (errorValue) {
      if (errorValue instanceof Error) {
        setDeleteError(errorValue.message);
      } else {
        setDeleteError("Не удалось удалить тег");
      }
      setIsDeleteSubmitting(false);
    }
  };

  useEffect(() => {
    if (!tagToDelete) {
      setIsDeleteSubmitting(false);
    }
  }, [tagToDelete]);

  if (!objectId) {
    return (
      <Card withBorder p="md">
        <Text c="red">Не удалось определить объект для загрузки тегов.</Text>
      </Card>
    );
  }

  return (
    <>
      <Stack gap="md">
        <PageToolbar
          title="Теги"
          search={
            <TextInput
              placeholder="Поиск..."
              value={filters.search ?? ""}
              onChange={(event) => setFilters({ search: event.currentTarget.value.trim() || undefined })}
            />
          }
          tabs={
            <Group gap="sm" wrap="wrap">
              <Select
                placeholder="Устройство"
                data={deviceOptions}
                clearable
                value={filters.device_id ?? null}
                onChange={(value) => setFilters({ device_id: value ?? undefined })}
                w={220}
              />
              <Select
                placeholder="Тип данных"
                data={dataTypeOptions}
                clearable
                value={filters.data_type_id ? String(filters.data_type_id) : null}
                onChange={(value) => setFilters({ data_type_id: value ? Number(value) : undefined })}
                w={180}
              />
              <Select
                placeholder="Единица"
                data={unitOptions}
                searchable
                clearable
                value={filters.unit_id ? String(filters.unit_id) : null}
                onChange={(value) => setFilters({ unit_id: value ? Number(value) : undefined })}
                w={220}
              />
            </Group>
          }
          primaryAction={
            <Button disabled={devices.length === 0} onClick={() => setDrawerMode("create")}> 
              <IconPlus size={14} />
              Добавить тег
            </Button>
          }
        />

        {error && !isLoading && (
          <Card withBorder p="md">
            <Stack gap="xs">
              <Text c="red">{error}</Text>
              <Button variant="secondary" onClick={() => void fetchTags()}>
                Повторить
              </Button>
            </Stack>
          </Card>
        )}

        {isLoading && <ListSkeleton cols={3} />}

        {!isLoading && devices.length === 0 && (
          <Card variant="flat" withBorder p="lg">
            <EmptyState
              icon={<Text size="xl">#</Text>}
              title="Сначала добавьте устройство"
              description="Теги создаются только внутри устройств"
              action={
                <Button variant="secondary" onClick={() => navigate(buildDevicesPath(objectId))}>
                  Перейти к устройствам
                </Button>
              }
            />
          </Card>
        )}

        {!isLoading && devices.length > 0 && tags.length === 0 && !error && (
          <Card variant="flat" withBorder p="lg">
            <EmptyState
              icon={<Text size="xl">#</Text>}
              title="Нет тегов"
              description="Добавьте первый тег для выбранного объекта"
              action={
                <Button onClick={() => setDrawerMode("create")}> 
                  <IconPlus size={14} />
                  Добавить тег
                </Button>
              }
            />
          </Card>
        )}

        {!isLoading && tags.length > 0 && (
          <Card variant="glass" p={0}>
            <TagsTable
              tags={tags}
              devices={devices}
              dataTypes={dataTypes}
              units={units}
              onEdit={(id) => {
                setEditTagId(id);
                setDrawerMode("edit");
              }}
              onDelete={setTagToDelete}
            />
          </Card>
        )}

        {!isLoading && devices.length > 0 && total > 0 && (
          <Group justify="end">
            <Pagination value={currentPage} total={pagesTotal} onChange={(page) => setFilters({ offset: (page - 1) * limit })} />
          </Group>
        )}
      </Stack>

      <TagDrawer
        opened={drawerMode !== "closed"}
        mode={drawerMode === "edit" ? "edit" : "create"}
        objectId={objectId}
        tagId={editTagId}
        onClose={() => {
          setDrawerMode("closed");
          setEditTagId(null);
        }}
      />

      <Modal opened={tagToDelete !== null} onClose={closeDeleteModal} title={`Удалить тег "${tagToDelete?.name ?? ""}"?`}>
        <Stack>
          <Text>Это действие нельзя отменить.</Text>
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
