import { FormEvent, useEffect, useMemo, useState } from "react";
import { Outlet, useLocation, useNavigate, useParams } from "react-router";

import { useObjectsStore } from "@entities/objects";
import { ApiRequestError } from "@shared/api";
import {
  buildDevicesPath,
  buildDiagramsPath,
  buildTagsPath,
  routePaths
} from "@shared/libs/router";
import {
  ActionIcon,
  Button,
  Group,
  Modal,
  Stack,
  Tabs,
  Text,
  Tooltip,
  TextInput
} from "@shared/ui";

const parseActiveTab = (pathname: string) => {
  const segments = pathname.split("/").filter(Boolean);
  const lastSegment = segments[segments.length - 1];
  if (lastSegment === "tags" || lastSegment === "diagrams") {
    return lastSegment;
  }

  return "devices";
};

const collectFieldErrors = (error: unknown) => {
  if (!(error instanceof ApiRequestError) || !error.payload.errors) {
    return {};
  }

  return error.payload.errors.reduce<Record<string, string>>((acc, fieldError) => {
    acc[fieldError.field] = fieldError.message;
    return acc;
  }, {});
};

export default function ObjectDetailPage() {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const { objectId } = useParams<{ objectId: string }>();
  const { objects, getObject, updateObject, deleteObject } = useObjectsStore();

  const [isObjectLoading, setIsObjectLoading] = useState(true);
  const [objectError, setObjectError] = useState<string | null>(null);
  const [isEditOpened, setIsEditOpened] = useState(false);
  const [isDeleteOpened, setIsDeleteOpened] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const object = useMemo(
    () => objects.find((currentObject) => currentObject.id === objectId),
    [objects, objectId]
  );

  useEffect(() => {
    if (!objectId) {
      setObjectError("Объект не найден");
      setIsObjectLoading(false);
      return;
    }

    let isCancelled = false;
    setIsObjectLoading(true);
    setObjectError(null);

    void getObject(objectId)
      .catch((error) => {
        if (isCancelled) {
          return;
        }

        if (error instanceof ApiRequestError) {
          setObjectError(error.payload.detail || error.payload.title);
          return;
        }

        if (error instanceof Error) {
          setObjectError(error.message);
          return;
        }

        setObjectError("Не удалось загрузить объект");
      })
      .finally(() => {
        if (!isCancelled) {
          setIsObjectLoading(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [getObject, objectId]);

  useEffect(() => {
    if (!object || !isEditOpened) {
      return;
    }

    setName(object.name);
    setDescription(object.description ?? "");
  }, [isEditOpened, object]);

  const closeEditModal = () => {
    setIsEditOpened(false);
    setIsSubmitting(false);
    setFieldErrors({});
    setFormError(null);
  };

  const onEditSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!objectId) {
      return;
    }

    const trimmedName = name.trim();
    if (!trimmedName) {
      setFieldErrors({ name: "Введите название объекта" });
      return;
    }

    setIsSubmitting(true);
    setFieldErrors({});
    setFormError(null);

    try {
      await updateObject(objectId, {
        name: trimmedName,
        description: description.trim() || null
      });
      closeEditModal();
    } catch (submitError) {
      setFieldErrors(collectFieldErrors(submitError));

      if (submitError instanceof ApiRequestError) {
        setFormError(submitError.payload.detail || submitError.payload.title);
      } else if (submitError instanceof Error) {
        setFormError(submitError.message);
      } else {
        setFormError("Не удалось сохранить изменения");
      }

      setIsSubmitting(false);
    }
  };

  const onDeleteConfirm = async () => {
    if (!objectId) {
      return;
    }

    setIsSubmitting(true);
    try {
      await deleteObject(objectId);
      navigate(routePaths.objects);
    } catch (deleteError) {
      if (deleteError instanceof ApiRequestError) {
        setFormError(deleteError.payload.detail || deleteError.payload.title);
      } else if (deleteError instanceof Error) {
        setFormError(deleteError.message);
      } else {
        setFormError("Не удалось удалить объект");
      }
      setIsSubmitting(false);
    }
  };

  if (isObjectLoading) {
    return (
      <Stack p="md">
        <Text c="dimmed">Загрузка объекта...</Text>
      </Stack>
    );
  }

  if (!object) {
    return (
      <Stack p="md" gap="sm">
        <Text c="red">{objectError ?? "Объект не найден"}</Text>
        <Button variant="secondary" onClick={() => navigate(routePaths.objects)}>
          К списку объектов
        </Button>
      </Stack>
    );
  }

  return (
    <>
      <Stack p="md" gap="md">
        <Stack gap={4}>
          <Group justify="space-between" align="start">
            <Stack gap={4}>
              <Text size="xl" fw={600}>
                {object.name}
              </Text>
              <Text c="dimmed">{object.description ?? "Без описания"}</Text>
            </Stack>
            <Group gap="xs">
              <Tooltip label="Редактировать">
                <ActionIcon
                  variant="light"
                  onClick={() => setIsEditOpened(true)}
                  aria-label="Edit object"
                >
                  E
                </ActionIcon>
              </Tooltip>
              <Tooltip label="Удалить">
                <ActionIcon
                  variant="light"
                  color="red"
                  onClick={() => setIsDeleteOpened(true)}
                  aria-label="Delete object"
                >
                  D
                </ActionIcon>
              </Tooltip>
            </Group>
          </Group>
        </Stack>

        <Tabs value={parseActiveTab(pathname)}>
          <Tabs.List>
            <Tabs.Tab
              value="devices"
              onClick={() => navigate(buildDevicesPath(object.id))}
            >
              Устройства
            </Tabs.Tab>
            <Tabs.Tab
              value="tags"
              onClick={() => navigate(buildTagsPath(object.id))}
            >
              Теги
            </Tabs.Tab>
            <Tabs.Tab
              value="diagrams"
              onClick={() => navigate(buildDiagramsPath(object.id))}
            >
              Мнемосхемы
            </Tabs.Tab>
          </Tabs.List>
        </Tabs>

        <Outlet />
      </Stack>

      <Modal opened={isEditOpened} onClose={closeEditModal} title="Редактировать объект">
        <form onSubmit={onEditSubmit}>
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
              Сохранить
            </Button>
            <Button type="button" variant="secondary" onClick={closeEditModal}>
              Отмена
            </Button>
          </Stack>
        </form>
      </Modal>

      <Modal opened={isDeleteOpened} onClose={() => setIsDeleteOpened(false)} title="Удаление объекта">
        <Stack>
          <Text>
            Удалить объект "{object.name}"? Все устройства, теги и мнемосхемы будут
            удалены.
          </Text>
          {formError && (
            <Text c="red" size="sm">
              {formError}
            </Text>
          )}
          <Button variant="danger" onClick={() => void onDeleteConfirm()} loading={isSubmitting}>
            Удалить
          </Button>
          <Button
            variant="secondary"
            onClick={() => setIsDeleteOpened(false)}
            disabled={isSubmitting}
          >
            Отмена
          </Button>
        </Stack>
      </Modal>
    </>
  );
}
