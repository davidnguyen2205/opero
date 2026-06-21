import { Children, useCallback, useEffect, useMemo, useState } from "react";
import type { CSSProperties, FormEvent, ReactNode } from "react";
import {
  authApi,
  departmentsApi,
  employeesApi,
  liveApi,
  locationsApi,
  rolesApi,
  setAuthToken,
  shiftsApi,
  type AuthResponse,
  type CreateDepartmentRequest,
  type CreateEmployeeRequest,
  type CreateLocationRequest,
  type CreateRoleRequest,
  type CreateShiftRequest,
  type CurrentUserResponse,
  type Department,
  type Employee,
  type LiveViewEntry,
  type Location,
  type Role,
  type Shift,
} from "./api/resources";

type View = "live" | "roster" | "people" | "departments" | "roles" | "locations";
type AuthMode = "login" | "signup";

type FormState = Record<string, string>;

const views: { id: View; label: string; section: string; icon: IconName }[] = [
  { id: "live", label: "Live", section: "Operations", icon: "activity" },
  { id: "roster", label: "Roster", section: "Operations", icon: "calendar" },
  { id: "people", label: "People", section: "People", icon: "users" },
  { id: "departments", label: "Departments", section: "People", icon: "grid" },
  { id: "roles", label: "Roles", section: "People", icon: "briefcase" },
  { id: "locations", label: "Locations", section: "Operations", icon: "pin" },
];

const navSections = ["Operations", "People"];

const shiftColors = [
  "#ea580c",
  "#2563eb",
  "#7c3aed",
  "#0d9488",
  "#db2777",
  "#d97706",
  "#15803d",
  "#4b5563",
];

type IconName =
  | "activity"
  | "calendar"
  | "users"
  | "grid"
  | "briefcase"
  | "pin"
  | "plus"
  | "refresh"
  | "send"
  | "x";

const iconPaths: Record<IconName, string> = {
  activity: "M22 12h-4l-3 9L9 3l-3 9H2",
  calendar:
    "M8 2v4M16 2v4M3 10h18M5 4h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z",
  users:
    "M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75",
  grid: "M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z",
  briefcase:
    "M20 7H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2zM16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2",
  pin: "M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0zM12 10a2 2 0 1 0 0-4 2 2 0 0 0 0 4z",
  plus: "M12 5v14M5 12h14",
  refresh: "M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15",
  send: "M22 2 11 13M22 2l-7 20-4-9-9-4 20-7z",
  x: "M18 6 6 18M6 6l12 12",
};

const employmentTypes: Employee["employment_type"][] = [
  "full_time",
  "part_time",
  "freelance",
  "seasonal",
];

function isNonEmpty(value: string): boolean {
  return value.trim().length > 0;
}

function optionalString(value: string): string | undefined {
  const trimmed = value.trim();
  return trimmed ? trimmed : undefined;
}

function optionalNullable(value: string): string | null | undefined {
  const trimmed = value.trim();
  return trimmed ? trimmed : undefined;
}

function optionalNumber(value: string): number | null | undefined {
  const trimmed = value.trim();
  return trimmed ? Number(trimmed) : undefined;
}

function toIsoFromLocal(value: string): string {
  return new Date(value).toISOString();
}

function dateTimeLocalFromIso(value: string): string {
  const date = new Date(value);
  const offsetMs = date.getTimezoneOffset() * 60_000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
}

function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function humanize(value: string): string {
  return value.replaceAll("_", " ");
}

function namesById<T extends { id: string; name: string }>(
  items: T[],
): Map<string, string> {
  return new Map(items.map((item) => [item.id, item.name]));
}

function initials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) {
    return parts[0].slice(0, 2).toUpperCase();
  }
  return `${parts[0][0]}${parts[parts.length - 1][0]}`.toUpperCase();
}

function startOfWeek(date: Date): Date {
  const copy = new Date(date);
  const day = copy.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  copy.setDate(copy.getDate() + diff);
  copy.setHours(0, 0, 0, 0);
  return copy;
}

function weekDays(anchor = new Date()): Date[] {
  const start = startOfWeek(anchor);
  return Array.from({ length: 7 }, (_, index) => {
    const date = new Date(start);
    date.setDate(start.getDate() + index);
    return date;
  });
}

function sameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

function formatDayHeader(date: Date): { dow: string; day: string } {
  return {
    dow: new Intl.DateTimeFormat(undefined, { weekday: "short" }).format(date),
    day: new Intl.DateTimeFormat(undefined, { day: "2-digit" }).format(date),
  };
}

