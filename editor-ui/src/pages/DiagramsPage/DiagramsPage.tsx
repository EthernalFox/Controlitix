import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router";

import type { Diagram } from "@entities/diagrams";
import { useDiagramsStore } from "@entities/diagrams";
import { ApiRequestError } from "@shared/api";
import { buildDiagramEditorPath } from "@shared/libs/router";
import {
  Button,
  Card,
  Group,
  Modal,
  SimpleGrid,
  Skeleton,
  Stack,
  Text
} from "@shared/ui";
import { DiagramCard } from "@widgets/DiagramCard";
import { DiagramFormModal } from "@widgets/DiagramFormModal";

const FALLBACK_NAME = "Без названия";

const toErrorMessage = (error: unknown) => {
  if (error instanceof ApiRequestError) {
    return error.payload.detail || error.payload.title;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Не удалось выполнить запрос";
};

export default function DiagramsPage() {
  const navigate = useNavigate();
  const { objectId } = useParams<{ objectId: string }>();
  const { diagrams, isLoading, error, fetchDiagrams, deleteDiagram, reset } =
    useDiagramsStore();

  const [isCreateOpened, setIsCreateOpened] = useState(false);
  const [diagramToDelete, setDiagramToDelete] = useState<Diagram | null>(null);
  const [isDeleteSubmitting, setIsDeleteSubmitting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  useEffect(() => {
    if (!objectId) {
      return;
    }

    reset();
    void fetchDiagrams(objectId).catch(() => undefined);

    return () => {
      reset();
    };
  }, [fetchDiagrams, objectId, reset]);

  const onEdit = (diagramId: string) => {
    if (!objectId) {
      return;
    }

    navigate(buildDiagramEditorPath(objectId, diagramId));
  };

  const closeDeleteModal = () => {
    if (isDeleteSubmitting) {
      return;
    }

    setDiagramToDelete(null);
    setDeleteError(null);
  };

  const onDeleteConfirm = async () => {
    if (!diagramToDelete) {
      return;
    }

    setIsDeleteSubmitting(true);
    setDeleteError(null);

    try {
      await deleteDiagram(diagramToDelete.id);
      closeDeleteModal();
    } catch (deleteErrorValue) {
      setDeleteError(toErrorMessage(deleteErrorValue));
      setIsDeleteSubmitting(false);
    }
  };

  useEffect(() => {
    if (!diagramToDelete) {
      setIsDeleteSubmitting(false);
    }
  }, [diagramToDelete]);

  if (!objectId) {
    return (
      <Card withBorder p="md">
        <Text c="red">Не удалось определить объект для загрузки мнемосхем.</Text>
      </Card>
    );
  }

  return (
    <>
      <Stack gap="md">
        <Group justify="space-between" align="center">
          <Text size="xl" fw={600}>
            Мнемосхемы
          </Text>
          <Button onClick={() => setIsCreateOpened(true)}>Создать мнемосхему</Button>
        </Group>

        {error && !isLoading && (
          <Card withBorder p="md">
            <Stack gap="xs">
              <Text c="red">{error}</Text>
              <Button variant="secondary" onClick={() => void fetchDiagrams(objectId)}>
                Повторить
              </Button>
            </Stack>
          </Card>
        )}

        {isLoading && !diagrams.length && (
          <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }}>
            <Card withBorder p="md">
              <Stack gap="sm">
                <Skeleton h={72} radius="md" />
                <Skeleton h={18} radius="sm" />
                <Skeleton h={16} radius="sm" w="70%" />
                <Skeleton h={16} radius="sm" w="45%" />
              </Stack>
            </Card>
            <Card withBorder p="md">
              <Stack gap="sm">
                <Skeleton h={72} radius="md" />
                <Skeleton h={18} radius="sm" />
                <Skeleton h={16} radius="sm" w="70%" />
                <Skeleton h={16} radius="sm" w="45%" />
              </Stack>
            </Card>
            <Card withBorder p="md">
              <Stack gap="sm">
                <Skeleton h={72} radius="md" />
                <Skeleton h={18} radius="sm" />
                <Skeleton h={16} radius="sm" w="70%" />
                <Skeleton h={16} radius="sm" w="45%" />
              </Stack>
            </Card>
          </SimpleGrid>
        )}

        {!isLoading && !error && !diagrams.length && (
          <Card withBorder p="md">
            <Stack gap="xs" align="center">
              <Text size="xl">[]</Text>
              <Text ta="center">У этого объекта нет мнемосхем</Text>
              <Button onClick={() => setIsCreateOpened(true)}>
                Создать первую мнемосхему
              </Button>
            </Stack>
          </Card>
        )}

        {!isLoading && diagrams.length > 0 && (
          <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }}>
            {diagrams.map((diagram) => (
              <DiagramCard
                key={diagram.id}
                diagram={diagram}
                onEdit={onEdit}
                onDelete={setDiagramToDelete}
              />
            ))}
          </SimpleGrid>
        )}
      </Stack>

      <DiagramFormModal
        opened={isCreateOpened}
        onClose={() => setIsCreateOpened(false)}
        objectId={objectId}
      />

      <Modal
        opened={diagramToDelete !== null}
        onClose={closeDeleteModal}
        title={`Удалить мнемосхему "${diagramToDelete?.name ?? FALLBACK_NAME}"?`}
      >
        <Stack>
          <Text>Все фигуры будут удалены.</Text>
          {deleteError && (
            <Text c="red" size="sm">
              {deleteError}
            </Text>
          )}
          <Button
            variant="danger"
            onClick={() => void onDeleteConfirm()}
            loading={isDeleteSubmitting}
          >
            Удалить
          </Button>
          <Button
            variant="secondary"
            onClick={closeDeleteModal}
            disabled={isDeleteSubmitting}
          >
            Отмена
          </Button>
        </Stack>
      </Modal>
    </>
  );
}

