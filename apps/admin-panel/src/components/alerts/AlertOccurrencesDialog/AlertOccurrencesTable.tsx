import { AlertOccurrencesSearch } from "@components/alerts/AlertOccurrencesDialog/AlertOccurrencesSearch";
import {
  buildAlertOccurrenceSearchText,
  formatAlertOccurrenceValue,
} from "@components/alerts/AlertOccurrencesDialog/helpers";
import { PaginationControl } from "@components/pagination";
import { useClientPagination } from "@domain/pagination/hooks/useClientPagination";
import { usePaginationState } from "@domain/pagination/hooks/usePaginationState";
import type { AlertOccurrenceItemDTO } from "@dto/alerts/alerts";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Typography from "@mui/material/Typography";
import { type ReactElement, useDeferredValue, useMemo, useState } from "react";
import type { AlertOccurrencesTableProps } from "./types";

const defaultOccurrencesPageSize = 25;

/**
 * Searchable and paginated table view for non-numeric alert occurrences.
 */
export const AlertOccurrencesTable = ({
  occurrences,
}: AlertOccurrencesTableProps): ReactElement => {
  const tableState = useAlertOccurrencesTableState(occurrences);

  return (
    <Stack spacing={2}>
      {renderAlertOccurrencesTableHeader(tableState.filteredCount, occurrences.length)}
      <AlertOccurrencesSearch search={tableState.search} setSearch={tableState.setSearch} />
      <PaginationControl
        pagination={{
          page: tableState.paginatedOccurrences.page,
          setPage: tableState.paginatedOccurrences.setPage,
          totalCount: tableState.paginatedOccurrences.totalCount,
          totalPages: tableState.paginatedOccurrences.totalPages,
        }}
        summary={buildAlertOccurrencesPaginationSummary(
          tableState.filteredCount,
          tableState.paginatedOccurrences.totalPages,
        )}
        testIdPrefix="alert-occurrences-top"
      />
      <TableContainer component={Paper} data-testid="alert-occurrences-table">
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Timestamp</TableCell>
              <TableCell>Value</TableCell>
              <TableCell>Value type</TableCell>
              <TableCell>Event ID</TableCell>
              <TableCell>Event record ID</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {renderAlertOccurrencesTableRows(tableState.paginatedOccurrences.records)}
          </TableBody>
        </Table>
      </TableContainer>
      <PaginationControl
        pagination={{
          page: tableState.paginatedOccurrences.page,
          setPage: tableState.paginatedOccurrences.setPage,
          totalCount: tableState.paginatedOccurrences.totalCount,
          totalPages: tableState.paginatedOccurrences.totalPages,
        }}
        summary={buildAlertOccurrencesPageSummary(
          tableState.paginatedOccurrences.page,
          tableState.paginatedOccurrences.totalPages,
        )}
        testIdPrefix="alert-occurrences-bottom"
      />
    </Stack>
  );
};

type AlertOccurrencesTableState = {
  filteredCount: number;
  paginatedOccurrences: ReturnType<typeof useClientPagination<AlertOccurrenceItemDTO>>;
  search: string;
  setSearch: (value: string) => void;
};

/**
 * Owns the local search and client-side pagination state for the occurrences table.
 */
const useAlertOccurrencesTableState = (
  occurrences: AlertOccurrenceItemDTO[],
): AlertOccurrencesTableState => {
  const [search, setSearch] = useState("");
  const deferredSearch = useDeferredValue(search);
  const pagination = usePaginationState({ initialPageSize: defaultOccurrencesPageSize });
  const filteredOccurrences = useMemo(() => {
    return filterAlertOccurrences(occurrences, deferredSearch);
  }, [deferredSearch, occurrences]);
  const paginatedOccurrences = useClientPagination({
    pagination,
    records: filteredOccurrences,
  });

  return {
    filteredCount: filteredOccurrences.length,
    paginatedOccurrences,
    search,
    setSearch: (value) => {
      pagination.setPage(1);
      setSearch(value);
    },
  };
};

type AlertOccurrencesTableRowProps = {
  /**
   * One occurrence row rendered by the table variant.
   */
  occurrence: AlertOccurrenceItemDTO;
};

/**
 * One occurrence table row showing timestamp and serialized value fields.
 */
const AlertOccurrencesTableRow = ({ occurrence }: AlertOccurrencesTableRowProps): ReactElement => {
  return (
    <TableRow
      data-test-class="alert-occurrences-table-row"
      data-test-name={`${occurrence.EventID}:${occurrence.EventRecID}:${occurrence.RecID}`}
      hover
    >
      <TableCell>{occurrence.Timestamp}</TableCell>
      <TableCell>{formatAlertOccurrenceValue(occurrence)}</TableCell>
      <TableCell>{occurrence.ValueType}</TableCell>
      <TableCell>{occurrence.EventID}</TableCell>
      <TableCell>{occurrence.EventRecID}</TableCell>
    </TableRow>
  );
};

/**
 * Renders the title and match summary for the occurrences table.
 */
const renderAlertOccurrencesTableHeader = (
  filteredCount: number,
  totalCount: number,
): ReactElement => {
  return (
    <Stack
      alignItems={{ md: "center", xs: "flex-start" }}
      direction={{ md: "row", xs: "column" }}
      justifyContent="space-between"
      spacing={2}
    >
      <Stack spacing={0.5}>
        <Typography data-testid="alert-occurrences-table-title" variant="h6">
          Occurrence history
        </Typography>
        <Typography color="text.secondary" variant="body2">
          {filteredCount} of {totalCount} occurrences match the current search.
        </Typography>
      </Stack>
    </Stack>
  );
};

/**
 * Filters the occurrence rows against the current free-text search term.
 */
const filterAlertOccurrences = (
  occurrences: AlertOccurrenceItemDTO[],
  search: string,
): AlertOccurrenceItemDTO[] => {
  const normalizedSearch = search.trim().toLocaleLowerCase();

  if (normalizedSearch.length === 0) {
    return occurrences;
  }

  return occurrences.filter((occurrence) => {
    return buildAlertOccurrenceSearchText(occurrence)
      .toLocaleLowerCase()
      .includes(normalizedSearch);
  });
};

/**
 * Renders the current page of occurrence rows or the matching empty state.
 */
const renderAlertOccurrencesTableRows = (
  occurrences: AlertOccurrenceItemDTO[],
): ReactElement | ReactElement[] => {
  if (occurrences.length === 0) {
    return (
      <TableRow data-testid="alert-occurrences-table-empty-row">
        <TableCell colSpan={5}>
          <Typography color="text.secondary" variant="body2">
            No occurrences match the current search.
          </Typography>
        </TableCell>
      </TableRow>
    );
  }

  return occurrences.map((occurrence) => {
    return (
      <AlertOccurrencesTableRow
        key={`${occurrence.EventID}:${occurrence.EventRecID}:${occurrence.RecID}`}
        occurrence={occurrence}
      />
    );
  });
};

/**
 * Builds the top pagination summary for the occurrences table.
 */
const buildAlertOccurrencesPaginationSummary = (
  filteredCount: number,
  totalPages: number,
): string => {
  return `${filteredCount} matching occurrences across ${Math.max(totalPages, 1)} pages`;
};

/**
 * Builds the bottom pagination summary for the occurrences table.
 */
const buildAlertOccurrencesPageSummary = (page: number, totalPages: number): string => {
  return `Page ${page} of ${Math.max(totalPages, 1)}`;
};
