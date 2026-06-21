import { Children, useCallback, useEffect, useMemo, useState } from "react";
import type { FormEvent, ReactNode } from "react";
import {
  authApi,
  departmentsApi,
  employeesApi,
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
  type Location,
  type Role,
  type Shift,
} from "./api/resources";

type View = "roster" | "people" | "departments" | "roles" | "locations";
type AuthMode = "login" | "signup";

type FormState = Record<string, string>;

const views: { id: View; label: string }[] = [
  { id: "roster", label: "Roster" },
  { id: "people", label: "People" },
  { id: "departments", label: "Departments" },
  { id: "roles", label: "Roles" },
  { id: "locations", label: "Locations" },
];

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

function employeeNamesById(items: Employee[]): Map<string, string> {
  return new Map(items.map((item) => [item.id, item.full_name]));
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
      const [me, deps, emps, locs, roleList, shiftList] = await Promise.all([
        authApi.me(),
        departmentsApi.list(),
        employeesApi.list(),
        locationsApi.list(),
        rolesApi.list(),
        shiftsApi.list(),
      ]);
      setCurrentUser(me);
      setDepartments(deps);
      setEmployees(emps);
      setLocations(locs);
      setRoles(roleList);
      setShifts(shiftList);
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
  const employeeNames = useMemo(() => employeeNamesById(employees), [employees]);

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
          <strong>Opero</strong>
          <span>{currentUser?.tenant.name ?? auth.tenant.name}</span>
        </div>
        <nav className="nav-list" aria-label="Primary">
          {views.map((item) => (
            <button
              className={view === item.id ? "active" : ""}
              key={item.id}
              onClick={() => setView(item.id)}
              type="button"
            >
              {item.label}
              <span>{countForView(item.id, { departments, employees, locations, roles, shifts })}</span>
            </button>
          ))}
        </nav>
        <div className="session-card">
          <span>Signed in as</span>
          <strong>{currentUser?.user.email ?? auth.user.email}</strong>
          <span>{humanize(currentUser?.user.role ?? auth.user.role)}</span>
          <button className="secondary-btn" onClick={signOut} type="button">
            Sign out
          </button>
        </div>
      </aside>

      <main className="content">
        <header className="topbar">
          <div>
            <h1>{views.find((item) => item.id === view)?.label}</h1>
            <p className="muted">
              Manage field operations against the current OpenAPI contract.
            </p>
          </div>
          <button
            className="secondary-btn"
            disabled={loading}
            onClick={() => void loadData()}
            type="button"
          >
            {loading ? "Refreshing" : "Refresh"}
          </button>
        </header>

        <section className="dashboard-grid" aria-label="Operational summary">
          <Metric label="Active employees" value={activeEmployees.length} />
          <Metric label="Departments" value={departments.length} />
          <Metric label="Published shifts" value={publishedShifts.length} />
          <Metric label="Locations" value={locations.length} />
        </section>

        {error ? <div className="error-banner">{error}</div> : null}
        {notice ? <div className="notice-banner">{notice}</div> : null}

        {view === "roster" ? (
          <RosterView
            employeeNames={employeeNames}
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
  employeeNames,
  locations,
  locationNames,
  onCreate,
  onDelete,
  onPublish,
  shifts,
}: {
  employees: Employee[];
  employeeNames: Map<string, string>;
  locations: Location[];
  locationNames: Map<string, string>;
  onCreate: (body: CreateShiftRequest) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onPublish: (id: string) => Promise<void>;
  shifts: Shift[];
}) {
  const now = dateTimeLocalFromIso(new Date().toISOString());
  const later = dateTimeLocalFromIso(new Date(Date.now() + 3_600_000).toISOString());
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
  }

  return (
    <section className="page-grid roster-grid">
      <div className="card">
        <div className="section-head">
          <h2>Shifts</h2>
          <span className="table-note">{shifts.length} total</span>
        </div>
        {shifts.length ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Employee</th>
                  <th>Location</th>
                  <th>Start</th>
                  <th>End</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {shifts.map((shift) => (
                  <tr key={shift.id}>
                    <td>{employeeNames.get(shift.employee_id) ?? "Unknown employee"}</td>
                    <td>
                      {shift.location_id
                        ? locationNames.get(shift.location_id) ?? "Unknown location"
                        : "Unassigned"}
                    </td>
                    <td>{formatDateTime(shift.starts_at)}</td>
                    <td>{formatDateTime(shift.ends_at)}</td>
                    <td>
                      <span className={`pill ${shift.status === "draft" ? "warn" : ""}`}>
                        {shift.status}
                      </span>
                    </td>
                    <td>
                      <div className="row-actions">
                        {shift.status === "draft" ? (
                          <button
                            className="secondary-btn"
                            onClick={() => void onPublish(shift.id)}
                            type="button"
                          >
                            Publish
                          </button>
                        ) : null}
                        <button
                          className="danger-btn"
                          onClick={() => void onDelete(shift.id)}
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
          <div className="empty-state">No shifts yet.</div>
        )}
      </div>
      <div className="card form-card roster-form-card">
        <h3>Create shift</h3>
        <form className="form-grid" onSubmit={(event) => void submit(event)}>
          <label>
            Employee
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
            Location
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
          <div className="form-grid two-col date-time-grid">
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
          <label className="span-all">
            Notes
            <textarea
              onChange={(event) => update("notes", event.target.value)}
              value={form.notes}
            />
          </label>
          <button className="primary-btn" disabled={!employees.length} type="submit">
            Create draft
          </button>
        </form>
      </div>
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
