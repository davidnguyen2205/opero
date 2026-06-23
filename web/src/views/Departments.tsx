import { useState } from "react";
import type {
  CreateDepartmentRequest,
  Department,
  Employee,
} from "../api/resources";
import {
  Avatar,
  AvatarStack,
  Btn,
  Card,
  Chip,
  Drawer,
  DrawerSectionLabel,
  Field,
  Icon,
  PageHeader,
  colorForId,
  controlStyle,
  humanize,
} from "../ui";
import type { IconName, Person } from "../ui";

// Departments in the API carry no icon/color; derive a stable presentational
// one from the id so cards read as distinct without inventing data.
const DEPT_ICONS: IconName[] = ["briefcase", "route", "activity", "users", "grid", "pin"];
function deptIcon(id: string): IconName {
  let total = 0;
  for (const c of id) total += c.charCodeAt(0);
  return DEPT_ICONS[total % DEPT_ICONS.length];
}

function toPeople(employees: Employee[]): Person[] {
  return employees.map((e) => ({ id: e.id, name: e.full_name }));
}

function DeptDrawer({
  dept,
  members,
  parentName,
  onClose,
  onDelete,
}: {
  dept: Department;
  members: Employee[];
  parentName: string | null;
  onClose: () => void;
  onDelete: (id: string) => void;
}) {
  const color = colorForId(dept.id);
  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={
        <>
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: 9,
              background: `color-mix(in srgb, ${color} 12%, transparent)`,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              flexShrink: 0,
            }}
          >
            <Icon name={deptIcon(dept.id)} size={20} color={color} />
          </div>
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 16, fontWeight: 600, color: "var(--adaptive-900)" }}>{dept.name}</div>
            <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>
              {members.length} {members.length === 1 ? "person" : "people"}
              {parentName ? ` · in ${parentName}` : ""}
            </div>
          </div>
        </>
      }
      footer={
        <Btn
          variant="secondary"
          icon="x"
          style={{ flex: 1, color: "var(--red-700)", borderColor: "var(--red-200)" }}
          onClick={() => {
            onDelete(dept.id);
            onClose();
          }}
        >
          Delete department
        </Btn>
      }
    >
      <div>
        <DrawerSectionLabel>Members · {members.length}</DrawerSectionLabel>
        {members.length === 0 ? (
          <div style={{ fontSize: 13, color: "var(--adaptive-500)" }}>No members assigned yet.</div>
        ) : (
          <div style={{ display: "flex", flexDirection: "column", gap: 2 }}>
            {members.map((s) => (
              <div key={s.id} style={{ display: "flex", alignItems: "center", gap: 10, padding: "7px 0" }}>
                <Avatar person={{ id: s.id, name: s.full_name }} size={30} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 600, color: "var(--adaptive-900)" }}>{s.full_name}</div>
                  {s.title && <div style={{ fontSize: 11.5, color: "var(--adaptive-500)" }}>{s.title}</div>}
                </div>
                <Chip>{humanize(s.employment_type)}</Chip>
              </div>
            ))}
          </div>
        )}
      </div>
    </Drawer>
  );
}

function AddDeptDrawer({
  departments,
  onClose,
  onCreate,
}: {
  departments: Department[];
  onClose: () => void;
  onCreate: (body: CreateDepartmentRequest) => Promise<void>;
}) {
  const [name, setName] = useState("");
  const [parentId, setParentId] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const canSubmit = name.trim().length > 0 && !submitting;

  async function submit() {
    if (!canSubmit) return;
    setSubmitting(true);
    try {
      await onCreate({ name: name.trim(), parent_id: parentId || undefined });
      onClose();
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Drawer
      onClose={onClose}
      width={420}
      header={<div style={{ fontSize: 16, fontWeight: 600 }}>Add department</div>}
      footer={
        <>
          <Btn variant="tertiary" onClick={onClose} style={{ marginRight: "auto" }}>
            Cancel
          </Btn>
          <Btn variant="primary" icon="plus" disabled={!canSubmit} onClick={() => void submit()}>
            Create department
          </Btn>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Field label="Name">
          <input value={name} onChange={(e) => setName(e.target.value)} style={controlStyle} required />
        </Field>
        <Field label="Parent department">
          <select value={parentId} onChange={(e) => setParentId(e.target.value)} style={controlStyle}>
            <option value="">None</option>
            {departments.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name}
              </option>
            ))}
          </select>
        </Field>
      </div>
    </Drawer>
  );
}

export function Departments({
  departments,
  employees,
  onCreate,
  onDelete,
}: {
  departments: Department[];
  employees: Employee[];
  onCreate: (body: CreateDepartmentRequest) => Promise<void>;
  onDelete: (id: string) => void;
}) {
  const [sel, setSel] = useState<Department | null>(null);
  const [adding, setAdding] = useState(false);
  const membersOf = (d: Department) => employees.filter((e) => e.department_id === d.id);
  const nameById = new Map(departments.map((d) => [d.id, d.name]));

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Departments"
        subtitle={`${departments.length} department${departments.length === 1 ? "" : "s"} · ${
          employees.length
        } people`}
        actions={
          <Btn variant="primary" icon="plus" onClick={() => setAdding(true)}>
            Add department
          </Btn>
        }
      />

      {departments.length === 0 ? (
        <Card style={{ padding: 28 }}>
          <div style={{ textAlign: "center", color: "var(--adaptive-500)" }}>No departments yet.</div>
        </Card>
      ) : (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(330px, 1fr))", gap: 16 }}>
          {departments.map((dep) => {
            const members = membersOf(dep);
            const color = colorForId(dep.id);
            return (
              <Card
                key={dep.id}
                hover
                onClick={() => setSel(dep)}
                style={{ padding: 18, borderTop: `3px solid ${color}` }}
              >
                <div style={{ display: "flex", alignItems: "flex-start", gap: 12 }}>
                  <div
                    style={{
                      width: 42,
                      height: 42,
                      borderRadius: 10,
                      background: `color-mix(in srgb, ${color} 12%, transparent)`,
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "center",
                      flexShrink: 0,
                    }}
                  >
                    <Icon name={deptIcon(dep.id)} size={21} color={color} />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 16, fontWeight: 600, color: "var(--adaptive-900)" }}>{dep.name}</div>
                    <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>
                      {members.length} {members.length === 1 ? "person" : "people"}
                    </div>
                  </div>
                </div>
                {dep.parent_id && (
                  <div style={{ marginTop: 12 }}>
                    <Chip>in {nameById.get(dep.parent_id) ?? "Unknown"}</Chip>
                  </div>
                )}
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 10,
                    marginTop: 16,
                    paddingTop: 14,
                    borderTop: "1px solid var(--adaptive-100)",
                    minHeight: 30,
                  }}
                >
                  {members.length ? (
                    <AvatarStack people={toPeople(members)} />
                  ) : (
                    <span style={{ fontSize: 12.5, color: "var(--adaptive-400)" }}>No members yet</span>
                  )}
                </div>
              </Card>
            );
          })}
        </div>
      )}

      {sel && (
        <DeptDrawer
          dept={sel}
          members={membersOf(sel)}
          parentName={sel.parent_id ? nameById.get(sel.parent_id) ?? null : null}
          onClose={() => setSel(null)}
          onDelete={onDelete}
        />
      )}
      {adding && (
        <AddDeptDrawer departments={departments} onClose={() => setAdding(false)} onCreate={onCreate} />
      )}
    </div>
  );
}
