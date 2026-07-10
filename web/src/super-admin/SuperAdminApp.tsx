import { useCallback, useEffect, useState } from "react";
import type { ReactNode } from "react";
import {
  getPlatformToken,
  platformAuditApi,
  platformAuthApi,
  platformSubscriptionsApi,
  platformSystemApi,
  platformTenantsApi,
  platformUsersApi,
  setPlatformToken,
  type CurrentPlatformUserResponse,
  type PlatformAuthResponse,
  type PlatformSubscription,
  type PlatformSystemHealth,
  type PlatformTenant,
  type PlatformTenantUser,
  type SuperAdminAuditEvent,
} from "../api/platform";
import { Icon, humanize, initials } from "../ui";
import type { IconName } from "../ui";
import { PlatformLogin } from "./PlatformLogin";
import { Tenants } from "./Tenants";
import { Users } from "./Users";
import { Subscriptions } from "./Subscriptions";
import { Health } from "./Health";
import { Audit } from "./Audit";

type View = "tenants" | "users" | "subscriptions" | "health" | "audit";

const NAV: { id: View; label: string; icon: IconName }[] = [
  { id: "tenants", label: "Tenants", icon: "briefcase" },
  { id: "users", label: "Login Users", icon: "users" },
  { id: "subscriptions", label: "Subscriptions", icon: "list" },
  { id: "health", label: "System Health", icon: "activity" },
  { id: "audit", label: "Audit Log", icon: "filter" },
];

const TITLES: Record<View, string> = {
  tenants: "Tenants",
  users: "Login Users",
  subscriptions: "Subscriptions",
  health: "System Health",
  audit: "Audit Log",
};

