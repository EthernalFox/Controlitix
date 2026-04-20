import { FormEvent, useEffect, useState } from "react";

import { useDiagramsStore } from "@entities/diagrams";
import { ApiRequestError } from "@shared/api";
import { Button, Modal, Stack, Text, Textarea, TextInput } from "@shared/ui";

interface DiagramFormModalProps {
  opened: boolean;
  onClose: () => void;
  objectId: string;
}

const collectFieldErrors = (error: unknown) => {
  if (!(error instanceof ApiRequestError) || !error.payload.errors) {
    return {};
  }

  return error.payload.errors.reduce<Record<string, string>>((acc, fieldError) => {
    acc[fieldError.field] = fieldError.message;
    return acc;
  }, {});
};

export const DiagramFormModal = ({
  opened,
  onClose,
  objectId
}: DiagramFormModalProps) => {
  const { createDiagram } = useDiagramsStore();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (!opened) {
      setName("");
      setDescription("");
      setIsSubmitting(false);
      setFormError(null);
      setFieldErrors({});
    }
  }, [opened]);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedName = name.trim();
    const trimmedDescription = description.trim();

    setIsSubmitting(true);
    setFormError(null);
    setFieldErrors({});

    try {
      await createDiagram(objectId, {
        name: trimmedName || undefined,
        description: trimmedDescription || undefined
      });
      onClose();
    } catch (error) {
      setFieldErrors(collectFieldErrors(error));

      if (error instanceof ApiRequestError) {
        setFormError(error.payload.detail || error.payload.title);
      } else if (error instanceof Error) {
        setFormError(error.message);
      } else {
        setFormError("Не удалось создать мнемосхему");
      }

      setIsSubmitting(false);
    }
  };

  return (
    <Modal opened={opened} onClose={onClose} title="Создать мнемосхему">
      <form onSubmit={onSubmit}>
        <Stack>
          <TextInput
            label="Название"
            placeholder="Без названия"
            value={name}
            onChange={(event) => setName(event.currentTarget.value)}
            error={fieldErrors.name}
          />
          <Textarea
            label="Описание"
            value={description}
            onChange={(event) => setDescription(event.currentTarget.value)}
            error={fieldErrors.description}
            minRows={3}
          />
          {formError && (
            <Text c="red" size="sm">
              {formError}
            </Text>
          )}
          <Button type="submit" loading={isSubmitting}>
            Создать
          </Button>
          <Button type="button" variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Отмена
          </Button>
        </Stack>
      </form>
    </Modal>
  );
};

