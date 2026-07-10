import { useCallback, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  ApiError,
  authApi,
  departmentsApi,
  employeesApi,
  leaveApi,
  liveApi,
  locationsApi,
  rolesApi,
  setAuthToken,
  shiftsApi,
  toursApi,
  type AuthResponse,
  type CurrentUserResponse,
  type Department,
  type Employee,
  type LeaveRequest,
  type LiveViewEntry,
  type Location,
  type Role,
  type Shift,
  type Tour,
} from "./api/resources";
import { Icon, humanize, initials } from "./ui";
import type { IconName } from "./ui";
import { AuthScreen } from "./views/Auth";
import { LiveView } from "./views/LiveView";
import { Roster } from "./views/Roster";
import { People } from "./views/People";
import { Departments } from "./views/Departments";
import { Roles } from "./views/Roles";
import { Locations } from "./views/Locations";
import { Tours } from "./views/Tours";
import { TimeOff } from "./views/TimeOff";
import { EmployeeDetail } from "./views/EmployeeDetail";

type View = "live" | "roster" | "tours" | "locations" | "people" | "departments" | "roles" | "timeoff";

const NAV: { section: string; items: { id: View; label: string; icon: IconName; badge?: string }[] }[] = [
  {
    section: "Operations",
    items: [
      { id: "live", label: "Live View", icon: "activity", badge: "LIVE" },
      { id: "roster", label: "Roster", icon: "calendar" },
      { id: "tours", label: "Tours", icon: "map" },
      { id: "locations", label: "Locations", icon: "pin" },
    ],
  },
  {
    section: "People",
    items: [
      { id: "people", label: "Directory", icon: "users" },
      { id: "departments", label: "Departments", icon: "briefcase" },
      { id: "roles", label: "Roles", icon: "route" },
      { id: "timeoff", label: "Time Off", icon: "calendar" },
    ],
  },
];

const TITLES: Record<View, string> = {
  live: "Live View",
  roster: "Roster",
  tours: "Tours",
  locations: "Locations",
  people: "Directory",
  departments: "Departments",
  roles: "Roles",
  timeoff: "Time Off",
};

const PEOPLE_VIEWS: View[] = ["people", "departments", "roles", "timeoff"];

// Turn a thrown API error into a message for the error banner. A 403
// ("forbidden") means the session is valid but the current role can't perform
// the action — surface a clear, non-destructive notice and never sign the user
// out. Any other error keeps its server/fallback message.
function describeError(err: unknown, fallback: string): string {
  if (err instanceof ApiError && err.isForbidden) {
    return "You don't have permission to do that. This action requires an admin.";
  }
  if (err instanceof Error) {
    return err.message;
  }
  return fallback;
}

function OperoMark({ size = 28 }: { size?: number }) {
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: 7,
        flexShrink: 0,
        background: "linear-gradient(180deg, var(--primary-500), var(--primary-600))",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        boxShadow: "inset 0 1px 0 rgba(255,255,255,0.25)",
      }}
    >
      <svg
        width={size * 0.58}
        height={size * 0.58}
        viewBox="0 0 24 24"
        fill="none"
        stroke="#fff"
        strokeWidth="2.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <circle cx="12" cy="12" r="7.5" />
        <circle cx="12" cy="12" r="1.6" fill="#fff" />
      </svg>
    </div>
  );
}

