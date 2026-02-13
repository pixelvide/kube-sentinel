## 2026-02-13 - [Repeated Computation in Column Accessors]
**Learning:** Table column accessors in `SimpleTable` implementations often call the same expensive helper function (e.g., `getPodStatus`) multiple times per row for different columns. This results in O(Columns * Rows) complexity instead of O(Rows).
**Action:** Memoize expensive calculations at the row level (e.g., using a `Map` in `useMemo` keyed by object ID) before passing to the table, or ensure the helper function itself is memoized.
