import { FormEvent, useEffect, useMemo, useState } from "react";
import { NavLink as RouterNavLink, useParams } from "react-router";

import { useObjectsStore } from "@entities/objects";
import { ApiRequestError } from "@shared/api";
import { buildObjectPath } from "@shared/libs/router";
import {
  Button,
  Modal,
  NavLink,
  ScrollArea,
  Skeleton,
  Stack,
  Text,
  TextInput
} from "@shared/ui";

const EMPTY_DESCRIPTION = "Без описания";

const collectFieldErrors = (error: unknown) => {
  if (!(error instanceof ApiRequestError) || !error.payload.errors) {
    return {};
  }

  return error.payload.errors.reduce<Record<string, string>>((acc, fieldError) => {
    acc[fieldError.field] = fieldError.message;
    return acc;
  }, {});
};

export const ObjectsNavbar = () => {
  const { objectId } = useParams<{ objectId: string }>();
  const { objects, isLoading, fetchObjects, createObject } = useObjectsStore();

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
    setFormError(null);
    setFieldErrors({});
    setIsSubmitting(false);
  };

  const onCreateSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedName = name.trim();
    if (!trimmedName) {
      setFieldErrors({ name: "Введите название объекта" });
      return;
    }

    setIsSubmitting(true);
    setFormError(null);
    setFieldErrors({});

    try {
      await createObject({
        name: trimmedName,
        description: description.trim() || undefined
      });
      closeModal();
    } catch (error) {
      const nextFieldErrors = collectFieldErrors(error);
      setFieldErrors(nextFieldErrors);

      if (error instanceof ApiRequestError) {
        setFormError(error.payload.detail || error.payload.title);
      } else if (error instanceof Error) {
        setFormError(error.message);
      } else {
        setFormError("Не удалось создать объект");
      }

      setIsSubmitting(false);
    }
  };

  return (
    <>
      <Stack h="100%" gap="sm" p="sm">
        <TextInput
          placeholder="Поиск объекта..."
          value={search}
          onChange={(event) => setSearch(event.currentTarget.value)}
        />

        <ScrollArea style={{ flex: 1 }}>
          <Stack gap="xs">
            {isLoading && !objects.length ? (
              <>
                <Skeleton h={28} radius="md" />
                <Skeleton h={28} radius="md" />
                <Skeleton h={28} radius="md" />
              </>
            ) : (
              filteredObjects.map((object) => (
                <NavLink
                  key={object.id}
                  component={RouterNavLink}
                  to={buildObjectPath(object.id)}
                  active={object.id === objectId}
                  label={object.name}
                  description={object.description ?? EMPTY_DESCRIPTION}
                />
              ))
            )}
          </Stack>
        </ScrollArea>

        <Button onClick={() => setIsCreateOpened(true)}>Создать объект</Button>
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
              required
              error={fieldErrors.name}
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
            <Button variant="secondary" type="button" onClick={closeModal}>
              Отмена
            </Button>
          </Stack>
        </form>
      </Modal>
    </>
  );
};
