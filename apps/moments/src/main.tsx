import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";

/** Mounts the standalone or embedded Moments browser shell. */
const root = document.getElementById("root");
if (root)
  createRoot(root).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
