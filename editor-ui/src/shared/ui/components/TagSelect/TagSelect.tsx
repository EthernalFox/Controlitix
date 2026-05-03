import type { SelectProps } from "@shared/ui/components/Select";
import { Select } from "@shared/ui/components/Select";

import type { AlarmBadgeState } from "../AlarmBadge";
import styles from "./TagSelect.module.css";

export type TagSelectMode = "local" | "remote";

export type TagSelectOption = {
  id: string;
  name: string;
  deviceName?: string;
  unit?: string;
  quality?: AlarmBadgeState;
};

type TagSelectDataItem = {
  value: string;
  label: string;
  deviceName?: string;
  unit?: string;
  quality?: AlarmBadgeState;
};

export type TagSelectProps = Omit<
  SelectProps,
  "data" | "value" | "onChange" | "renderOption" | "searchable" | "onSearchChange"
> & {
  value: string | null;
  onChange: (value: string | null) => void;
  options: TagSelectOption[];
  mode?: TagSelectMode;
  onSearchChange?: (value: string) => void;
};

const joinClassNames = (...classNames: Array<string | undefined>) =>
  classNames.filter(Boolean).join(" ");

export const TagSelect = ({
  value,
  onChange,
  options,
  mode = "local",
  onSearchChange,
  ...props
}: TagSelectProps) => {
  const data: TagSelectDataItem[] = options.map((option) => ({
    value: option.id,
    label: option.name,
    deviceName: option.deviceName,
    unit: option.unit,
    quality: option.quality
  }));

  return (
    <Select
      {...props}
      value={value}
      onChange={onChange}
      data={data}
      searchable
      nothingFoundMessage="Теги не найдены"
      filter={mode === "remote" ? ({ options: filteredOptions }) => filteredOptions : undefined}
      onSearchChange={onSearchChange}
      classNames={{
        dropdown: styles.dropdown,
        option: styles.option
      }}
      renderOption={({ option }) => {
        const typedOption = option as TagSelectDataItem;

        return (
          <div className={styles.row}>
            <div className={styles.texts}>
              <span className={styles.name}>{typedOption.label}</span>
              {typedOption.deviceName && (
                <span className={styles.device}>{typedOption.deviceName}</span>
              )}
            </div>
            {typedOption.unit && (
              <span
                className={joinClassNames(
                  styles.unit,
                  typedOption.quality ? styles[typedOption.quality] : undefined
                )}
              >
                {typedOption.unit}
              </span>
            )}
          </div>
        );
      }}
    />
  );
};
