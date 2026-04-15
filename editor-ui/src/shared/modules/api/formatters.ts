export interface Formatter<TRaw, TModel> {
  toModel: (raw: TRaw) => TModel;
  toCollection: (raw: TRaw[]) => TModel[];
}

export const createFormatter = <TRaw, TModel>(
  mapper: (raw: TRaw) => TModel
): Formatter<TRaw, TModel> => ({
  toModel: mapper,
  toCollection: (raw) => raw.map(mapper)
});
