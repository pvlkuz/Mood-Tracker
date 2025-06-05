import { parseISO, format as fnsFormat } from "date-fns";

export const formatDate = (isoDate, pattern = "yyyy-MM-dd") => {
  return fnsFormat(parseISO(isoDate), pattern);
};
