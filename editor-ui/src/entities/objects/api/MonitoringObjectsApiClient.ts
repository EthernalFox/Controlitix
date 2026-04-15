import { BaseApiClient } from "@shared/modules/api";

import { monitoringObjectsFormatter } from "./formatters";
import type {
  MonitoringObjectApiModel,
  MonitoringObjectId,
  MonitoringObjectModel,
  MonitoringObjectPayload
} from "./types";

export class MonitoringObjectsApiClient extends BaseApiClient {
  private readonly objectsUrl = "/objects";

  private readonly objectUrl = "/objects/:id";

  public async create(
    payload: MonitoringObjectPayload
  ): Promise<MonitoringObjectModel> {
    const data = await this.post<
      MonitoringObjectApiModel,
      MonitoringObjectPayload
    >({
      url: this.objectsUrl,
      body: payload
    });

    return monitoringObjectsFormatter.toModel(data);
  }

  public async getById(id: MonitoringObjectId): Promise<MonitoringObjectModel> {
    const data = await this.get<MonitoringObjectApiModel>({
      url: this.objectUrl,
      urlParams: { id }
    });

    return monitoringObjectsFormatter.toModel(data);
  }

  public async update(
    id: MonitoringObjectId,
    payload: MonitoringObjectPayload
  ): Promise<MonitoringObjectModel> {
    const data = await this.patch<
      MonitoringObjectApiModel,
      MonitoringObjectPayload
    >({
      url: this.objectUrl,
      urlParams: { id },
      body: payload
    });

    return monitoringObjectsFormatter.toModel(data);
  }

  public async remove(id: MonitoringObjectId): Promise<void> {
    await this.delete<void>({
      url: this.objectUrl,
      urlParams: { id }
    });
  }
}

export const getMonitoringObjectsApiClient = () =>
  MonitoringObjectsApiClient.getClient<MonitoringObjectsApiClient>(
    MonitoringObjectsApiClient
  );
