import type { DiagramFigure } from "@/features/diagram/model/diagramApi";

export const computeBoundTags = (figures: DiagramFigure[]): string[] => {
  const uniqueTagIds = new Set<string>();

  figures.forEach((figure) => {
    const tagId = figure.tagId?.trim();
    if (!tagId) {
      return;
    }

    uniqueTagIds.add(tagId);
  });

  return Array.from(uniqueTagIds);
};
