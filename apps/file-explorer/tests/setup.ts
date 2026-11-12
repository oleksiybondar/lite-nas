import "@testing-library/jest-dom/vitest";

/** Enables React act diagnostics for interaction tests. */
(globalThis as typeof globalThis & { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT =
  true;
