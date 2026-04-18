export interface MimicModel {
  id: string;
  objectId: string;
  name: string | null;
  description: string | null;
  createdAt: string;
  updatedAt: string;
  publishedAt: string | null;
}

export type SaveMimicPayload<TMimic extends MimicModel = MimicModel> = Partial<
  Omit<TMimic, "createdAt" | "updatedAt">
> & {
  id?: TMimic["id"];
};

export interface MimicsStore<TMimic extends MimicModel = MimicModel> {
  mimics: TMimic[];
  activeMimicId: TMimic["id"] | null;
  isLoading: boolean;
  error: string | null;
  save: (payload: SaveMimicPayload<TMimic>) => Promise<TMimic>;
  read: (objectId?: TMimic["objectId"]) => Promise<TMimic[]>;
  delete: (id: TMimic["id"]) => Promise<void>;
}
