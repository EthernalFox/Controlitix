import { BaseApiClient } from "@shared/modules/api";

import { mimicsFormatter } from "./formatters";
import type {
  MimicApiModel,
  MimicId,
  MimicModel,
  MimicObjectId,
  MimicPayload
} from "./types";

export class MimicsApiClient extends BaseApiClient {
  private readonly objectDiagramsUrl = "/objects/:objectId/diagrams";

  private readonly diagramUrl = "/diagrams/:id";

  private readonly publishUrl = "/diagrams/:id/publish";

  public async create(
    objectId: MimicObjectId,
    payload: MimicPayload
  ): Promise<MimicModel> {
    const data = await this.post<MimicApiModel, MimicPayload>({
      url: this.objectDiagramsUrl,
      urlParams: { objectId },
      body: payload
    });

    return mimicsFormatter.toModel(data);
  }

  public async getById(id: MimicId): Promise<MimicModel> {
    const data = await this.get<MimicApiModel>({
      url: this.diagramUrl,
      urlParams: { id }
    });

    return mimicsFormatter.toModel(data);
  }

  public async update(
    id: MimicId,
    payload: MimicPayload
  ): Promise<MimicModel> {
    const data = await this.patch<MimicApiModel, MimicPayload>({
      url: this.diagramUrl,
      urlParams: { id },
      body: payload
    });

    return mimicsFormatter.toModel(data);
  }

  public async remove(id: MimicId): Promise<void> {
    await this.delete<void>({
      url: this.diagramUrl,
      urlParams: { id }
    });
  }

  public async publish(id: MimicId): Promise<MimicModel> {
    const data = await this.post<MimicApiModel>({
      url: this.publishUrl,
      urlParams: { id }
    });

    return mimicsFormatter.toModel(data);
  }
}

export const getMimicsApiClient = () =>
  MimicsApiClient.getClient<MimicsApiClient>(MimicsApiClient);
