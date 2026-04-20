import { TextInput } from "@shared/ui";

interface SnmpAddressFieldsProps {
  oid: string;
  errors?: Record<string, string>;
  onChange: (oid: string) => void;
}

export const SnmpAddressFields = ({ errors, oid, onChange }: SnmpAddressFieldsProps) => {
  return (
    <TextInput
      label="OID"
      placeholder="1.3.6.1.2.1.1.1.0"
      value={oid}
      onChange={(event) => onChange(event.currentTarget.value)}
      error={errors?.["address.oid"]}
      required
    />
  );
};
