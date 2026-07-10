import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { SuperAdminApp } from "./super-admin/SuperAdminApp";
import "./styles.css";

const rootEl = document.getElementById("root");
if (!rootEl) {
  throw new Error("root element not found");
}

// Route split: the internal Super Admin console lives under a dedicated
// /super-admin path group, fully separate from the tenant manager app. It has
// its own auth (platform JWT) and chrome. Kept as a simple path check to avoid
// pulling in a router; Vite's SPA fallback serves index.html for the deep path.
const isSuperAdmin = window.location.pathname.startsWith("/super-admin");

createRoot(rootEl).render(
  <StrictMode>{isSuperAdmin ? <SuperAdminApp /> : <App />}</StrictMode>,
);
