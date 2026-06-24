import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";

type AlertOccurrencesSearchProps = {
  /**
   * Current free-text search term.
   */
  search: string;
  /**
   * Updates the free-text search term.
   */
  setSearch: (value: string) => void;
};

/**
 * Search input used by the table variant of the occurrences dialog.
 */
export const AlertOccurrencesSearch = ({
  search,
  setSearch,
}: AlertOccurrencesSearchProps): ReactElement => {
  return (
    <TextField
      data-testid="alert-occurrences-search-control"
      label="Search occurrences"
      name="alertOccurrencesSearch"
      onChange={(event) => {
        setSearch(event.target.value);
      }}
      size="small"
      sx={{ maxWidth: "100%", width: 260 }}
      value={search}
    />
  );
};
