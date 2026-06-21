// Typed API client for Opero, generated-spec-driven.
//
// Types come from src/api/schema.d.ts (produced by `npm run gen:api` from
// ../api/openapi.yaml — do not hand-edit). This module is the ONLY place that
// talks to the API; components call `api.GET/POST/...` through it.
import createClient, { type Middleware } from "openapi-fetch";
import type { paths } from "./schema";

const baseUrl =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

export const api = createClient<paths>({ baseUrl });

// --- auth: attach the bearer token to every request ---
//
// The token comes from POST /auth/login (or /auth/signup). Keep it in memory /
// an auth context; set it here. Never log it.
let authToken: string | null = null;

export function setAuthToken(token: string | null): void {
  authToken = token;
}

const authMiddleware: Middleware = {
  onRequest({ request }) {
    if (authToken) {
      request.headers.set("Authorization", `Bearer ${authToken}`);
    }
    return request;
  },
};

api.use(authMiddleware);

// NOTE: the exact openapi-fetch middleware API (`Middleware`, `.use()`,
// `onRequest` signature) can differ across versions. If `npm run build` fails
// here, align this with the installed openapi-fetch version's docs — the intent
// is simply "inject Authorization: Bearer <token> on each request".