function formatTime(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

function colorForId(id: string): string {
  let total = 0;
  for (const char of id) {
    total += char.charCodeAt(0);
  }
  return shiftColors[total % shiftColors.length];
}

function Icon({
  name,
  size = 18,
}: {
  name: IconName;
  size?: number;
}) {
  return (
    <svg
      aria-hidden="true"
      fill="none"
      height={size}
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.75"
      viewBox="0 0 24 24"
      width={size}
    >
      {iconPaths[name]
        .split("M")
        .filter(Boolean)
        .map((segment, index) => (
          <path d={`M${segment}`} key={index} />
        ))}
    </svg>
  );
}

function App() {
  const [auth, setAuth] = useState<AuthResponse | null>(null);
  const [currentUser, setCurrentUser] = useState<CurrentUserResponse | null>(
    null,
  );
  const [view, setView] = useState<View>("roster");
  const [departments, setDepartments] = useState<Department[]>([]);
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [locations, setLocations] = useState<Location[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [shifts, setShifts] = useState<Shift[]>([]);
  const [live, setLive] = useState<LiveViewEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    if (!auth) {
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [me, deps, emps, locs, roleList, shiftList, liveList] = await Promise.all([
        authApi.me(),
        departmentsApi.list(),
        employeesApi.list(),
        locationsApi.list(),
        rolesApi.list(),
        shiftsApi.list(),
        liveApi.list(),
      ]);
      setCurrentUser(me);
      setDepartments(deps);
      setEmployees(emps);
      setLocations(locs);
      setRoles(roleList);
      setShifts(shiftList);
      setLive(liveList);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load data.");
    } finally {
      setLoading(false);
    }
  }, [auth]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const departmentNames = useMemo(() => namesById(departments), [departments]);
  const roleNames = useMemo(() => namesById(roles), [roles]);
  const locationNames = useMemo(() => namesById(locations), [locations]);

  async function completeAuth(result: AuthResponse): Promise<void> {
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
    setNotice(null);
    setError(null);
  }

  async function runMutation(
    action: () => Promise<void>,
    successMessage: string,
  ): Promise<void> {
    setError(null);
    setNotice(null);
    try {
      await action();
      await loadData();
      setNotice(successMessage);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed.");
    }
  }

  if (!auth) {
    return <AuthScreen onAuthenticated={completeAuth} />;
  }

  const publishedShifts = shifts.filter((shift) => shift.status === "published");
  const activeEmployees = employees.filter((employee) => employee.status === "active");

  return (
    <div className="app-shell workspace">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">O</div>
          <div>
            <strong>Opero</strong>
            <span>{currentUser?.tenant.name ?? auth.tenant.name}</span>
          </div>
        </div>
        <div className="sidebar-search">
          <span>⌕</span>
          <span>Search</span>
          <kbd>⌘K</kbd>
        </div>
        <nav className="nav-list" aria-label="Primary">
          {navSections.map((section) => (
            <div className="nav-section" key={section}>
              <div className="nav-section-label">{section}</div>
              {views
                .filter((item) => item.section === section)
                .map((item) => (
                  <button
                    className={view === item.id ? "active" : ""}
                    key={item.id}
                    onClick={() => setView(item.id)}
                    type="button"
                  >
                    <Icon name={item.icon} size={16} />
                    <span>{item.label}</span>
                    <b>
                      {countForView(item.id, {
                        departments,
                        employees,
                        locations,
                        roles,
                        shifts,
                        live,
                      })}
                    </b>
                  </button>
                ))}
            </div>
          ))}
        </nav>
        <div className="session-card">
          <div className="avatar avatar-sm">
            {initials(currentUser?.user.email ?? auth.user.email)}
          </div>
          <div>
            <strong>{currentUser?.user.email ?? auth.user.email}</strong>
            <span>{humanize(currentUser?.user.role ?? auth.user.role)}</span>
          </div>
          <button className="secondary-btn" onClick={signOut} type="button">
            Sign out
          </button>
        </div>
      </aside>

      <main className="main-shell">
        <header className="topbar">
          <div className="breadcrumb">
            <span>Workspace</span>
            <span>/</span>
            <strong>{views.find((item) => item.id === view)?.label}</strong>
          </div>
          <button
            className="secondary-btn"
            disabled={loading}
            onClick={() => void loadData()}
            type="button"
          >
            <Icon name="refresh" size={15} />
            {loading ? "Refreshing" : "Refresh"}
          </button>
        </header>

        <div className="content">
          <section className="dashboard-grid" aria-label="Operational summary">
            <Metric label="Active employees" value={activeEmployees.length} />
            <Metric label="Departments" value={departments.length} />
            <Metric label="Published shifts" value={publishedShifts.length} />
            <Metric label="Locations" value={locations.length} />
          </section>

          {error ? <div className="error-banner">{error}</div> : null}
          {notice ? <div className="notice-banner">{notice}</div> : null}

        {view === "live" ? <LiveView entries={live} /> : null}

        {view === "roster" ? (
          <RosterView
            employees={employees}
            locationNames={locationNames}
            locations={locations}
            onCreate={(body) =>
              runMutation(
                async () => {
                  await shiftsApi.create(body);
                },
                "Shift created.",
              )
            }
            onDelete={(id) =>
              runMutation(
                async () => {
                  await shiftsApi.delete(id);
                },
                "Shift deleted.",
              )
            }
            onPublish={(id) =>
              runMutation(
                async () => {
                  await shiftsApi.publish(id);
                },
                "Shift published.",
              )
            }
            shifts={shifts}
          />
        ) : null}

        {view === "people" ? (
          <PeopleView
            departmentNames={departmentNames}
            departments={departments}
            employees={employees}
            onCreate={(body) =>
              runMutation(
                async () => {
                  await employeesApi.create(body);
                },
                "Employee created.",
              )
            }
            onDelete={(id) =>
              runMutation(
                async () => {
                  await employeesApi.delete(id);
                },
                "Employee deleted.",
              )
            }
            roleNames={roleNames}
            roles={roles}
          />
        ) : null}

        {view === "departments" ? (
          <DepartmentsView
            departments={departments}
            onCreate={(body) =>
              runMutation(
                async () => {
                  await departmentsApi.create(body);
                },
                "Department created.",
              )
            }
            onDelete={(id) =>
              runMutation(
                async () => {
                  await departmentsApi.delete(id);
                },
                "Department deleted.",
              )
            }
          />
        ) : null}

        {view === "roles" ? (
          <RolesView
            onCreate={(body) =>
              runMutation(
                async () => {
                  await rolesApi.create(body);
                },
                "Role created.",
              )
            }
            onDelete={(id) =>
              runMutation(
                async () => {
                  await rolesApi.delete(id);
                },
                "Role deleted.",
              )
            }
            roles={roles}
          />
        ) : null}

        {view === "locations" ? (
          <LocationsView
            locations={locations}
            onCreate={(body) =>
              runMutation(
                async () => {
                  await locationsApi.create(body);
                },
                "Location created.",
              )
            }
            onDelete={(id) =>
              runMutation(
                async () => {
                  await locationsApi.delete(id);
                },
                "Location deleted.",
              )
            }
          />
        ) : null}
        </div>
      </main>
    </div>
  );
}

function countForView(
  view: View,
  data: {
    departments: Department[];
    employees: Employee[];
    locations: Location[];
    roles: Role[];
    shifts: Shift[];
    live: LiveViewEntry[];
  },
): number {
  switch (view) {
    case "departments":
      return data.departments.length;
    case "locations":
      return data.locations.length;
    case "people":
      return data.employees.length;
    case "roles":
      return data.roles.length;
    case "roster":
      return data.shifts.length;
    case "live":
      return data.live.length;
  }
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

const liveStatusLabels: Record<LiveViewEntry["attendance_status"], string> = {
  not_checked_in: "Not checked in",
  checked_in: "On shift",
  checked_out: "Checked out",
};

function LiveView({ entries }: { entries: LiveViewEntry[] }) {
  const onShift = entries.filter((e) => e.attendance_status === "checked_in").length;
  return (
    <div className="card">
      <div className="section-head">
        <h2>Who's working now</h2>
        <span className="table-note">
          {onShift} on shift · {entries.length} scheduled
        </span>
      </div>
      {entries.length ? (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Employee</th>
                <th>Shift</th>
                <th>Status</th>
                <th>Checked in</th>
                <th>Checked out</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((e) => (
                <tr key={e.shift.id}>
                  <td>
                    <strong>{e.employee_name}</strong>
                  </td>
                  <td>
                    {formatDateTime(e.shift.starts_at)} → {formatDateTime(e.shift.ends_at)}
                  </td>
                  <td>
                    <span
                      className={`pill ${e.attendance_status === "not_checked_in" ? "warn" : ""}`}
                    >
                      {liveStatusLabels[e.attendance_status]}
                    </span>
                  </td>
                  <td>{e.check_in_at ? formatDateTime(e.check_in_at) : "—"}</td>
                  <td>{e.check_out_at ? formatDateTime(e.check_out_at) : "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="empty-state">No published shifts in the current window.</div>
      )}
    </div>
  );
}

function AuthScreen({
  onAuthenticated,
}: {
  onAuthenticated: (auth: AuthResponse) => Promise<void>;
}) {
  const [mode, setMode] = useState<AuthMode>("login");
  const [form, setForm] = useState<FormState>({
    admin_email: "",
    admin_full_name: "",
    admin_password: "",
    company_name: "",
    email: "",
    password: "",
    slug: "",
    tenant_slug: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function update(key: string, value: string): void {
    setForm((current) => ({ ...current, [key]: value }));
  }

  async function submit(event: FormEvent<HTMLFormElement>): Promise<void> {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const result =
        mode === "login"
          ? await authApi.login({
              email: form.email,
              password: form.password,
              tenant_slug: form.tenant_slug,
            })
          : await authApi.signup({
              admin_email: form.admin_email,
              admin_full_name: optionalString(form.admin_full_name),
              admin_password: form.admin_password,
              company_name: form.company_name,
              slug: optionalString(form.slug),
            });
      await onAuthenticated(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="auth-screen">
      <section className="auth-panel">
        <div>
          <h1>Opero</h1>
          <p className="muted">Manager console for people, roster, and field ops.</p>
        </div>
        <div className="mode-switch" role="tablist">
          <button
            className={mode === "login" ? "active" : ""}
            onClick={() => setMode("login")}
            type="button"
          >
            Log in
          </button>
          <button
            className={mode === "signup" ? "active" : ""}
            onClick={() => setMode("signup")}
            type="button"
          >
            Sign up
          </button>
        </div>
        {error ? <div className="error-banner">{error}</div> : null}
        <form className="form-grid" onSubmit={(event) => void submit(event)}>
          {mode === "login" ? (
            <>
              <label>
                Tenant slug
                <input
                  autoComplete="organization"
                  onChange={(event) => update("tenant_slug", event.target.value)}
                  required
                  value={form.tenant_slug}
                />
              </label>
              <label>
                Email
                <input
                  autoComplete="email"
                  onChange={(event) => update("email", event.target.value)}
                  required
                  type="email"
                  value={form.email}
                />
              </label>
              <label>
                Password
                <input
                  autoComplete="current-password"
                  onChange={(event) => update("password", event.target.value)}
                  required
                  type="password"
                  value={form.password}
                />
              </label>
            </>
          ) : (
            <>
              <label>
                Company name
                <input
                  onChange={(event) => update("company_name", event.target.value)}
                  required
                  value={form.company_name}
                />
              </label>
              <label>
                Tenant slug
                <input
                  onChange={(event) => update("slug", event.target.value)}
                  pattern="^[a-z0-9]+(?:-[a-z0-9]+)*$"
                  value={form.slug}
                />
              </label>
              <label>
                Admin full name
                <input
                  onChange={(event) =>
                    update("admin_full_name", event.target.value)
                  }
                  value={form.admin_full_name}
                />
              </label>
              <label>
                Admin email
                <input
                  autoComplete="email"
                  onChange={(event) => update("admin_email", event.target.value)}
                  required
                  type="email"
                  value={form.admin_email}
                />
              </label>
              <label>
                Admin password
                <input
                  autoComplete="new-password"
                  minLength={8}
                  onChange={(event) =>
                    update("admin_password", event.target.value)
                  }
                  required
                  type="password"
                  value={form.admin_password}
                />
              </label>
            </>
          )}
          <button className="primary-btn" disabled={submitting} type="submit">
            {submitting ? "Working" : mode === "login" ? "Log in" : "Create tenant"}
          </button>
        </form>
      </section>
      <section className="auth-aside">
        <h1>Run the daily field-ops loop from one place.</h1>
        <p>
          Build the roster, maintain the people core, and publish assignments for
          guides, drivers, operators, and office staff.
        </p>
      </section>
    </main>
  );
}

function RosterView({
  employees,
  locations,
  locationNames,
  onCreate,
  onDelete,
  onPublish,
  shifts,
}: {
  employees: Employee[];
  locations: Location[];
  locationNames: Map<string, string>;
  onCreate: (body: CreateShiftRequest) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onPublish: (id: string) => Promise<void>;
  shifts: Shift[];
}) {
  const now = dateTimeLocalFromIso(new Date().toISOString());
  const later = dateTimeLocalFromIso(new Date(Date.now() + 3_600_000).toISOString());
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [form, setForm] = useState<FormState>({
    employee_id: "",
    ends_at: later,
    location_id: "",
    notes: "",
    starts_at: now,
  });

  function update(key: string, value: string): void {
    setForm((current) => ({ ...current, [key]: value }));
  }

  async function submit(event: FormEvent<HTMLFormElement>): Promise<void> {
    event.preventDefault();
    await onCreate({
      employee_id: form.employee_id,
      ends_at: toIsoFromLocal(form.ends_at),
      location_id: optionalNullable(form.location_id),
      notes: optionalNullable(form.notes),
      starts_at: toIsoFromLocal(form.starts_at),
    });
    setForm((current) => ({ ...current, notes: "" }));
    setDrawerOpen(false);
  }

  const days = weekDays();
  const today = new Date();
  const draftCount = shifts.filter((shift) => shift.status === "draft").length;
  const activeEmployees = employees.filter((employee) => employee.status === "active");
  const displayedEmployees = activeEmployees.length ? activeEmployees : employees;

  function openDrawer(employeeId?: string, day?: Date): void {
    setForm((current) => {
      const start = day ? new Date(day) : new Date();
      start.setHours(10, 0, 0, 0);
      const end = new Date(start);
      end.setHours(start.getHours() + 3);
      return {
        ...current,
        employee_id: employeeId ?? current.employee_id,
        starts_at: day ? dateTimeLocalFromIso(start.toISOString()) : current.starts_at,
        ends_at: day ? dateTimeLocalFromIso(end.toISOString()) : current.ends_at,
      };
    });
    setDrawerOpen(true);
  }

  return (
    <section className="roster-screen">
      <div className="roster-head">
        <div>
          <h1>Roster</h1>
          <div className="week-selector">
            <button className="icon-btn" type="button">‹</button>
            <span>
              Week of{" "}
              {new Intl.DateTimeFormat(undefined, {
                day: "2-digit",
                month: "short",
              }).format(days[0])}{" "}
              -{" "}
              {new Intl.DateTimeFormat(undefined, {
                day: "2-digit",
                month: "short",
                year: "numeric",
              }).format(days[6])}
            </span>
            <button className="icon-btn" type="button">›</button>
          </div>
        </div>
        <div className="roster-actions">
          {draftCount ? <span className="chip orange">{draftCount} unpublished</span> : null}
          <button
            className="secondary-btn"
            onClick={() => openDrawer()}
            type="button"
          >
            <Icon name="plus" size={15} />
            Add Shift
          </button>
        </div>
      </div>

      <div className="roster-board">
        <div className="roster-board-header">
          <div className="staff-col">Staff · {displayedEmployees.length}</div>
          {days.map((day) => {
            const header = formatDayHeader(day);
            const isToday = sameDay(day, today);
            return (
              <div className={isToday ? "day-head today" : "day-head"} key={day.toISOString()}>
                <span>{header.dow}</span>
                <strong>{header.day}</strong>
              </div>
            );
          })}
        </div>

        {displayedEmployees.length ? (
          displayedEmployees.map((employee, index) => {
            const employeeShifts = shifts.filter(
              (shift) => shift.employee_id === employee.id,
            );
            return (
              <div className="roster-row" key={employee.id}>
                <div className="staff-cell">
                  <div
                    className="avatar"
                    style={{ background: shiftColors[index % shiftColors.length] }}
                  >
                    {initials(employee.full_name)}
                  </div>
                  <div>
                    <strong>{employee.full_name}</strong>
                    <span>{humanize(employee.employment_type)}</span>
                  </div>
                </div>
                {days.map((day) => {
                  const dayShifts = employeeShifts.filter((shift) =>
                    sameDay(new Date(shift.starts_at), day),
                  );
                  return (
                    <div
                      className={sameDay(day, today) ? "shift-cell today" : "shift-cell"}
                      key={`${employee.id}-${day.toISOString()}`}
                    >
                      {dayShifts.length ? (
                        dayShifts.map((shift) => {
                          const color = colorForId(shift.location_id ?? shift.id);
                          return (
                            <div
                              className={
                                shift.status === "draft"
                                  ? "shift-chip draft"
                                  : "shift-chip"
                              }
                              key={shift.id}
                              style={{ "--shift-color": color } as CSSProperties}
                            >
                              <strong>
                                {shift.location_id
                                  ? locationNames.get(shift.location_id) ??
                                    "Assigned shift"
                                  : "Assigned shift"}
                              </strong>
                              <span>
                                {formatTime(shift.starts_at)} - {formatTime(shift.ends_at)}
                              </span>
                              <div className="shift-actions">
                                {shift.status === "draft" ? (
                                  <button
                                    onClick={() => void onPublish(shift.id)}
                                    type="button"
                                  >
                                    Publish
                                  </button>
                                ) : null}
                                <button
                                  onClick={() => void onDelete(shift.id)}
                                  type="button"
                                >
                                  Delete
                                </button>
                              </div>
                            </div>
                          );
                        })
                      ) : (
                        <button
                          className="empty-shift"
                          onClick={() => openDrawer(employee.id, day)}
                          type="button"
                        >
                          <Icon name="plus" size={15} />
                        </button>
                      )}
                    </div>
                  );
                })}
              </div>
            );
          })
        ) : (
          <div className="roster-empty">Create employees before building the roster.</div>
        )}
      </div>

      {drawerOpen ? (
        <>
          <button
            aria-label="Close create shift drawer"
            className="drawer-scrim"
            onClick={() => setDrawerOpen(false)}
            type="button"
          />
          <aside className="shift-drawer">
            <div className="drawer-head">
              <h2>Add Shift</h2>
              <button
                className="icon-btn"
                onClick={() => setDrawerOpen(false)}
                type="button"
              >
                <Icon name="x" size={16} />
              </button>
            </div>
            <form className="drawer-body" onSubmit={(event) => void submit(event)}>
              <label>
                Staff member
                <select
                  onChange={(event) => update("employee_id", event.target.value)}
                  required
                  value={form.employee_id}
                >
                  <option value="">Select employee</option>
                  {employees.map((employee) => (
                    <option key={employee.id} value={employee.id}>
                      {employee.full_name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Location / assignment
                <select
                  onChange={(event) => update("location_id", event.target.value)}
                  value={form.location_id}
                >
                  <option value="">Unassigned</option>
                  {locations.map((location) => (
                    <option key={location.id} value={location.id}>
                      {location.name}
                    </option>
                  ))}
                </select>
              </label>
              <div className="form-grid two-col">
                <label>
                  Starts
                  <input
                    onChange={(event) => update("starts_at", event.target.value)}
                    required
                    type="datetime-local"
                    value={form.starts_at}
                  />
                </label>
                <label>
                  Ends
                  <input
                    onChange={(event) => update("ends_at", event.target.value)}
                    required
                    type="datetime-local"
                    value={form.ends_at}
                  />
                </label>
              </div>
              <label>
                Notes
                <textarea
                  onChange={(event) => update("notes", event.target.value)}
                  value={form.notes}
                />
              </label>
              <div className="drawer-note">
                Draft shifts stay internal until published.
              </div>
              <div className="drawer-actions">
                <button
                  className="ghost-btn"
                  onClick={() => setDrawerOpen(false)}
                  type="button"
                >
                  Cancel
                </button>
                <button
                  className="primary-btn"
                  disabled={!employees.length}
                  type="submit"
                >
                  <Icon name="plus" size={15} />
                  Add Draft Shift
                </button>
              </div>
            </form>
          </aside>
        </>
      ) : null}

      {shifts.length ? (
        <div className="roster-legend">
          <span>Assignments:</span>
          {locations.slice(0, 6).map((location) => (
            <span key={location.id}>
              <i style={{ background: colorForId(location.id) }} />
              {location.name}
            </span>
          ))}
          {!locations.length ? <span>No locations created yet</span> : null}
        </div>
      ) : null}
    </section>
  );
}

function PeopleView({
  departmentNames,
  departments,
  employees,
  onCreate,
  onDelete,
  roleNames,
  roles,
}: {
  departmentNames: Map<string, string>;
  departments: Department[];
  employees: Employee[];
  onCreate: (body: CreateEmployeeRequest) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  roleNames: Map<string, string>;
  roles: Role[];
}) {
  const [form, setForm] = useState<FormState>({
    department_id: "",
    email: "",
    employment_type: "full_time",
    full_name: "",
    hired_at: "",
    phone: "",
    role_id: "",
    status: "active",
    title: "",
  });

  function update(key: string, value: string): void {
    setForm((current) => ({ ...current, [key]: value }));
  }

  async function submit(event: FormEvent<HTMLFormElement>): Promise<void> {
    event.preventDefault();
    await onCreate({
      department_id: optionalNullable(form.department_id),
      email: optionalNullable(form.email),
      employment_type: form.employment_type as Employee["employment_type"],
      full_name: form.full_name,
      hired_at: optionalNullable(form.hired_at),
      phone: optionalNullable(form.phone),
      role_id: optionalNullable(form.role_id),
      status: form.status as Employee["status"],
      title: optionalNullable(form.title),
    });
    setForm((current) => ({
      ...current,
      email: "",
      full_name: "",
      phone: "",
      title: "",
    }));
  }

  return (
    <section className="page-grid">
      <div className="card">
        <div className="section-head">
          <h2>Employees</h2>
          <span className="table-note">{employees.length} total</span>
        </div>
        {employees.length ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Department</th>
                  <th>Role</th>
                  <th>Employment</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {employees.map((employee) => (
                  <tr key={employee.id}>
                    <td>
                      <strong>{employee.full_name}</strong>
                      <div className="table-note">{employee.email ?? employee.phone ?? ""}</div>
                    </td>
                    <td>
                      {employee.department_id
                        ? departmentNames.get(employee.department_id) ?? "Unknown"
                        : "Unassigned"}
                    </td>
                    <td>
                      {employee.role_id
                        ? roleNames.get(employee.role_id) ?? "Unknown"
                        : "None"}
                    </td>
                    <td>{humanize(employee.employment_type)}</td>
                    <td>
                      <span className={`pill ${employee.status === "inactive" ? "warn" : ""}`}>
                        {employee.status}
                      </span>
                    </td>
                    <td>
                      <div className="row-actions">
                        <button
                          className="danger-btn"
                          onClick={() => void onDelete(employee.id)}
                          type="button"
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="empty-state">No employees yet.</div>
        )}
      </div>
      <div className="card">
        <h3>Add employee</h3>
        <form className="form-grid" onSubmit={(event) => void submit(event)}>
          <label>
            Full name
            <input
              onChange={(event) => update("full_name", event.target.value)}
              required
              value={form.full_name}
            />
          </label>
          <div className="form-grid two-col">
            <label>
              Department
              <select
                onChange={(event) => update("department_id", event.target.value)}
                value={form.department_id}
              >
                <option value="">Unassigned</option>
                {departments.map((department) => (
                  <option key={department.id} value={department.id}>
                    {department.name}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Role
              <select
                onChange={(event) => update("role_id", event.target.value)}
                value={form.role_id}
              >
                <option value="">None</option>
                {roles.map((role) => (
                  <option key={role.id} value={role.id}>
                    {role.name}
                  </option>
                ))}
              </select>
            </label>
          </div>
          <div className="form-grid two-col">
            <label>
              Employment
              <select
                onChange={(event) =>
                  update("employment_type", event.target.value)
                }
                value={form.employment_type}
              >
                {employmentTypes.map((type) => (
                  <option key={type} value={type}>
                    {humanize(type)}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Status
              <select
                onChange={(event) => update("status", event.target.value)}
                value={form.status}
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </label>
          </div>
          <label>
            Title
            <input
              onChange={(event) => update("title", event.target.value)}
              value={form.title}
            />
          </label>
          <div className="form-grid two-col">
            <label>
              Email
              <input
                onChange={(event) => update("email", event.target.value)}
                type="email"
                value={form.email}
              />
            </label>
            <label>
              Phone
              <input
                onChange={(event) => update("phone", event.target.value)}
                value={form.phone}
              />
            </label>
          </div>
          <label>
            Hired at
            <input
              onChange={(event) => update("hired_at", event.target.value)}
              type="date"
              value={form.hired_at}
            />
          </label>
          <button className="primary-btn" type="submit">
            Add employee
          </button>
        </form>
      </div>
    </section>
  );
}

function DepartmentsView({
  departments,
  onCreate,
  onDelete,
}: {
  departments: Department[];
  onCreate: (body: CreateDepartmentRequest) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
}) {
  const [form, setForm] = useState<FormState>({ name: "", parent_id: "" });

  async function submit(event: FormEvent<HTMLFormElement>): Promise<void> {
    event.preventDefault();
    await onCreate({
      name: form.name,
      parent_id: optionalNullable(form.parent_id),
    });
    setForm({ name: "", parent_id: "" });
  }

  return (
    <SimpleResourceView
      columns={["Name", "Parent", "Created", "Actions"]}
      empty="No departments yet."
      form={
        <form className="form-grid" onSubmit={(event) => void submit(event)}>
          <label>
            Name
            <input
              onChange={(event) =>
                setForm((current) => ({ ...current, name: event.target.value }))
              }
              required
              value={form.name}
            />
          </label>
          <label>
            Parent department
            <select
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  parent_id: event.target.value,
                }))
              }
              value={form.parent_id}
            >
              <option value="">None</option>
              {departments.map((department) => (
                <option key={department.id} value={department.id}>
                  {department.name}
                </option>
              ))}
            </select>
          </label>
          <button className="primary-btn" type="submit">
            Create department
          </button>
        </form>
      }
      formTitle="Create department"
      title="Departments"
    >
      {departments.map((department) => (
        <tr key={department.id}>
          <td>{department.name}</td>
          <td>
            {department.parent_id
              ? departments.find((item) => item.id === department.parent_id)?.name ??
                "Unknown"
              : "None"}
          </td>
          <td>{formatDateTime(department.created_at)}</td>
          <td>
            <div className="row-actions">
              <button
                className="danger-btn"
                onClick={() => void onDelete(department.id)}
                type="button"
              >
                Delete
              </button>
            </div>
          </td>
        </tr>
      ))}
    </SimpleResourceView>
  );
}

function RolesView({
  onCreate,
  onDelete,
  roles,
}: {
  onCreate: (body: CreateRoleRequest) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  roles: Role[];
}) {
  const [form, setForm] = useState<FormState>({
    description: "",
    name: "",
    permissions: "",
  });

  async function submit(event: FormEvent<HTMLFormElement>): Promise<void> {
    event.preventDefault();
    await onCreate({
      description: optionalNullable(form.description),
      name: form.name,
      permissions: form.permissions
        .split(",")
        .map((permission) => permission.trim())
        .filter(isNonEmpty),
    });
    setForm({ description: "", name: "", permissions: "" });
  }

  return (
    <SimpleResourceView
      columns={["Name", "Description", "Permissions", "Actions"]}
      empty="No roles yet."
      form={
        <form className="form-grid" onSubmit={(event) => void submit(event)}>
          <label>
            Name
            <input
              onChange={(event) =>
                setForm((current) => ({ ...current, name: event.target.value }))
              }
              required
              value={form.name}
            />
          </label>
          <label>
            Description
            <textarea
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  description: event.target.value,
                }))
              }
              value={form.description}
            />
          </label>
          <label>
            Permissions
            <input
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  permissions: event.target.value,
                }))
              }
              placeholder="employees.read, shifts.publish"
              value={form.permissions}
            />
          </label>
          <button className="primary-btn" type="submit">
            Create role
          </button>
        </form>
      }
      formTitle="Create role"
      title="Roles"
    >
      {roles.map((role) => (
        <tr key={role.id}>
          <td>{role.name}</td>
          <td>{role.description ?? "None"}</td>
          <td>{role.permissions.length ? role.permissions.join(", ") : "None"}</td>
          <td>
            <div className="row-actions">
              <button
                className="danger-btn"
                onClick={() => void onDelete(role.id)}
                type="button"
              >
                Delete
              </button>
            </div>
          </td>
        </tr>
      ))}
    </SimpleResourceView>
  );
}

function LocationsView({
  locations,
  onCreate,
  onDelete,
}: {
  locations: Location[];
  onCreate: (body: CreateLocationRequest) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
}) {
  const [form, setForm] = useState<FormState>({
    address: "",
    lat: "",
    lng: "",
    name: "",
  });

  async function submit(event: FormEvent<HTMLFormElement>): Promise<void> {
    event.preventDefault();
    await onCreate({
      address: optionalNullable(form.address),
      lat: optionalNumber(form.lat),
      lng: optionalNumber(form.lng),
      name: form.name,
    });
    setForm({ address: "", lat: "", lng: "", name: "" });
  }

  return (
    <SimpleResourceView
      columns={["Name", "Address", "Coordinates", "Actions"]}
      empty="No locations yet."
      form={
        <form className="form-grid" onSubmit={(event) => void submit(event)}>
          <label>
            Name
            <input
              onChange={(event) =>
                setForm((current) => ({ ...current, name: event.target.value }))
              }
              required
              value={form.name}
            />
          </label>
          <label>
            Address
            <textarea
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  address: event.target.value,
                }))
              }
              value={form.address}
            />
          </label>
          <div className="form-grid two-col">
            <label>
              Latitude
              <input
                onChange={(event) =>
                  setForm((current) => ({ ...current, lat: event.target.value }))
                }
                step="any"
                type="number"
                value={form.lat}
              />
            </label>
            <label>
              Longitude
              <input
                onChange={(event) =>
                  setForm((current) => ({ ...current, lng: event.target.value }))
                }
                step="any"
                type="number"
                value={form.lng}
              />
            </label>
          </div>
          <button className="primary-btn" type="submit">
            Create location
          </button>
        </form>
      }
      formTitle="Create location"
      title="Locations"
    >
      {locations.map((location) => (
        <tr key={location.id}>
          <td>{location.name}</td>
          <td>{location.address ?? "None"}</td>
          <td>
            {location.lat != null && location.lng != null
              ? `${location.lat}, ${location.lng}`
              : "None"}
          </td>
          <td>
            <div className="row-actions">
              <button
                className="danger-btn"
                onClick={() => void onDelete(location.id)}
                type="button"
              >
                Delete
              </button>
            </div>
          </td>
        </tr>
      ))}
    </SimpleResourceView>
  );
}

function SimpleResourceView({
  children,
  columns,
  empty,
  form,
  formTitle,
  title,
}: {
  children: ReactNode;
  columns: string[];
  empty: string;
  form: ReactNode;
  formTitle: string;
  title: string;
}) {
  const hasRows = Children.count(children) > 0;

  return (
    <section className="page-grid">
      <div className="card">
        <div className="section-head">
          <h2>{title}</h2>
        </div>
        {hasRows ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  {columns.map((column) => (
                    <th key={column}>{column}</th>
                  ))}
                </tr>
              </thead>
              <tbody>{children}</tbody>
            </table>
          </div>
        ) : (
          <div className="empty-state">{empty}</div>
        )}
      </div>
      <div className="card">
        <h3>{formTitle}</h3>
        {form}
      </div>
    </section>
  );
}

export { App };
