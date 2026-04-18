import { FormEvent, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router";

import { useObjectsStore } from "@entities/objects";
import { ApiRequestError } from "@shared/api";
import { buildObjectPath } from "@shared/libs/router";
import {
  Button,
  Card,
  Group,
  Modal,
  Skeleton,
  Stack,
  Text,
  TextInput
} from "@shared/ui";

const EMPTY_DESCRIPTION = "Без описания";

const formatDate = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "short",
    timeStyle: "short"
  }).format(new Date(value));

const collectFieldErrors = (error: unknown) => {
  if (!(error instanceof ApiRequestError) || !error.payload.errors) {
    return {};
  }

  return error.payload.errors.reduce<Record<string, string>>((acc, fieldError) => {
    acc[fieldError.field] = fieldError.message;
    return acc;
  }, {});
};

export default function ObjectsPage() {
  const navigate = useNavigate();
  const { objects, isLoading, error, fetchObjects, createObject } = useObjectsStore();

  const [search, setSearch] = useState("");
  const [isCreateOpened, setIsCreateOpened] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (!objects.length) {
      void fetchObjects();
    }
  }, [fetchObjects, objects.length]);

  const filteredObjects = useMemo(() => {
    const normalizedSearch = search.trim().toLowerCase();
    if (!normalizedSearch) {
      return objects;
    }

    return objects.filter((object) =>
      object.name.toLowerCase().includes(normalizedSearch)
    );
  }, [objects, search]);

  const closeModal = () => {
    setIsCreateOpened(false);
    setName("");
    setDescription("");
    setIsSubmitting(false);
    setFieldErrors({});
    setFormError(null);
  };

  const onCreateSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedName = name.trim();
    if (!trimmedName) {
      setFieldErrors({ name: "Введите название объекта" });
      return;
    }

    setIsSubmitting(true);
    setFieldErrors({});
    setFormError(null);

    try {
      await createObject({
        name: trimmedName,
        description: description.trim() || undefined
      });
      closeModal();
    } catch (submitError) {
      const nextFieldErrors = collectFieldErrors(submitError);
      setFieldErrors(nextFieldErrors);

      if (submitError instanceof ApiRequestError) {
        setFormError(submitError.payload.detail || submitError.payload.title);
      } else if (submitError instanceof Error) {
        setFormError(submitError.message);
      } else {
        setFormError("Не удалось создать объект");
      }

      setIsSubmitting(false);
    }
  };

  return (
    <>
      <Stack p="md">
        <Group justify="space-between" align="center">
          <Text size="xl" fw={600}>
            Объекты мониторинга
          </Text>
          <Button onClick={() => setIsCreateOpened(true)}>Создать объект</Button>
        </Group>

        <TextInput
          placeholder="Поиск по названию..."
          value={search}
          onChange={(event) => setSearch(event.currentTarget.value)}
        />

        {error && !isLoading && (
          <Card withBorder p="md">
            <Stack gap="xs">
              <Text c="red">{error}</Text>
              <Button variant="secondary" onClick={() => void fetchObjects()}>
                Повторить
              </Button>
            </Stack>
          </Card>
        )}

        {isLoading && !objects.length && (
          <Stack>
            <Skeleton h={96} radius="md" />
            <Skeleton h={96} radius="md" />
            <Skeleton h={96} radius="md" />
          </Stack>
        )}

        {!isLoading && !error && !filteredObjects.length && (
          <Card withBorder p="md">
            <Stack gap="xs">
              <Text>Нет объектов мониторинга</Text>
              <Button onClick={() => setIsCreateOpened(true)}>
                Создать первый объект
              </Button>
            </Stack>
          </Card>
        )}

        {!isLoading && filteredObjects.length > 0 && (
          <Stack>
            {filteredObjects.map((object) => (
              <div
                key={object.id}
                role="button"
                tabIndex={0}
                style={{ cursor: "pointer" }}
                onClick={() => navigate(buildObjectPath(object.id))}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    navigate(buildObjectPath(object.id));
                  }
                }}
              >
                <Card withBorder p="md">
                  <Stack gap={4}>
                    <Text fw={600}>{object.name}</Text>
                    <Text c="dimmed" size="sm">
                      {object.description ?? EMPTY_DESCRIPTION}
                    </Text>
                    <Text c="dimmed" size="xs">
                      Обновлено: {formatDate(object.updatedAt)}
                    </Text>
                  </Stack>
                </Card>
              </div>
            ))}
          </Stack>
        )}
      </Stack>

      <Modal
        opened={isCreateOpened}
        onClose={closeModal}
        title="Создать объект мониторинга"
      >
        <form onSubmit={onCreateSubmit}>
          <Stack>
            <TextInput
              label="Название"
              value={name}
              onChange={(event) => setName(event.currentTarget.value)}
              error={fieldErrors.name}
              required
            />
            <TextInput
              label="Описание"
              value={description}
              onChange={(event) => setDescription(event.currentTarget.value)}
              error={fieldErrors.description}
            />
            {formError && (
              <Text c="red" size="sm">
                {formError}
              </Text>
            )}
            <Button type="submit" loading={isSubmitting}>
              Создать
            </Button>
            <Button type="button" variant="secondary" onClick={closeModal}>
              Отмена
            </Button>
          </Stack>
        </form>
      </Modal>
    </>
  );
}
