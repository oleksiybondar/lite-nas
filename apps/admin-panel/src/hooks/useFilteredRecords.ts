import { useDeferredValue, useMemo, useState } from "react";

/**
 * Filter keys supported by one filtered-records hook instance.
 */
export type FilteredRecordFilterState<TFilterKey extends string> = Record<TFilterKey, string>;

/**
 * Optional externally controlled search state shared across multiple filtered lists.
 */
export type FilteredRecordSearchState = {
  search: string;
  setSearch: (value: string) => void;
};

/**
 * Input accepted by the generic filtered-records hook.
 */
export type UseFilteredRecordsInput<TRecord, TFilterKey extends string> = {
  /**
   * Stable record list that should be searched and filtered.
   */
  records: TRecord[];
  /**
   * Resolves the combined search text for one record.
   */
  getSearchText: (record: TRecord) => string;
  /**
   * Initial string-valued filter state owned by the hook.
   */
  initialFilters: FilteredRecordFilterState<TFilterKey>;
  /**
   * Matches one record against one named filter value.
   */
  matchesFilter: (record: TRecord, key: TFilterKey, value: string) => boolean;
  /**
   * Optional externally controlled search state shared by multiple hook instances.
   */
  searchState?: FilteredRecordSearchState;
};

/**
 * Runtime state returned by the generic filtered-records hook.
 */
export type UseFilteredRecordsResult<TRecord, TFilterKey extends string> = {
  /**
   * Records that match the current search term and active filters.
   */
  records: TRecord[];
  /**
   * Current free-text search term.
   */
  search: string;
  /**
   * Updates the free-text search term.
   */
  setSearch: (value: string) => void;
  /**
   * Current named filter values.
   */
  filters: FilteredRecordFilterState<TFilterKey>;
  /**
   * Updates one named filter value.
   */
  setFilterValue: (key: TFilterKey, value: string) => void;
  /**
   * Resets the search term and all filters to their initial values.
   */
  resetFilters: () => void;
  /**
   * Total unfiltered record count.
   */
  totalCount: number;
  /**
   * Count after applying search and filters.
   */
  filteredCount: number;
};

/**
 * Provides one generic local search-and-filter state machine for keyed record lists.
 */
export const useFilteredRecords = <TRecord, TFilterKey extends string>({
  records,
  getSearchText,
  initialFilters,
  matchesFilter,
  searchState,
}: UseFilteredRecordsInput<TRecord, TFilterKey>): UseFilteredRecordsResult<TRecord, TFilterKey> => {
  const [internalSearch, setInternalSearch] = useState("");
  const [filters, setFilters] = useState<FilteredRecordFilterState<TFilterKey>>(initialFilters);
  const search = searchState?.search ?? internalSearch;
  const setSearch = searchState?.setSearch ?? setInternalSearch;
  const deferredSearch = useDeferredValue(search);

  const filteredRecords = useMemo(() => {
    const normalizedSearch = deferredSearch.trim().toLocaleLowerCase();
    const activeFilters = Object.entries(filters) as Array<[TFilterKey, string]>;

    return records.filter((record) => {
      return (
        matchesSearch(record, normalizedSearch, getSearchText) &&
        matchesActiveFilters(record, activeFilters, matchesFilter)
      );
    });
  }, [deferredSearch, filters, getSearchText, matchesFilter, records]);

  return {
    filteredCount: filteredRecords.length,
    filters,
    records: filteredRecords,
    resetFilters: () => {
      setFilters(initialFilters);
      setSearch("");
    },
    search,
    setFilterValue: (key, value) => {
      setFilters((currentFilters) => ({
        ...currentFilters,
        [key]: value,
      }));
    },
    setSearch,
    totalCount: records.length,
  };
};

/**
 * Matches one record against the current free-text search term.
 */
const matchesSearch = <TRecord>(
  record: TRecord,
  normalizedSearch: string,
  getSearchText: (record: TRecord) => string,
): boolean => {
  if (normalizedSearch.length === 0) {
    return true;
  }

  const searchText = getSearchText(record).toLocaleLowerCase();
  return searchText.includes(normalizedSearch);
};

/**
 * Matches one record against the current set of non-empty named filters.
 */
const matchesActiveFilters = <TRecord, TFilterKey extends string>(
  record: TRecord,
  filters: Array<[TFilterKey, string]>,
  matchesFilter: (record: TRecord, key: TFilterKey, value: string) => boolean,
): boolean => {
  return filters.every(([key, value]) => {
    if (value.trim().length === 0) {
      return true;
    }

    return matchesFilter(record, key, value);
  });
};
