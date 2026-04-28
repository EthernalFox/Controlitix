export const formatNumber = (
  value: number,
  fractionDigits = 2,
  minimumIntegerDigits = 1
) => {
  return value.toLocaleString(undefined, {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
    minimumIntegerDigits,
    useGrouping: false
  });
};
