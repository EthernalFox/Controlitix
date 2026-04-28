import { create } from "zustand";

import type {
  ConfigChangedMessage,
  ConnectionStatus,
  SnapshotMessage,
  ValueMessage
} from "@/shared/modules/realtime/types";

type LastTagMessage = ValueMessage | SnapshotMessage;

interface RealtimeState {
  status: ConnectionStatus;
  topicRefCount: Record<string, number>;
  activeTopicCount: number;
  lastMessageByTag: Record<string, LastTagMessage | undefined>;
  lastConfigChange: ConfigChangedMessage | null;

  setStatus: (status: ConnectionStatus) => void;
  setTopicRefCount: (topicRefCount: Map<string, number>) => void;
  upsertLastMessage: (message: LastTagMessage) => void;
  setLastConfigChange: (message: ConfigChangedMessage | null) => void;
  reset: (clearSubscriptions: boolean) => void;
}

const mapToObject = (value: Map<string, number>): Record<string, number> => {
  const result: Record<string, number> = {};

  value.forEach((count, topic) => {
    if (count > 0) {
      result[topic] = count;
    }
  });

  return result;
};

export const useRealtimeStore = create<RealtimeState>((set) => ({
  status: "closed",
  topicRefCount: {},
  activeTopicCount: 0,
  lastMessageByTag: {},
  lastConfigChange: null,

  setStatus: (status) => {
    set({ status });
  },

  setTopicRefCount: (topicRefCount) => {
    const next = mapToObject(topicRefCount);

    set({
      topicRefCount: next,
      activeTopicCount: Object.keys(next).length
    });
  },

  upsertLastMessage: (message) => {
    set((state) => ({
      lastMessageByTag: {
        ...state.lastMessageByTag,
        [message.tag_id]: message
      }
    }));
  },

  setLastConfigChange: (message) => {
    set({ lastConfigChange: message });
  },

  reset: (clearSubscriptions) => {
    set((state) => ({
      status: "closed",
      topicRefCount: clearSubscriptions ? {} : state.topicRefCount,
      activeTopicCount: clearSubscriptions ? 0 : state.activeTopicCount,
      lastConfigChange: null
    }));
  }
}));

export type { LastTagMessage };
