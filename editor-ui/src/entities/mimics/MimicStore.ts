import { create } from "zustand";

import { MimicModel, MimicsStore } from "./types";

export const MimicStore = create<
  MimicsStore<MimicModel>
>(() => ({
  mimics: [],
  activeMimicId: null,
  isLoading: false,
  error: null,
  save: null as unknown as MimicsStore<MimicModel>["save"],
  read: null as unknown as MimicsStore<MimicModel>["read"],
  delete: null as unknown as MimicsStore<MimicModel>["delete"]
}));
