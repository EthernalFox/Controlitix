import { useEffect, useMemo, useRef } from "react";

import { realtimeClient } from "./client";
import type { ErrorMessage, RealtimeTopic } from "./types";

export interface RealtimeTagValue {
  tag_id: string;
  ts: string;
  v: number | null;
  q:
    | "ok"
    | "hi"
    | "hihi"
    | "uncertain"
    | "bad"
    | "comm_loss"
    | "offline"
    | "acknowledged";
  isSnapshot: boolean;
}

interface TopicSubscriptionOptions {
  onLimitExceeded?: (message: ErrorMessage) => void;
}

const toTagTopic = (tagId: string): RealtimeTopic => `tag:${tagId}`;

const normalizeTopics = (topics: string[]): string[] => {
  const unique = new Set<string>();

  topics.forEach((topic) => {
    const normalizedTopic = topic.trim();
    if (!normalizedTopic) {
      return;
    }

    unique.add(normalizedTopic);
  });

  return Array.from(unique);
};

const isTagTopic = (topic: string): boolean => {
  return topic.startsWith("tag:");
};

export const useTopicSubscription = (
  topics: string[],
  onValue: (message: RealtimeTagValue) => void,
  options?: TopicSubscriptionOptions
): void => {
  const callbackRef = useRef(onValue);
  const optionsRef = useRef(options);

  useEffect(() => {
    callbackRef.current = onValue;
  }, [onValue]);

  useEffect(() => {
    optionsRef.current = options;
  }, [options]);

  const topicsSignature = useMemo(() => {
    return normalizeTopics(topics).join(",");
  }, [topics]);

  useEffect(() => {
    if (!topicsSignature) {
      return;
    }

    const normalizedTopics = topicsSignature.split(",");
    realtimeClient.subscribe(normalizedTopics);

    return () => {
      realtimeClient.unsubscribe(normalizedTopics);
    };
  }, [topicsSignature]);

  useEffect(() => {
    if (!topicsSignature) {
      return;
    }

    const topicSet = new Set(topicsSignature.split(","));
    const hasGroupedTopic = Array.from(topicSet).some((topic) => !isTagTopic(topic));

    const shouldForwardByTagID = (tagID: string): boolean => {
      if (hasGroupedTopic) {
        return true;
      }

      return topicSet.has(toTagTopic(tagID));
    };

    const unsubscribeValue = realtimeClient.on("value", (message) => {
      if (!shouldForwardByTagID(message.tag_id)) {
        return;
      }

      callbackRef.current({
        tag_id: message.tag_id,
        ts: message.ts,
        v: message.v,
        q: message.q,
        isSnapshot: false
      });
    });

    const unsubscribeSnapshot = realtimeClient.on("snapshot", (message) => {
      if (!shouldForwardByTagID(message.tag_id)) {
        return;
      }

      callbackRef.current({
        tag_id: message.tag_id,
        ts: message.ts,
        v: message.v,
        q: message.q,
        isSnapshot: true
      });
    });

    const unsubscribeError = realtimeClient.on("error", (message) => {
      if (message.code !== "limit-exceeded") {
        return;
      }

      if (!message.topics || message.topics.length === 0) {
        return;
      }

      const hasIntersection = message.topics.some((topic) => topicSet.has(topic));
      if (!hasIntersection) {
        return;
      }

      optionsRef.current?.onLimitExceeded?.(message);
    });

    return () => {
      unsubscribeValue();
      unsubscribeSnapshot();
      unsubscribeError();
    };
  }, [topicsSignature]);
};

