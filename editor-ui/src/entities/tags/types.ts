export type DeviceProtocol =
  | "modbus_rtu"
  | "modbus_tcp"
  | "snmp_v1"
  | "snmp_v2c"
  | "snmp_v3";

export type ModbusRegisterType = "holding" | "input" | "coil" | "discrete";

export interface ModbusAddress {
  registerType: ModbusRegisterType;
  address: number;
}

export interface SnmpAddress {
  oid: string;
}

export type TagAddress = ModbusAddress | SnmpAddress;

export interface ModbusAddressPayload {
  register_type: ModbusRegisterType;
  address: number;
}

export interface SnmpAddressPayload {
  oid: string;
}

export type TagAddressPayload = ModbusAddressPayload | SnmpAddressPayload;

export interface TagParams {
  id: string;
  dataTypeId: number;
  unitId: number | null;
  address: TagAddress;
  createdAt: string;
  updatedAt: string;
}

export interface TagSetpoints {
  lolo: number | null;
  lo: number | null;
  hi: number | null;
  hihi: number | null;
}

export interface TagScaling {
  rawMin: number | null;
  rawMax: number | null;
  engMin: number | null;
  engMax: number | null;
  factor: number | null;
  offset: number | null;
}

export interface TagScalingPayload {
  raw_min: number | null;
  raw_max: number | null;
  eng_min: number | null;
  eng_max: number | null;
  factor: number | null;
  offset: number | null;
}

export interface Tag {
  id: string;
  deviceId: string;
  name: string;
  description: string | null;
  params: TagParams | null;
  setpoints: TagSetpoints | null;
  scaling: TagScaling | null;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTagPayload {
  name: string;
  description?: string;
  params: {
    data_type_id: number;
    unit_id?: number | null;
    address: TagAddressPayload;
  };
  setpoints?: TagSetpoints;
  scaling?: TagScalingPayload;
}

export interface UpdateTagPayload {
  name?: string;
  description?: string;
}

export interface UpdateTagParamsPayload {
  data_type_id: number;
  unit_id: number | null;
  address: TagAddressPayload;
}

export interface TagsListFilters {
  object_id?: string;
  device_id?: string;
  data_type_id?: number;
  unit_id?: number;
  search?: string;
  offset?: number;
  limit?: number;
}

export type TagEditDiff = {
  meta?: UpdateTagPayload;
  params?: UpdateTagParamsPayload;
  setpoints?: TagSetpoints | "delete";
  scaling?: TagScalingPayload | "delete";
};
