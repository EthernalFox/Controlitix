export type DeviceTypeName =
  | "modbus_rtu"
  | "modbus_tcp"
  | "snmp_v1"
  | "snmp_v2c"
  | "snmp_v3";

export interface DeviceType {
  id: number;
  name: DeviceTypeName;
}

export interface Device {
  id: string;
  objectId: string | null;
  typeId: number;
  typeName: DeviceTypeName;
  name: string;
  description: string | null;
  tagsCount?: number;
  createdAt: string;
  updatedAt: string;
}

export interface ModbusTcpSettings {
  host: string;
  port: number;
  slave_id: number;
  timeout_ms: number;
}

export interface ModbusRtuSettings {
  serial_port: string;
  baud_rate: number;
  data_bits: number;
  stop_bits: number;
  parity: "none" | "even" | "odd";
  slave_id: number;
  timeout_ms: number;
}

export interface SnmpV1V2cSettings {
  host: string;
  port: number;
  community: string;
  timeout_ms: number;
  retry_count: number;
}

export interface SnmpV3Settings {
  host: string;
  port: number;
  security_name: string;
  security_level: "noAuthNoPriv" | "authNoPriv" | "authPriv";
  auth_protocol: "MD5" | "SHA";
  auth_password: string;
  priv_protocol: "DES" | "AES";
  priv_password: string;
  timeout_ms: number;
  retry_count: number;
}

export type DeviceSettings = ModbusTcpSettings | ModbusRtuSettings | SnmpV1V2cSettings | SnmpV3Settings;

export interface DeviceWithParams extends Device {
  settings: DeviceSettings;
}

export interface CreateDevicePayload {
  name: string;
  description?: string;
  type_id: number;
  settings: DeviceSettings;
}

export interface UpdateDevicePayload {
  name?: string;
  description?: string;
  settings?: DeviceSettings;
}
