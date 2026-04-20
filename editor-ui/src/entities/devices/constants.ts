import type {
  DeviceSettings,
  DeviceTypeName,
  ModbusRtuSettings,
  ModbusTcpSettings,
  SnmpV1V2cSettings,
  SnmpV3Settings
} from "./types";

export const DEVICE_TYPE_LABELS: Record<DeviceTypeName, string> = {
  modbus_rtu: "Modbus RTU",
  modbus_tcp: "Modbus TCP",
  snmp_v1: "SNMP v1",
  snmp_v2c: "SNMP v2c",
  snmp_v3: "SNMP v3"
};

export const DEFAULT_MODBUS_TCP_SETTINGS: ModbusTcpSettings = {
  host: "",
  port: 502,
  slave_id: 1,
  timeout_ms: 3000
};

export const DEFAULT_MODBUS_RTU_SETTINGS: ModbusRtuSettings = {
  serial_port: "",
  baud_rate: 9600,
  data_bits: 8,
  stop_bits: 1,
  parity: "none",
  slave_id: 1,
  timeout_ms: 3000
};

export const DEFAULT_SNMP_V1_V2C_SETTINGS: SnmpV1V2cSettings = {
  host: "",
  port: 161,
  community: "public",
  timeout_ms: 5000
};

export const DEFAULT_SNMP_V3_SETTINGS: SnmpV3Settings = {
  host: "",
  port: 161,
  security_name: "",
  auth_protocol: "SHA",
  auth_password: "",
  priv_protocol: "AES",
  priv_password: "",
  timeout_ms: 5000
};

export const DEFAULT_SETTINGS: Record<DeviceTypeName, DeviceSettings> = {
  modbus_rtu: DEFAULT_MODBUS_RTU_SETTINGS,
  modbus_tcp: DEFAULT_MODBUS_TCP_SETTINGS,
  snmp_v1: DEFAULT_SNMP_V1_V2C_SETTINGS,
  snmp_v2c: DEFAULT_SNMP_V1_V2C_SETTINGS,
  snmp_v3: DEFAULT_SNMP_V3_SETTINGS
};

export const DEVICE_TYPE_OPTIONS = (
  Object.entries(DEVICE_TYPE_LABELS) as Array<[DeviceTypeName, string]>
).map(([value, label]) => ({ value, label }));

