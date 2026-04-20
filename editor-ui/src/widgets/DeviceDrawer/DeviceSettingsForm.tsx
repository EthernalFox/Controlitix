import type {
  DeviceSettings,
  DeviceTypeName,
  ModbusRtuSettings,
  ModbusTcpSettings,
  SnmpV1V2cSettings,
  SnmpV3Settings
} from "@entities/devices";

import { ModbusRtuFields } from "./ModbusRtuFields";
import { ModbusTcpFields } from "./ModbusTcpFields";
import { SnmpV1V2cFields } from "./SnmpV1V2cFields";
import { SnmpV3Fields } from "./SnmpV3Fields";

interface DeviceSettingsFormProps {
  typeName: DeviceTypeName | null;
  settings: DeviceSettings | null;
  onChange: (settings: DeviceSettings) => void;
  errors?: Record<string, string>;
}

export const DeviceSettingsForm = ({
  typeName,
  settings,
  onChange,
  errors
}: DeviceSettingsFormProps) => {
  if (!typeName || !settings) {
    return null;
  }

  switch (typeName) {
    case "modbus_tcp":
      return (
        <ModbusTcpFields
          settings={settings as ModbusTcpSettings}
          onChange={(nextSettings) => onChange(nextSettings)}
          errors={errors}
        />
      );
    case "modbus_rtu":
      return (
        <ModbusRtuFields
          settings={settings as ModbusRtuSettings}
          onChange={(nextSettings) => onChange(nextSettings)}
          errors={errors}
        />
      );
    case "snmp_v1":
    case "snmp_v2c":
      return (
        <SnmpV1V2cFields
          settings={settings as SnmpV1V2cSettings}
          onChange={(nextSettings) => onChange(nextSettings)}
          errors={errors}
        />
      );
    case "snmp_v3":
      return (
        <SnmpV3Fields
          settings={settings as SnmpV3Settings}
          onChange={(nextSettings) => onChange(nextSettings)}
          errors={errors}
        />
      );
    default:
      return null;
  }
};

