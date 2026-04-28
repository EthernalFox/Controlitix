import dayjs from "dayjs";

export const formatTimestamp = (
  value: string | number | Date,
  format = "YYYY-MM-DD HH:mm:ss"
) => {
  return dayjs(value).format(format);
};
