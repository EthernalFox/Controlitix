import type { ModbusRegisterType } from "./types";

export const REGISTER_TYPE_LABELS: Record<ModbusRegisterType, string> = {
  holding: "Holding",
  input: "Input",
  coil: "Coil",
  discrete: "Discrete"
};

export const OID_PATTERN = /^[0-9]+(\.[0-9]+)+$/;