export function SuperAdminApp() {
  const [me, setMe] = useState<CurrentPlatformUserResponse | null>(null);
  const [booting, setBooting] = useState(true);
  const [view, setView] = useState<View>("tenants");

  const [tenants, setTenants] = useState<PlatformTenant[]>([]);
  const [users, setUsers] = useState<PlatformTenantUser[]>([]);
  const [subscriptions, setSubscriptions] = useState<PlatformSubscription[]>([]);
  const [health, setHealth] = useState<PlatformSystemHealth | null>(null);
  const [events, setEvents] = useState<SuperAdminAuditEvent[]>([]);
  const [auditFilters, setAuditFilters] = useState<{ action?: string; limit?: number }>({});

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  // Re-validate a persisted platform session against /platform/auth/me on boot.
  useEffect(() => {
    let cancelled = false;
    async function boot() {
      if (!getPlatformToken()) {
        setBooting(false);
        return;
      }
      try {
        const result = await platformAuthApi.me();
        if (!cancelled) setMe(result);
      } catch {
        setPlatformToken(null);
      } finally {
        if (!cancelled) setBooting(false);
      }
    }
    void boot();
    return () => {
      cancelled = true;
    };
  }, []);

  const loadData = useCallback(async () => {
    if (!me) return;
    setLoading(true);
    setError(null);
    try {
      const [tenantList, userList, subList, healthResult, eventList] = await Promise.all([
        platformTenantsApi.list(),
        platformUsersApi.list(),
        platformSubscriptionsApi.list(),
        platformSystemApi.health(),
        platformAuditApi.list(auditFilters),
      ]);
      setTenants(tenantList);
      setUsers(userList);
      setSubscriptions(subList);
      setHealth(healthResult);
      setEvents(eventList);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load platform data.");
    } finally {
      setLoading(false);
    }
  }, [me, auditFilters]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  function completeAuth(result: PlatformAuthResponse): void {
    setPlatformToken(result.token);
    setMe({ user: result.user });
  }

  function signOut(): void {
    setPlatformToken(null);
    setMe(null);
    setTenants([]);
    setUsers([]);
    setSubscriptions([]);
    setHealth(null);
    setEvents([]);
    setError(null);
    setNotice(null);
  }

  const runMutation = useCallback(
    async (action: () => Promise<void>, successMessage: string) => {
      setError(null);
      setNotice(null);
      try {
        await action();
        await loadData();
        setNotice(successMessage);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Request failed.");
      }
    },
    [loadData],
  );

  if (booting) {
    return (
      <div
        style={{
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "var(--background)",
          color: "var(--adaptive-500)",
          fontFamily: "var(--font-sans)",
          fontSize: 14,
        }}
      >
        Loading platform console…
      </div>
    );
  }

  if (!me) {
    return <PlatformLogin onAuthenticated={completeAuth} />;
  }

  return (
    <div style={{ display: "flex", height: "100vh", overflow: "hidden", background: "var(--background)" }}>
      <PlatformSidebar
        active={view}
        onSelect={setView}
        email={me.user.email}
        role={me.user.role}
        onSignOut={signOut}
      />
      <main style={{ flex: 1, display: "flex", flexDirection: "column", minWidth: 0 }}>
        <PlatformTopBar title={TITLES[view]} loading={loading} onRefresh={() => void loadData()} />
        <div style={{ flex: 1, overflow: "auto", background: "var(--background)" }}>
          {view === "tenants" && (
            <Tenants
              tenants={tenants}
              loading={loading}
              onUpdate={(id, body, message) =>
                runMutation(async () => void (await platformTenantsApi.update(id, body)), message)
              }
            />
          )}
          {view === "users" && (
            <Users
              users={users}
              loading={loading}
              onSetStatus={(id, status, message) =>
                runMutation(async () => void (await platformUsersApi.update(id, { status })), message)
              }
            />
          )}
          {view === "subscriptions" && (
            <Subscriptions
              subscriptions={subscriptions}
              loading={loading}
              onUpdate={(id, body, message) =>
                runMutation(async () => void (await platformSubscriptionsApi.update(id, body)), message)
              }
            />
          )}
          {view === "health" && <Health health={health} loading={loading} />}
          {view === "audit" && (
            <Audit
              events={events}
              loading={loading}
              onFilter={async (filters) => {
                setAuditFilters(filters);
              }}
            />
          )}
        </div>
      </main>

      {error && (
        <PlatformBanner kind="error" onClose={() => setError(null)}>
          {error}
        </PlatformBanner>
      )}
      {notice && !error && (
        <PlatformBanner kind="notice" onClose={() => setNotice(null)}>
          {notice}
        </PlatformBanner>
      )}
    </div>
  );
}

function PlatformSidebar({
  active,
  onSelect,
  email,
  role,
  onSignOut,
}: {
  active: View;
  onSelect: (v: View) => void;
  email: string;
  role: string;
  onSignOut: () => void;
}) {
  return (
    <aside
      style={{
        width: 236,
        height: "100%",
        background: "var(--adaptive-950)",
        color: "#fff",
        display: "flex",
        flexDirection: "column",
        fontFamily: "var(--font-sans)",
        flexShrink: 0,
      }}
    >
      <div
        style={{
          padding: "16px 16px",
          display: "flex",
          alignItems: "center",
          gap: 10,
          borderBottom: "1px solid rgba(255,255,255,0.08)",
        }}
      >
        <div
          style={{
            width: 30,
            height: 30,
            borderRadius: 7,
            background: "var(--blue-600)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flexShrink: 0,
          }}
        >
          <Icon name="activity" size={17} color="#fff" />
        </div>
        <div style={{ minWidth: 0 }}>
          <div style={{ fontWeight: 700, fontSize: 14.5, letterSpacing: "-0.01em" }}>Opero Platform</div>
          <div style={{ fontSize: 10.5, color: "rgba(255,255,255,0.5)", fontWeight: 600, letterSpacing: ".06em" }}>
            INTERNAL CONSOLE
          </div>
        </div>
      </div>

      <nav style={{ flex: 1, overflow: "auto", padding: "12px 10px" }}>
        {NAV.map((item) => {
          const isActive = item.id === active;
          return (
            <button
              key={item.id}
              onClick={() => onSelect(item.id)}
              style={{
                width: "100%",
                display: "flex",
                alignItems: "center",
                gap: 10,
                padding: "9px 10px",
                borderRadius: 7,
                border: 0,
                cursor: "pointer",
                background: isActive ? "var(--blue-600)" : "transparent",
                color: isActive ? "#fff" : "rgba(255,255,255,0.72)",
                fontSize: 13.5,
                fontWeight: isActive ? 600 : 500,
                fontFamily: "inherit",
                textAlign: "left",
                marginTop: 2,
                transition: "background .15s",
              }}
            >
              <Icon name={item.icon} size={17} color={isActive ? "#fff" : "rgba(255,255,255,0.55)"} />
              {item.label}
            </button>
          );
        })}
      </nav>

      <div
        style={{
          borderTop: "1px solid rgba(255,255,255,0.08)",
          padding: "12px",
          display: "flex",
          alignItems: "center",
          gap: 10,
        }}
      >
        <div
          style={{
            width: 30,
            height: 30,
            borderRadius: "50%",
            background: "var(--blue-600)",
            color: "#fff",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontWeight: 600,
            fontSize: 12,
            flexShrink: 0,
          }}
        >
          {initials(email)}
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              fontSize: 12.5,
              fontWeight: 600,
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {email}
          </div>
          <div style={{ fontSize: 11, color: "rgba(255,255,255,0.5)" }}>{humanize(role)}</div>
        </div>
        <button
          onClick={onSignOut}
          title="Sign out"
          style={{
            width: 30,
            height: 30,
            borderRadius: 6,
            border: "1px solid rgba(255,255,255,0.15)",
            background: "transparent",
            cursor: "pointer",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flexShrink: 0,
          }}
        >
          <Icon name="x" size={15} color="rgba(255,255,255,0.7)" />
        </button>
      </div>
    </aside>
  );
}

function PlatformTopBar({
  title,
  loading,
  onRefresh,
}: {
  title: string;
  loading: boolean;
  onRefresh: () => void;
}) {
  return (
    <header
      style={{
        height: 56,
        display: "flex",
        alignItems: "center",
        gap: 14,
        padding: "0 24px",
        borderBottom: "1px solid var(--adaptive-200)",
        background: "var(--background)",
        fontFamily: "var(--font-sans)",
        flexShrink: 0,
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 9, fontSize: 13, color: "var(--adaptive-500)" }}>
        <span>Platform</span>
        <Icon name="chevron" size={14} color="var(--adaptive-300)" />
        <span style={{ color: "var(--adaptive-900)", fontWeight: 600 }}>{title}</span>
      </div>
      <span
        style={{
          fontSize: 9.5,
          fontWeight: 700,
          letterSpacing: ".08em",
          padding: "3px 7px",
          borderRadius: 5,
          background: "var(--blue-50)",
          color: "var(--blue-700)",
          border: "1px solid var(--blue-200)",
        }}
      >
        INTERNAL
      </span>
      <div style={{ flex: 1 }} />
      <button
        onClick={onRefresh}
        disabled={loading}
        title="Refresh"
        style={{
          width: 36,
          height: 36,
          borderRadius: 6,
          border: "1px solid var(--adaptive-200)",
          background: "var(--card)",
          cursor: loading ? "not-allowed" : "pointer",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Icon name="refresh" size={17} color="var(--adaptive-600)" />
      </button>
    </header>
  );
}

function PlatformBanner({
  kind,
  children,
  onClose,
}: {
  kind: "error" | "notice";
  children: ReactNode;
  onClose: () => void;
}) {
  const isError = kind === "error";
  return (
    <div
      style={{
        position: "fixed",
        bottom: 24,
        left: "50%",
        transform: "translateX(-50%)",
        zIndex: 90,
        display: "flex",
        alignItems: "center",
        gap: 10,
        padding: "11px 16px",
        borderRadius: 8,
        fontSize: 13,
        fontWeight: 500,
        boxShadow: "var(--shadow-lg)",
        background: isError ? "var(--red-700)" : "var(--adaptive-900)",
        color: "#fff",
        maxWidth: "90vw",
        fontFamily: "var(--font-sans)",
      }}
    >
      <Icon name={isError ? "alert" : "check"} size={16} color={isError ? "#fff" : "var(--green-400)"} />
      <span>{children}</span>
      <button
        onClick={onClose}
        style={{
          border: 0,
          background: "transparent",
          color: "rgba(255,255,255,0.7)",
          cursor: "pointer",
          display: "flex",
          marginLeft: 4,
        }}
      >
        <Icon name="x" size={14} />
      </button>
    </div>
  );
}
