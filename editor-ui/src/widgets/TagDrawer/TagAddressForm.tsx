import type { DeviceProtocol, ModbusRegisterType } from "@entities/tags";
import { Stack, Text } from "@shared/ui";

import { ModbusAddressFields } from "./ModbusAddressFields";
import { SnmpAddressFields } from "./SnmpAddressFields";

interface AddressValue {
  registerType: ModbusRegisterType | null;
  address: number | null;
  oid: string;
}

interface TagAddressFormProps {
  protocol: DeviceProtocol | null;
  address: AddressValue;
  errors?: Record<string, string>;
  onChange: (patch: Partial<AddressValue>) => void;
}

export const TagAddressForm = ({
  address,
  errors,
  onChange,
  protocol
}: TagAddressFormProps) => {
  if (!protocol) {
    return null;
  }

  const isModbus = protocol === "modbus_rtu" || protocol === "modbus_tcp";

  return (
    <Stack gap="sm">
      <Text fw={600} size="sm">
        Адрес
      </Text>

      {isModbus ? (
        <ModbusAddressFields
          registerType={address.registerType}
          address={address.address}
          errors={errors}
          onChange={onChange}
        />
      ) : (
        <SnmpAddressFields
          oid={address.oid}
          errors={errors}
          onChange={(oid) => onChange({ oid })}
        />
      )}
    </Stack>
  );
};