function Sidebar({
  active,
  onSelect,
  tenantName,
  userEmail,
  userRole,
  onSignOut,
}: {
  active: View;
  onSelect: (v: View) => void;
  tenantName: string;
  userEmail: string;
  userRole: string;
  onSignOut: () => void;
}) {
  return (
    <aside
      style={{
        width: 232,
        height: "100%",
        borderRight: "1px solid var(--adaptive-200)",
        background: "var(--adaptive-50)",
        display: "flex",
        flexDirection: "column",
        fontFamily: "var(--font-sans)",
        flexShrink: 0,
      }}
    >
      <div
        style={{
          padding: "14px 16px",
          display: "flex",
          alignItems: "center",
          gap: 10,
          borderBottom: "1px solid var(--adaptive-200)",
        }}
      >
        <OperoMark />
        <div style={{ fontWeight: 700, fontSize: 16, letterSpacing: "-0.02em", color: "var(--adaptive-900)" }}>
          Opero
        </div>
        <div
          style={{
            marginLeft: "auto",
            fontSize: 10,
            color: "var(--adaptive-400)",
            fontWeight: 600,
            border: "1px solid var(--adaptive-200)",
            borderRadius: 5,
            padding: "1px 5px",
          }}
        >
          BETA
        </div>
      </div>

      <div style={{ padding: "12px 12px 4px" }}>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 9,
            padding: "8px 10px",
            background: "var(--card)",
            border: "1px solid var(--adaptive-200)",
            borderRadius: 7,
          }}
        >
          <div
            style={{
              width: 26,
              height: 26,
              borderRadius: 6,
              background: "var(--primary-600)",
              color: "#fff",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              fontWeight: 700,
              fontSize: 12,
            }}
          >
            {initials(tenantName)}
          </div>
          <div style={{ minWidth: 0, flex: 1 }}>
            <div
              style={{
                fontSize: 13,
                fontWeight: 600,
                color: "var(--adaptive-900)",
                whiteSpace: "nowrap",
                overflow: "hidden",
                textOverflow: "ellipsis",
              }}
            >
              {tenantName}
            </div>
            <div style={{ fontSize: 11, color: "var(--adaptive-500)" }}>Workspace</div>
          </div>
        </div>
      </div>

      <nav style={{ flex: 1, overflow: "auto", padding: "8px 8px 16px" }}>
        {NAV.map((group) => (
          <div key={group.section} style={{ marginTop: 12 }}>
            <div
              style={{
                padding: "4px 10px",
                fontSize: 11,
                fontWeight: 600,
                letterSpacing: ".06em",
                textTransform: "uppercase",
                color: "var(--adaptive-500)",
              }}
            >
              {group.section}
            </div>
            {group.items.map((item) => {
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
                    padding: "8px 10px",
                    borderRadius: 6,
                    border: 0,
                    cursor: "pointer",
                    background: isActive ? "var(--primary-50)" : "transparent",
                    color: isActive ? "var(--primary-700)" : "var(--adaptive-700)",
                    fontSize: 13.5,
                    fontWeight: isActive ? 600 : 500,
                    fontFamily: "inherit",
                    textAlign: "left",
                    marginTop: 2,
                    transition: "background .15s",
                  }}
                >
                  <Icon name={item.icon} size={18} color={isActive ? "var(--primary-600)" : "var(--adaptive-500)"} />
                  <span style={{ flex: 1 }}>{item.label}</span>
                  {item.badge && (
                    <span
                      style={{
                        fontSize: 9,
                        fontWeight: 700,
                        padding: "2px 6px",
                        borderRadius: 9999,
                        letterSpacing: ".05em",
                        display: "inline-flex",
                        alignItems: "center",
                        gap: 4,
                        background: "var(--green-50)",
                        color: "var(--green-700)",
                        border: "1px solid var(--green-200)",
                      }}
                    >
                      <span style={{ width: 5, height: 5, borderRadius: "50%", background: "var(--green-500)" }} />
                      {item.badge}
                    </span>
                  )}
                </button>
              );
            })}
          </div>
        ))}
      </nav>

      <div
        style={{
          borderTop: "1px solid var(--adaptive-200)",
          padding: "10px 12px",
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
            background: "var(--adaptive-700)",
            color: "#fff",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            fontWeight: 600,
            fontSize: 12,
            flexShrink: 0,
          }}
        >
          {initials(userEmail)}
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              fontSize: 12.5,
              fontWeight: 600,
              color: "var(--adaptive-900)",
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
          >
            {userEmail}
          </div>
          <div style={{ fontSize: 11, color: "var(--adaptive-500)" }}>{humanize(userRole)}</div>
        </div>
        <button
          onClick={onSignOut}
          title="Sign out"
          style={{
            width: 30,
            height: 30,
            borderRadius: 6,
            border: "1px solid var(--adaptive-200)",
            background: "var(--card)",
            cursor: "pointer",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flexShrink: 0,
          }}
        >
          <Icon name="x" size={15} color="var(--adaptive-500)" />
        </button>
      </div>
    </aside>
  );
}

