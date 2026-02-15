## 2026-02-13 - [Repeated Computation in Column Accessors]
**Learning:** Table column accessors in `SimpleTable` implementations often call the same expensive helper function (e.g., `getPodStatus`) multiple times per row for different columns. This results in O(Columns * Rows) complexity instead of O(Rows).
**Action:** Memoize expensive calculations at the row level (e.g., using a `Map` in `useMemo` keyed by object ID) before passing to the table, or ensure the helper function itself is memoized.

## 2026-02-13 - [Memoization of Helper Functions]
**Learning:** When multiple components or table columns call the same expensive helper function (like `getPodStatus`) with the same object reference, memoizing the helper function itself using a `WeakMap` is a clean and effective optimization that avoids refactoring all call sites.
**Action:** Use `WeakMap` to cache results of expensive stateless functions that take an object as input.
