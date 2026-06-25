import { Fragment, useMemo, useState } from "react";
import type { ReactNode } from "react";
import type {
  CreateEmployeeRequest,
  Department,
  Employee,
  Role,
  UpdateEmployeeRequest,
} from "../api/resources";
import {
  Avatar,
  Btn,
  Card,
  Chip,
  Drawer,
  Field,
  Icon,
  IconButton,
  PageHeader,
  SortTh,
  controlStyle,
  humanize,
  sortRows,
  useSort,
} from "../ui";
import type { ChipTone } from "../ui";

const employmentTypes: Employee["employment_type"][] = [
  "full_time",
  "part_time",
  "freelance",
  "seasonal",
];

const TYPE_TONE: Record<Employee["employment_type"], ChipTone> = {
  full_time: "blue",
  part_time: "neutral",
  freelance: "neutral",
  seasonal: "orange",
};

function MemberDrawer({
  employee,
  employees,
  departments,
  roles,
  onClose,
  onCreate,
  onUpdate,
}: {
  employee?: Employee;
  employees: Employee[];
  departments: Department[];
  roles: Role[];
  onClose: () => void;
  onCreate: (body: CreateEmployeeRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateEmployeeRequest) => Promise<void>;
}) {
  const isEdit = Boolean(employee);
  const [fullName, setFullName] = useState(employee?.full_name ?? "");
  const [departmentId, setDepartmentId] = useState(employee?.department_id ?? "");
  const [roleId, setRoleId] = useState(employee?.role_id ?? "");
  const [employmentType, setEmploymentType] = useState<Employee["employment_type"]>(
    employee?.employment_type ?? "full_time",
  );
  const [status, setStatus] = useState<Employee["status"]>(employee?.status ?? "active");
  const [title, setTitle] = useState(employee?.title ?? "");
  const [email, setEmail] = useState(employee?.email ?? "");
  const [phone, setPhone] = useState(employee?.phone ?? "");
  const [hiredAt, setHiredAt] = useState(employee?.hired_at ?? "");
  const [location, setLocation] = useState(employee?.location ?? "");
  const [languages, setLanguages] = useState((employee?.languages ?? []).join(", "));
  const [emergencyName, setEmergencyName] = useState(employee?.emergency_contact_name ?? "");
  const [emergencyPhone, setEmergencyPhone] = useState(employee?.emergency_contact_phone ?? "");
  const [reportsTo, setReportsTo] = useState(employee?.reports_to ?? "");
  const [employeeCode, setEmployeeCode] = useState(employee?.employee_code ?? "");
  const [submitting, setSubmitting] = useState(false);

  const canSubmit = fullName.trim().length > 0 && !submitting;
  // A person can't report to themselves.
  const reportsToOptions = employees.filter((e) => e.id !== employee?.id);

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      // PATCH in v1 can't clear fields back to null, so empties are omitted
      // rather than sent — same body shape for create and update.
      const langs = languages
        .split(",")
        .map((l) => l.trim())
        .filter(Boolean);
      const body = {
        full_name: fullName.trim(),
        department_id: departmentId || undefined,
        role_id: roleId || undefined,
        employment_type: employmentType,
        status,
        title: title.trim() || undefined,
        email: email.trim() || undefined,
        phone: phone.trim() || undefined,
        hired_at: hiredAt || undefined,
        location: location.trim() || undefined,
        languages: langs.length ? langs : undefined,
        emergency_contact_name: emergencyName.trim() || undefined,
        emergency_contact_phone: emergencyPhone.trim() || undefined,
        reports_to: reportsTo || undefined,
        employee_code: employeeCode.trim() || undefined,
      };
      if (employee) {
        await onUpdate(employee.id, body);
      } else {
        await onCreate(body);
      }
      onClose();
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={<div style={{ fontSize: 16, fontWeight: 600 }}>{isEdit ? "Edit member" : "Add member"}</div>}
      footer={
        <>
          <Btn variant="tertiary" onClick={onClose} style={{ marginRight: "auto" }}>
            Cancel
          </Btn>
          <Btn variant="primary" icon={isEdit ? "check" : "plus"} disabled={!canSubmit} onClick={() => void submit()}>
            {isEdit ? "Save changes" : "Add member"}
          </Btn>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Field label="Full name">
          <input value={fullName} onChange={(e) => setFullName(e.target.value)} style={controlStyle} required />
        </Field>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Department">
              <select value={departmentId} onChange={(e) => setDepartmentId(e.target.value)} style={controlStyle}>
                <option value="">Unassigned</option>
                {departments.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Role">
              <select value={roleId} onChange={(e) => setRoleId(e.target.value)} style={controlStyle}>
                <option value="">None</option>
                {roles.map((r) => (
                  <option key={r.id} value={r.id}>
                    {r.name}
                  </option>
                ))}
              </select>
            </Field>
          </div>
        </div>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Employment">
              <select
                value={employmentType}
                onChange={(e) => setEmploymentType(e.target.value as Employee["employment_type"])}
                style={controlStyle}
              >
                {employmentTypes.map((t) => (
                  <option key={t} value={t}>
                    {humanize(t)}
                  </option>
                ))}
              </select>
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Status">
              <select value={status} onChange={(e) => setStatus(e.target.value as Employee["status"])} style={controlStyle}>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </Field>
          </div>
        </div>
        <Field label="Title">
          <input value={title} onChange={(e) => setTitle(e.target.value)} style={controlStyle} />
        </Field>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Email">
              <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} style={controlStyle} />
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Phone">
              <input value={phone} onChange={(e) => setPhone(e.target.value)} style={controlStyle} />
            </Field>
          </div>
        </div>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Hired at">
              <input type="date" value={hiredAt} onChange={(e) => setHiredAt(e.target.value)} style={controlStyle} />
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Employee ID">
              <input value={employeeCode} onChange={(e) => setEmployeeCode(e.target.value)} placeholder="e.g. TT-1007" style={controlStyle} />
            </Field>
          </div>
        </div>
        <Field label="Location">
          <input value={location} onChange={(e) => setLocation(e.target.value)} placeholder="e.g. Lisbon" style={controlStyle} />
        </Field>
        <Field label="Languages (comma-separated)">
          <input value={languages} onChange={(e) => setLanguages(e.target.value)} placeholder="English, Portuguese" style={controlStyle} />
        </Field>
        <Field label="Reports to">
          <select value={reportsTo} onChange={(e) => setReportsTo(e.target.value)} style={controlStyle}>
            <option value="">None</option>
            {reportsToOptions.map((e) => (
              <option key={e.id} value={e.id}>
                {e.full_name}
              </option>
            ))}
          </select>
        </Field>
        <div style={{ display: "flex", gap: 12 }}>
          <div style={{ flex: 1 }}>
            <Field label="Emergency contact">
              <input value={emergencyName} onChange={(e) => setEmergencyName(e.target.value)} placeholder="Name" style={controlStyle} />
            </Field>
          </div>
          <div style={{ flex: 1 }}>
            <Field label="Emergency phone">
              <input value={emergencyPhone} onChange={(e) => setEmergencyPhone(e.target.value)} style={controlStyle} />
            </Field>
          </div>
        </div>
      </div>
    </Drawer>
  );
}

