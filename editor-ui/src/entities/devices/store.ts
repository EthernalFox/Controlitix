import { create } from "zustand";

import { ApiRequestError } from "@shared/api";

import { deviceTypesApi, devicesApi } from "./api";
import { DEFAULT_SETTINGS, DEVICE_TYPE_LABELS } from "./constants";
import type {
  CreateDevicePayload,
  Device,
  DeviceSettings,
  DeviceType,
  DeviceTypeName,
  DeviceWithParams,
  UpdateDevicePayload
} from "./types";

interface DeviceFilters {
  name?: string;
  type_id?: number;
}

interface DevicesState {
  devices: Device[];
  deviceTypes: DeviceType[];
  isLoading: boolean;
  error: string | null;
  fetchDevices: (objectId: string, filters?: DeviceFilters) => Promise<void>;
  fetchDeviceTypes: () => Promise<void>;
  getDevice: (id: string) => Promise<DeviceWithParams>;
  createDevice: (
    objectId: string,
    payload: CreateDevicePayload
  ) => Promise<DeviceWithParams>;
  updateDevice: (id: string, payload: UpdateDevicePayload) => Promise<void>;
  deleteDevice: (id: string) => Promise<void>;
  reset: () => void;
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const isDeviceTypeName = (value: string): value is DeviceTypeName =>
  Object.hasOwn(DEVICE_TYPE_LABELS, value);

const toErrorMessage = (error: unknown): string => {
  if (error instanceof ApiRequestError) {
    return error.payload.detail || error.payload.title;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "Request failed";
};

const readNumber = (
  source: Record<string, unknown>,
  snakeKey: string,
  camelKey: string,
  fallback: number
) => {
  const value = source[snakeKey] ?? source[camelKey];
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }

  return fallback;
};

const readString = (
  source: Record<string, unknown>,
  snakeKey: string,
  camelKey: string,
  fallback: string
) => {
  const value = source[snakeKey] ?? source[camelKey];
  return typeof value === "string" ? value : fallback;
};

const cloneDefaultSettings = (typeName: DeviceTypeName): DeviceSettings => ({
  ...DEFAULT_SETTINGS[typeName]
});

const normalizeSettings = (
  rawSettings: unknown,
  typeName: DeviceTypeName
): DeviceSettings => {
  const source = isRecord(rawSettings) ? rawSettings : {};

  if (typeName === "modbus_rtu") {
    return {
      serial_port: readString(source, "serial_port", "serialPort", ""),
      baud_rate: readNumber(source, "baud_rate", "baudRate", 9600),
      data_bits: readNumber(source, "data_bits", "dataBits", 8),
      stop_bits: readNumber(source, "stop_bits", "stopBits", 1),
      parity: (() => {
        const value = readString(source, "parity", "parity", "none");
        return value === "even" || value === "odd" ? value : "none";
      })(),
      slave_id: readNumber(source, "slave_id", "slaveId", 1),
      timeout_ms: readNumber(source, "timeout_ms", "timeoutMs", 3000)
    };
  }

  if (typeName === "snmp_v1" || typeName === "snmp_v2c") {
    return {
      host: readString(source, "host", "host", ""),
      port: readNumber(source, "port", "port", 161),
      community: readString(source, "community", "community", "public"),
      timeout_ms: readNumber(source, "timeout_ms", "timeoutMs", 5000)
    };
  }

  if (typeName === "snmp_v3") {
    return {
      host: readString(source, "host", "host", ""),
      port: readNumber(source, "port", "port", 161),
      security_name: readString(source, "security_name", "securityName", ""),
      auth_protocol: (() => {
        const value = readString(source, "auth_protocol", "authProtocol", "SHA");
        return value === "MD5" ? "MD5" : "SHA";
      })(),
      auth_password: readString(source, "auth_password", "authPassword", ""),
      priv_protocol: (() => {
        const value = readString(source, "priv_protocol", "privProtocol", "AES");
        return value === "DES" ? "DES" : "AES";
      })(),
      priv_password: readString(source, "priv_password", "privPassword", ""),
      timeout_ms: readNumber(source, "timeout_ms", "timeoutMs", 5000)
    };
  }

  return {
    host: readString(source, "host", "host", ""),
    port: readNumber(source, "port", "port", 502),
    slave_id: readNumber(source, "slave_id", "slaveId", 1),
    timeout_ms: readNumber(source, "timeout_ms", "timeoutMs", 3000)
  };
};

const resolveTypeName = (
  device: Device | DeviceWithParams | Record<string, unknown>,
  deviceTypes: DeviceType[]
): DeviceTypeName => {
  const rawTypeFromLegacyField =
    isRecord(device) && typeof device.type === "string" ? device.type : null;

  const rawTypeName =
    typeof device.typeName === "string"
      ? device.typeName
      : typeof rawTypeFromLegacyField === "string"
        ? rawTypeFromLegacyField
        : null;

  if (rawTypeName && isDeviceTypeName(rawTypeName)) {
    return rawTypeName;
  }

  const rawTypeId = typeof device.typeId === "number" ? device.typeId : null;
  if (rawTypeId !== null) {
    const foundType = deviceTypes.find((deviceType) => deviceType.id === rawTypeId);
    if (foundType) {
      return foundType.name;
    }
  }

  return "modbus_tcp";
};

const normalizeDevice = (
  rawDevice: Device | DeviceWithParams | Record<string, unknown>,
  deviceTypes: DeviceType[]
): Device => {
  const typeName = resolveTypeName(rawDevice, deviceTypes);
  const fallbackTypeId =
    deviceTypes.find((deviceType) => deviceType.name === typeName)?.id ?? 0;
  const typeId =
    typeof rawDevice.typeId === "number" ? rawDevice.typeId : fallbackTypeId;

  return {
    id: String(rawDevice.id ?? ""),
    objectId:
      rawDevice.objectId === null || typeof rawDevice.objectId === "string"
        ? rawDevice.objectId
        : null,
    typeId,
    typeName,
    name: typeof rawDevice.name === "string" ? rawDevice.name : "",
    description:
      rawDevice.description === null || typeof rawDevice.description === "string"
        ? rawDevice.description
        : null,
    tagsCount: typeof rawDevice.tagsCount === "number" ? rawDevice.tagsCount : undefined,
    createdAt:
      typeof rawDevice.createdAt === "string" ? rawDevice.createdAt : new Date().toISOString(),
    updatedAt:
      typeof rawDevice.updatedAt === "string" ? rawDevice.updatedAt : new Date().toISOString()
  };
};

const normalizeDeviceWithParams = (
  rawDevice: DeviceWithParams | Record<string, unknown>,
  deviceTypes: DeviceType[]
): DeviceWithParams => {
  const normalizedDevice = normalizeDevice(rawDevice, deviceTypes);
  return {
    ...normalizedDevice,
    settings: normalizeSettings(rawDevice.settings, normalizedDevice.typeName)
  };
};

const upsertDevice = (devices: Device[], device: Device) => {
  const existingIndex = devices.findIndex((currentDevice) => currentDevice.id === device.id);
  if (existingIndex === -1) {
    return [device, ...devices];
  }

  return devices.map((currentDevice) =>
    currentDevice.id === device.id ? device : currentDevice
  );
};

const normalizeDeviceTypes = (deviceTypes: DeviceType[]): DeviceType[] =>
  deviceTypes.filter((deviceType) => isDeviceTypeName(deviceType.name));

export const useDevicesStore = create<DevicesState>((set, get) => ({
  devices: [],
  deviceTypes: [],
  isLoading: false,
  error: null,

  fetchDevices: async (objectId, filters) => {
    set({ isLoading: true, error: null });

    try {
      const response = await devicesApi.listByObject(objectId, filters);
      const normalizedDevices = response.items.map((device) =>
        normalizeDevice(device, get().deviceTypes)
      );
      set({ devices: normalizedDevices, isLoading: false });
    } catch (error) {
      set({ error: toErrorMessage(error), isLoading: false });
      throw error;
    }
  },

  fetchDeviceTypes: async () => {
    if (get().deviceTypes.length > 0) {
      return;
    }

    set({ error: null });
    try {
      const deviceTypes = normalizeDeviceTypes(await deviceTypesApi.list());
      set((state) => ({
        deviceTypes,
        devices: state.devices.map((device) => normalizeDevice(device, deviceTypes))
      }));
    } catch (error) {
      set({ error: toErrorMessage(error) });
      throw error;
    }
  },

  getDevice: async (id) => {
    const device = await devicesApi.get(id);
    return normalizeDeviceWithParams(device, get().deviceTypes);
  },

  createDevice: async (objectId, payload) => {
    set({ error: null });
    const createdDevice = normalizeDeviceWithParams(
      await devicesApi.create(objectId, payload),
      get().deviceTypes
    );

    set((state) => ({
      devices: upsertDevice(state.devices, createdDevice)
    }));

    return createdDevice;
  },

  updateDevice: async (id, payload) => {
    set({ error: null });
    const updatedDevice = normalizeDeviceWithParams(
      await devicesApi.update(id, payload),
      get().deviceTypes
    );

    set((state) => ({
      devices: upsertDevice(state.devices, updatedDevice)
    }));
  },

  deleteDevice: async (id) => {
    set({ error: null });
    await devicesApi.delete(id);

    set((state) => ({
      devices: state.devices.filter((device) => device.id !== id)
    }));
  },

  reset: () => {
    set({
      devices: [],
      isLoading: false,
      error: null
    });
  }
}));

export { cloneDefaultSettings };