function TopBar({
  view,
  crumb,
  onCrumbRoot,
  loading,
  onRefresh,
}: {
  view: View;
  crumb?: string | null;
  onCrumbRoot?: () => void;
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
        <span>{PEOPLE_VIEWS.includes(view) ? "People" : "Operations"}</span>
        <Icon name="chevron" size={14} color="var(--adaptive-300)" />
        {crumb ? (
          <>
            <button
              onClick={onCrumbRoot}
              style={{
                border: 0,
                background: "none",
                padding: 0,
                cursor: "pointer",
                fontFamily: "inherit",
                fontSize: 13,
                color: "var(--adaptive-500)",
              }}
            >
              {TITLES[view]}
            </button>
            <Icon name="chevron" size={14} color="var(--adaptive-300)" />
            <span style={{ color: "var(--adaptive-900)", fontWeight: 600 }}>{crumb}</span>
          </>
        ) : (
          <span style={{ color: "var(--adaptive-900)", fontWeight: 600 }}>{TITLES[view]}</span>
        )}
      </div>
      <div style={{ flex: 1 }} />
      <div
        style={{
          height: 34,
          display: "flex",
          alignItems: "center",
          gap: 8,
          padding: "0 10px",
          background: "var(--adaptive-100)",
          borderRadius: 6,
          fontSize: 13,
          color: "var(--adaptive-500)",
          minWidth: 170,
        }}
      >
        <Icon name="search" size={15} />
        <span>Search people, shifts…</span>
        <span style={{ marginLeft: "auto", fontSize: 10, fontWeight: 600, color: "var(--adaptive-400)" }}>⌘K</span>
      </div>
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
      <button
        style={{
          position: "relative",
          width: 36,
          height: 36,
          borderRadius: 6,
          border: "1px solid var(--adaptive-200)",
          background: "var(--card)",
          cursor: "pointer",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Icon name="bell" size={17} color="var(--adaptive-600)" />
        <span
          style={{
            position: "absolute",
            top: 6,
            right: 7,
            width: 7,
            height: 7,
            borderRadius: "50%",
            background: "var(--red-500)",
            border: "1.5px solid var(--card)",
          }}
        />
      </button>
    </header>
  );
}

function Banner({ kind, children, onClose }: { kind: "error" | "notice"; children: ReactNode; onClose: () => void }) {
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
        animation: "opero-slide .2s ease-out",
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

export function App() {
  const [auth, setAuth] = useState<AuthResponse | null>(null);
  const [currentUser, setCurrentUser] = useState<CurrentUserResponse | null>(null);
  const [view, setView] = useState<View>("live");
  const [departments, setDepartments] = useState<Department[]>([]);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [shifts, setShifts] = useState<Shift[]>([]);
  const [live, setLive] = useState<LiveViewEntry[]>([]);
  const [tours, setTours] = useState<Tour[]>([]);
  const [leaveRequests, setLeaveRequests] = useState<LeaveRequest[]>([]);
  const [selectedEmployee, setSelectedEmployee] = useState<Employee | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    if (!auth) return;
    setLoading(true);
    setError(null);
    try {
      const [me, deps, emps, locs, roleList, shiftList, liveList, tourList, leaveList] = await Promise.all([
        authApi.me(),
        departmentsApi.list(),
        employeesApi.list(),
        locationsApi.list(),
        rolesApi.list(),
        shiftsApi.list(),
        liveApi.list(),
        toursApi.list(),
        leaveApi.list(),
      ]);
      setCurrentUser(me);
      setDepartments(deps);
      setEmployees(emps);
      setLocations(locs);
      setRoles(roleList);
      setShifts(shiftList);
      setLive(liveList);
      setTours(tourList);
      setLeaveRequests(leaveList);
    } catch (err) {
      setError(describeError(err, "Unable to load data."));
    } finally {
      setLoading(false);
    }
  }, [auth]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const departmentNames = useMemo(() => new Map(departments.map((d) => [d.id, d.name])), [departments]);
  const roleNames = useMemo(() => new Map(roles.map((r) => [r.id, r.name])), [roles]);
  const locationNames = useMemo(() => new Map(locations.map((l) => [l.id, l.name])), [locations]);

  function completeAuth(result: AuthResponse): void {
    setAuthToken(result.token);
    setAuth(result);
    setCurrentUser({ user: result.user, tenant: result.tenant });
  }

  function signOut(): void {
    setAuthToken(null);
    setAuth(null);
    setCurrentUser(null);
    setDepartments([]);
    setEmployees([]);
    setLocations([]);
    setRoles([]);
    setShifts([]);
    setLive([]);
    setTours([]);
    setLeaveRequests([]);
    setSelectedEmployee(null);
    setNotice(null);
    setError(null);
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
        setError(describeError(err, "Request failed."));
      }
    },
    [loadData],
  );

  if (!auth) {
    return <AuthScreen onAuthenticated={completeAuth} />;
  }

  const tenantName = currentUser?.tenant.name ?? auth.tenant.name;
  const userEmail = currentUser?.user.email ?? auth.user.email;
  const userRole = currentUser?.user.role ?? auth.user.role;
  // Coarse RBAC (mirrors the backend): admin-only surfaces are hidden from
  // managers/employees so they never see buttons that would only 403. The
  // server still enforces this; the UI just avoids dead controls.
  const isAdmin = userRole === "admin";

  return (
    <div style={{ display: "flex", height: "100vh", overflow: "hidden", background: "var(--background)" }}>
      <Sidebar
        active={view}
        onSelect={(v) => {
          setSelectedEmployee(null);
          setView(v);
        }}
        tenantName={tenantName}
        userEmail={userEmail}
        userRole={userRole}
        onSignOut={signOut}
      />
      <main style={{ flex: 1, display: "flex", flexDirection: "column", minWidth: 0, background: "var(--background)" }}>
        <TopBar
          view={view}
          crumb={view === "people" && selectedEmployee ? selectedEmployee.full_name : null}
          onCrumbRoot={() => setSelectedEmployee(null)}
          loading={loading}
          onRefresh={() => void loadData()}
        />
        <div style={{ flex: 1, overflow: "auto", background: "var(--background)" }}>
          {view === "live" && (
            <LiveView
              entries={live}
              leaveRequests={leaveRequests}
              locationNames={locationNames}
              onRefresh={() => void loadData()}
              loading={loading}
            />
          )}
          {view === "roster" && (
            <Roster
              employees={employees}
              locations={locations}
              departments={departments}
              shifts={shifts}
              tours={tours}
              leaveRequests={leaveRequests}
              locationNames={locationNames}
              onCreate={(body) => runMutation(async () => void (await shiftsApi.create(body)), "Draft shift added.")}
              onUpdate={(id, body) => runMutation(async () => void (await shiftsApi.update(id, body)), "Shift updated.")}
              onDelete={(id) => void runMutation(async () => await shiftsApi.delete(id), "Shift deleted.")}
              onPublish={(id) => void runMutation(async () => void (await shiftsApi.publish(id)), "Shift published.")}
              onPublishMany={(ids) =>
                runMutation(async () => {
                  for (const id of ids) {
                    await shiftsApi.publish(id);
                  }
                }, `Published ${ids.length} shift${ids.length === 1 ? "" : "s"} · field staff notified.`)
              }
            />
          )}
          {view === "tours" && (
            <Tours
              tours={tours}
              employees={employees}
              shifts={shifts}
              live={live}
              onCreate={(body) => runMutation(async () => void (await toursApi.create(body)), "Tour created.")}
              onUpdate={(id, body) => runMutation(async () => void (await toursApi.update(id, body)), "Tour updated.")}
              onDelete={(id) => void runMutation(async () => await toursApi.delete(id), "Tour deleted.")}
            />
          )}
          {view === "locations" && (
            <Locations
              locations={locations}
              onCreate={(body) => runMutation(async () => void (await locationsApi.create(body)), "Location created.")}
              onDelete={(id) => void runMutation(async () => await locationsApi.delete(id), "Location deleted.")}
            />
          )}
          {view === "people" &&
            (selectedEmployee ? (
              <EmployeeDetail
                employee={selectedEmployee}
                shifts={shifts}
                live={live}
                leaveRequests={leaveRequests}
                locationNames={locationNames}
                departmentName={
                  selectedEmployee.department_id ? departmentNames.get(selectedEmployee.department_id) ?? null : null
                }
                roleName={selectedEmployee.role_id ? roleNames.get(selectedEmployee.role_id) ?? null : null}
                roleAccessLevel={
                  selectedEmployee.role_id
                    ? roles.find((r) => r.id === selectedEmployee.role_id)?.access_level ?? null
                    : null
                }
                reportsToName={
                  selectedEmployee.reports_to
                    ? employees.find((e) => e.id === selectedEmployee.reports_to)?.full_name ?? null
                    : null
                }
                onNavigate={(v) => {
                  setSelectedEmployee(null);
                  setView(v);
                }}
              />
            ) : (
              <People
                employees={employees}
                departments={departments}
                roles={roles}
                departmentNames={departmentNames}
                roleNames={roleNames}
                canDelete={isAdmin}
                onOpen={setSelectedEmployee}
                onCreate={(body) => runMutation(async () => void (await employeesApi.create(body)), "Employee added.")}
                onUpdate={(id, body) => runMutation(async () => void (await employeesApi.update(id, body)), "Employee updated.")}
                onDelete={(id) => void runMutation(async () => await employeesApi.delete(id), "Employee deleted.")}
              />
            ))}
          {view === "departments" && (
            <Departments
              departments={departments}
              employees={employees}
              canManage={isAdmin}
              onCreate={(body) => runMutation(async () => void (await departmentsApi.create(body)), "Department created.")}
              onUpdate={(id, body) => runMutation(async () => void (await departmentsApi.update(id, body)), "Department updated.")}
              onDelete={(id) => void runMutation(async () => await departmentsApi.delete(id), "Department deleted.")}
            />
          )}
          {view === "roles" && (
            <Roles
              roles={roles}
              employees={employees}
              departments={departments}
              canManage={isAdmin}
              onCreate={(body) => runMutation(async () => void (await rolesApi.create(body)), "Role created.")}
              onUpdate={(id, body) => runMutation(async () => void (await rolesApi.update(id, body)), "Role updated.")}
              onDelete={(id) => void runMutation(async () => await rolesApi.delete(id), "Role deleted.")}
            />
          )}
          {view === "timeoff" && (
            <TimeOff
              requests={leaveRequests}
              employees={employees}
              onApprove={(id) => void runMutation(async () => void (await leaveApi.approve(id)), "Request approved.")}
              onReject={(id) => void runMutation(async () => void (await leaveApi.reject(id)), "Request rejected.")}
            />
          )}
        </div>
      </main>

      {error && (
        <Banner kind="error" onClose={() => setError(null)}>
          {error}
        </Banner>
      )}
      {notice && !error && (
        <Banner kind="notice" onClose={() => setNotice(null)}>
          {notice}
        </Banner>
      )}
    </div>
  );
}
