import { useMemo } from "react";

import { useTopicSubscription } from "./useTopicSubscription";

export type { RealtimeTagValue } from "./useTopicSubscription";

const normalizeTagIDs = (tagIDs: string[]): string[] => {
  const unique = new Set<string>();

  tagIDs.forEach((tagID) => {
    const normalizedTagID = tagID.trim();
    if (!normalizedTagID) {
      return;
    }

    unique.add(normalizedTagID);
  });

  return Array.from(unique);
};

export const useTagSubscription = (
  tagIDs: string[],
  onValue: Parameters<typeof useTopicSubscription>[1]
): void => {
  const topics = useMemo(() => {
    return normalizeTagIDs(tagIDs).map((tagID) => `tag:${tagID}`);
  }, [tagIDs]);

  useTopicSubscription(topics, onValue);
};