export function People({
  employees,
  departments,
  roles,
  departmentNames,
  roleNames,
  onOpen,
  onCreate,
  onUpdate,
  onDelete,
}: {
  employees: Employee[];
  departments: Department[];
  roles: Role[];
  departmentNames: Map<string, string>;
  roleNames: Map<string, string>;
  onOpen: (employee: Employee) => void;
  onCreate: (body: CreateEmployeeRequest) => Promise<void>;
  onUpdate: (id: string, body: UpdateEmployeeRequest) => Promise<void>;
  onDelete: (id: string) => void;
}) {
  const [tab, setTab] = useState<string>("all");
  const [groupByRole, setGroupByRole] = useState(false);
  const [adding, setAdding] = useState(false);
  const [editing, setEditing] = useState<Employee | null>(null);
  const [sort, toggleSort] = useSort("full_name");

  const tabs: { id: string; label: string; count: number }[] = [
    { id: "all", label: "All", count: employees.length },
    ...departments.map((d) => ({
      id: d.id,
      label: d.name,
      count: employees.filter((e) => e.department_id === d.id).length,
    })),
  ];
  const unassignedCount = employees.filter((e) => !e.department_id).length;
  if (unassignedCount > 0) {
    tabs.push({ id: "unassigned", label: "Unassigned", count: unassignedCount });
  }

  const filtered =
    tab === "all"
      ? employees
      : tab === "unassigned"
        ? employees.filter((e) => !e.department_id)
        : employees.filter((e) => e.department_id === tab);

  const sortAccessors: Record<string, (e: Employee) => string | number | null | undefined> = {
    full_name: (e) => e.full_name,
    role: (e) => (e.role_id ? roleNames.get(e.role_id) : ""),
    department: (e) => (e.department_id ? departmentNames.get(e.department_id) : ""),
    employment_type: (e) => e.employment_type,
    phone: (e) => e.phone ?? "",
    status: (e) => e.status,
  };
  const shown = sortRows(filtered, sort, sortAccessors);

  // When grouping by role, partition the filtered list into role sections
  // (ordered by the roles list; employees without a role go to a final group).
  const NO_ROLE = "__none";
  const roleGroups = useMemo(() => {
    const byRole = new Map<string, Employee[]>();
    for (const e of shown) {
      const key = e.role_id ?? NO_ROLE;
      const list = byRole.get(key) ?? [];
      list.push(e);
      byRole.set(key, list);
    }
    const ordered: { key: string; label: string; list: Employee[] }[] = roles
      .filter((r) => byRole.has(r.id))
      .map((r) => ({ key: r.id, label: r.name, list: byRole.get(r.id) ?? [] }));
    if (byRole.has(NO_ROLE)) {
      ordered.push({ key: NO_ROLE, label: "No role", list: byRole.get(NO_ROLE) ?? [] });
    }
    return ordered;
  }, [shown, roles]);

  function renderRow(s: Employee): ReactNode {
    return (
      <tr
        key={s.id}
        style={{ borderBottom: "1px solid var(--adaptive-100)" }}
        onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
        onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
      >
        <td style={{ padding: "10px 16px" }}>
          <button
            onClick={() => onOpen(s)}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 10,
              border: 0,
              background: "none",
              padding: 0,
              cursor: "pointer",
              fontFamily: "inherit",
              textAlign: "left",
            }}
          >
            <Avatar person={{ id: s.id, name: s.full_name }} size={32} />
            <div>
              <div style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{s.full_name}</div>
              {(() => {
                // Title is a distinct, more-specific job title — shown only when
                // set and different from the role (role has its own column, so we
                // never duplicate it here).
                const roleName = s.role_id ? roleNames.get(s.role_id) : null;
                const title = s.title?.trim();
                return title && title !== roleName ? (
                  <div style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>{title}</div>
                ) : null;
              })()}
            </div>
          </button>
        </td>
        <td style={{ padding: "10px 16px", color: "var(--adaptive-700)" }}>
          {s.role_id ? roleNames.get(s.role_id) ?? "Unknown" : "—"}
        </td>
        <td style={{ padding: "10px 16px", color: "var(--adaptive-700)" }}>
          {s.department_id ? departmentNames.get(s.department_id) ?? "Unknown" : "Unassigned"}
        </td>
        <td style={{ padding: "10px 16px" }}>
          <Chip tone={TYPE_TONE[s.employment_type]}>{humanize(s.employment_type)}</Chip>
        </td>
        <td style={{ padding: "10px 16px", color: "var(--adaptive-500)", fontFeatureSettings: "'tnum'" }}>
          {s.phone ?? "—"}
        </td>
        <td style={{ padding: "10px 16px" }}>
          <Chip tone={s.status === "active" ? "blue" : "neutral"}>{s.status}</Chip>
        </td>
        <td style={{ padding: "10px 16px" }}>
          <div style={{ display: "flex", gap: 6, justifyContent: "flex-end" }}>
            <IconButton icon="pencil" title="Edit" onClick={() => setEditing(s)} />
            <IconButton icon="x" title="Delete" tone="danger" onClick={() => onDelete(s.id)} />
          </div>
        </td>
      </tr>
    );
  }

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 16 }}>
      <PageHeader
        title="Directory"
        subtitle={`${employees.length} people across ${departments.length} department${
          departments.length === 1 ? "" : "s"
        }`}
        actions={
          <>
            <div
              style={{
                display: "inline-flex",
                background: "var(--adaptive-100)",
                borderRadius: 7,
                padding: 3,
                gap: 2,
              }}
            >
              {([
                { id: false, label: "List" },
                { id: true, label: "By role" },
              ] as const).map((o) => {
                const on = o.id === groupByRole;
                return (
                  <button
                    key={o.label}
                    onClick={() => setGroupByRole(o.id)}
                    style={{
                      display: "inline-flex",
                      alignItems: "center",
                      gap: 6,
                      padding: "5px 11px",
                      borderRadius: 5,
                      border: 0,
                      cursor: "pointer",
                      fontFamily: "inherit",
                      fontSize: 12.5,
                      fontWeight: 600,
                      background: on ? "var(--card)" : "transparent",
                      color: on ? "var(--adaptive-900)" : "var(--adaptive-500)",
                      boxShadow: on ? "var(--shadow-xs)" : "none",
                      transition: "all .15s",
                    }}
                  >
                    <Icon name={o.id ? "route" : "list"} size={15} color={on ? "var(--primary-600)" : "var(--adaptive-400)"} />
                    {o.label}
                  </button>
                );
              })}
            </div>
            <Btn variant="primary" icon="plus" onClick={() => setAdding(true)}>
              Add member
            </Btn>
          </>
        }
      />

      <div style={{ display: "flex", gap: 4, borderBottom: "1px solid var(--adaptive-200)", flexWrap: "wrap" }}>
        {tabs.map((t) => {
          const on = t.id === tab;
          return (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              style={{
                padding: "9px 14px",
                border: 0,
                background: "none",
                cursor: "pointer",
                fontFamily: "inherit",
                fontSize: 13.5,
                fontWeight: 600,
                color: on ? "var(--primary-700)" : "var(--adaptive-500)",
                borderBottom: `2px solid ${on ? "var(--primary-600)" : "transparent"}`,
                marginBottom: -1,
                display: "flex",
                alignItems: "center",
                gap: 7,
              }}
            >
              {t.label}
              <span
                style={{
                  fontSize: 11,
                  fontWeight: 700,
                  padding: "1px 7px",
                  borderRadius: 9999,
                  background: on ? "var(--primary-50)" : "var(--adaptive-100)",
                  color: on ? "var(--primary-700)" : "var(--adaptive-500)",
                }}
              >
                {t.count}
              </span>
            </button>
          );
        })}
      </div>

      <Card style={{ overflow: "hidden" }}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13, minWidth: 820 }}>
            <thead>
              <tr>
                {(
                  [
                    ["Name", "full_name"],
                    ["Role", "role"],
                    ["Department", "department"],
                    ["Employment", "employment_type"],
                    ["Phone", "phone"],
                    ["Status", "status"],
                    ["", null],
                  ] as const
                ).map(([label, key], i) => (
                  <SortTh key={i} label={label} sortKey={key} sort={sort} onSort={toggleSort} align={key === null ? "right" : "left"} />
                ))}
              </tr>
            </thead>
            <tbody>
              {shown.length === 0 ? (
                <tr>
                  <td colSpan={7} style={{ padding: 28, textAlign: "center", color: "var(--adaptive-500)" }}>
                    No people in this view yet.
                  </td>
                </tr>
              ) : groupByRole ? (
                roleGroups.map((g) => (
                  <Fragment key={g.key}>
                    <tr>
                      <td
                        colSpan={7}
                        style={{
                          padding: "7px 16px",
                          fontSize: 11,
                          fontWeight: 700,
                          letterSpacing: ".06em",
                          textTransform: "uppercase",
                          color: "var(--adaptive-500)",
                          background: "var(--adaptive-50)",
                          borderBottom: "1px solid var(--adaptive-100)",
                        }}
                      >
                        {g.label} · {g.list.length}
                      </td>
                    </tr>
                    {g.list.map(renderRow)}
                  </Fragment>
                ))
              ) : (
                shown.map(renderRow)
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {(adding || editing) && (
        <MemberDrawer
          employee={editing ?? undefined}
          employees={employees}
          departments={departments}
          roles={roles}
          onClose={() => {
            setAdding(false);
            setEditing(null);
          }}
          onCreate={onCreate}
          onUpdate={onUpdate}
        />
      )}
    </div>
  );
}
