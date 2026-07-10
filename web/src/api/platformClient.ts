// Typed API client for the Opero PLATFORM (Super Admin) surface.
//
// This is a SEPARATE client instance from the tenant `api` in ./client.ts.
// It injects a distinct platform JWT (kind=platform, no tenant_id) and is the
// ONLY place platform-token-bearing requests are made. The tenant token and
// the platform token are kept fully separate so a tenant login and a platform
// login can coexist without clobbering each other:
//   - the tenant client injects the tenant token on tenant requests;
//   - this client injects the platform token on /platform/* requests only.
// Never send one token to the other surface.
import createClient, { type Middleware } from "openapi-fetch";
import type { paths } from "./schema";

const baseUrl = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

// Distinct storage key from the tenant token so the two never collide. The
// platform session is persisted so a Super Admin reload keeps the console open;
// it is re-validated against GET /platform/auth/me on boot.
const PLATFORM_TOKEN_KEY = "opero_platform_token";

let platformToken: string | null = readStoredToken();

function readStoredToken(): string | null {
  try {
    return window.localStorage.getItem(PLATFORM_TOKEN_KEY);
  } catch {
    return null;
  }
}

export function getPlatformToken(): string | null {
  return platformToken;
}

export function setPlatformToken(token: string | null): void {
  platformToken = token;
  try {
    if (token) {
      window.localStorage.setItem(PLATFORM_TOKEN_KEY, token);
    } else {
      window.localStorage.removeItem(PLATFORM_TOKEN_KEY);
    }
  } catch {
    // localStorage may be unavailable (private mode); in-memory token still works.
  }
}

export const platformApi = createClient<paths>({ baseUrl });

const platformAuthMiddleware: Middleware = {
  onRequest({ request }) {
    if (platformToken) {
      request.headers.set("Authorization", `Bearer ${platformToken}`);
    }
    return request;
  },
};

platformApi.use(platformAuthMiddleware);
