// Typed platform (Super Admin) data layer. Wraps the generated openapi-fetch
// client (platformClient.ts) so components never touch fetch or raw paths.
// All request/response types come from the generated schema.d.ts — never
// hand-written.
import { platformApi, setPlatformToken, getPlatformToken } from "./platformClient";
import type { components } from "./schema";

export type PlatformAuthResponse = components["schemas"]["PlatformAuthResponse"];
export type CurrentPlatformUserResponse =
  components["schemas"]["CurrentPlatformUserResponse"];
export type PlatformUserSummary = components["schemas"]["PlatformUserSummary"];
export type PlatformLoginRequest = components["schemas"]["PlatformLoginRequest"];
export type PlatformTenant = components["schemas"]["PlatformTenant"];
export type PlatformUpdateTenantRequest =
  components["schemas"]["PlatformUpdateTenantRequest"];
export type PlatformTenantUser = components["schemas"]["PlatformTenantUser"];
export type PlatformUpdateUserRequest =
  components["schemas"]["PlatformUpdateUserRequest"];
export type PlatformSubscription = components["schemas"]["PlatformSubscription"];
export type PlatformUpdateSubscriptionRequest =
  components["schemas"]["PlatformUpdateSubscriptionRequest"];
export type PlatformSystemHealth = components["schemas"]["PlatformSystemHealth"];
export type SuperAdminAuditEvent = components["schemas"]["SuperAdminAuditEvent"];
export type TenantStatus = PlatformTenant["status"];
export type UserSummary = components["schemas"]["UserSummary"];

export { setPlatformToken, getPlatformToken };

type ApiResult<T> = { data?: T; error?: unknown; response: Response };

function errorMessage(error: unknown, fallback: string): string {
  if (typeof error === "object" && error && "message" in error) {
    const message = (error as { message?: unknown }).message;
    if (typeof message === "string" && message.trim()) {
      return message;
    }
  }
  return fallback;
}

async function unwrap<T>(result: ApiResult<T>, fallback: string): Promise<T> {
  if (result.error || !result.response.ok) {
    throw new Error(errorMessage(result.error, fallback));
  }
  if (result.data === undefined) {
    throw new Error(fallback);
  }
  return result.data;
}

export const platformAuthApi = {
  async login(body: PlatformLoginRequest): Promise<PlatformAuthResponse> {
    return unwrap(
      await platformApi.POST("/platform/auth/login", { body }),
      "Unable to sign in to the platform console.",
    );
  },
  async me(): Promise<CurrentPlatformUserResponse> {
    return unwrap(
      await platformApi.GET("/platform/auth/me"),
      "Unable to load the current platform user.",
    );
  },
};

export const platformTenantsApi = {
  async list(): Promise<PlatformTenant[]> {
    return unwrap(
      await platformApi.GET("/platform/tenants"),
      "Unable to load tenants.",
    );
  },
  async get(id: string): Promise<PlatformTenant> {
    return unwrap(
      await platformApi.GET("/platform/tenants/{id}", { params: { path: { id } } }),
      "Unable to load the tenant.",
    );
  },
  async update(
    id: string,
    body: PlatformUpdateTenantRequest,
  ): Promise<PlatformTenant> {
    return unwrap(
      await platformApi.PATCH("/platform/tenants/{id}", {
        params: { path: { id } },
        body,
      }),
      "Unable to update the tenant.",
    );
  },
};

export const platformUsersApi = {
  async list(
    filters: {
      tenant_id?: string;
      role?: PlatformTenantUser["role"];
      status?: PlatformTenantUser["status"];
    } = {},
  ): Promise<PlatformTenantUser[]> {
    return unwrap(
      await platformApi.GET("/platform/users", { params: { query: filters } }),
      "Unable to load users.",
    );
  },
  async update(id: string, body: PlatformUpdateUserRequest): Promise<UserSummary> {
    return unwrap(
      await platformApi.PATCH("/platform/users/{id}", {
        params: { path: { id } },
        body,
      }),
      "Unable to update the user.",
    );
  },
};

export const platformSubscriptionsApi = {
  async list(
    filters: { tenant_id?: string; plan?: string; status?: string } = {},
  ): Promise<PlatformSubscription[]> {
    return unwrap(
      await platformApi.GET("/platform/subscriptions", {
        params: { query: filters },
      }),
      "Unable to load subscriptions.",
    );
  },
  async update(
    id: string,
    body: PlatformUpdateSubscriptionRequest,
  ): Promise<PlatformSubscription> {
    return unwrap(
      await platformApi.PATCH("/platform/subscriptions/{id}", {
        params: { path: { id } },
        body,
      }),
      "Unable to update the subscription.",
    );
  },
};

export const platformSystemApi = {
  async health(): Promise<PlatformSystemHealth> {
    return unwrap(
      await platformApi.GET("/platform/system/health"),
      "Unable to load system health.",
    );
  },
};

export const platformAuditApi = {
  async list(
    filters: {
      tenant_id?: string;
      actor_platform_user_id?: string;
      action?: string;
      limit?: number;
    } = {},
  ): Promise<SuperAdminAuditEvent[]> {
    return unwrap(
      await platformApi.GET("/platform/audit-events", {
        params: { query: filters },
      }),
      "Unable to load audit events.",
    );
  },
};
