import type { CSSProperties, ReactNode } from "react";
import type { Employee, LiveViewEntry, Shift } from "../api/resources";
import { Avatar, Chip, Icon, StatusChip, colorForId, formatTime, humanize } from "../ui";
import type { ChipTone, LiveStatus } from "../ui";

const TYPE_TONE: Record<Employee["employment_type"], ChipTone> = {
  full_time: "blue",
  part_time: "neutral",
  freelance: "neutral",
  seasonal: "orange",
};

// Week (Mon–Sun) of the current date.
function weekDays(): Date[] {
  const now = new Date();
  const copy = new Date(now);
  const day = copy.getDay();
  copy.setDate(copy.getDate() + (day === 0 ? -6 : 1 - day));
  copy.setHours(0, 0, 0, 0);
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(copy);
    d.setDate(copy.getDate() + i);
    return d;
  });
}

function sameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

function deriveLiveStatus(entry: LiveViewEntry, nowMs: number): LiveStatus {
  switch (entry.attendance_status) {
    case "checked_in":
      return "working";
    case "checked_out":
      return "done";
    default:
      return new Date(entry.shift.starts_at).getTime() > nowMs ? "upcoming" : "late";
  }
}

function SectionTitle({ children, action }: { children: ReactNode; action?: ReactNode }) {
  return (
    <div style={{ display: "flex", alignItems: "center", marginBottom: 12 }}>
      <span style={{ fontSize: 13, fontWeight: 700, color: "var(--adaptive-700)", textTransform: "uppercase", letterSpacing: ".05em" }}>
        {children}
      </span>
      {action}
    </div>
  );
}

function FieldRow({ icon, label, value }: { icon: Parameters<typeof Icon>[0]["name"]; label: string; value: string }) {
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 11, padding: "10px 0", borderBottom: "1px solid var(--adaptive-100)" }}>
      <Icon name={icon} size={17} color="var(--adaptive-400)" />
      <span style={{ fontSize: 13, color: "var(--adaptive-500)", width: 116, flexShrink: 0 }}>{label}</span>
      <span style={{ fontSize: 13.5, fontWeight: 500, color: "var(--adaptive-900)", textAlign: "right", marginLeft: "auto" }}>{value}</span>
    </div>
  );
}

const cardStyle: CSSProperties = {
  background: "var(--card)",
  border: "1px solid var(--adaptive-200)",
  borderRadius: 8,
  padding: 18,
};

export function EmployeeDetail({
  employee,
  shifts,
  live,
  locationNames,
  departmentName,
  roleName,
  onBack,
}: {
  employee: Employee;
  shifts: Shift[];
  live: LiveViewEntry[];
  locationNames: Map<string, string>;
  departmentName: string | null;
  roleName: string | null;
  onBack: () => void;
}) {
  const nowMs = Date.now();
  const days = weekDays();
  const myShifts = shifts.filter((s) => s.employee_id === employee.id);
  const liveEntry = live.find((e) => e.employee_id === employee.id) ?? null;
  const liveStatus = liveEntry ? deriveLiveStatus(liveEntry, nowMs) : null;
  const person = { id: employee.id, name: employee.full_name };
  const locName = (id?: string | null) => (id ? locationNames.get(id) ?? "Assigned" : "Assigned");

  return (
    <div style={{ padding: "20px 24px 40px", display: "flex", flexDirection: "column", gap: 18 }}>
      <button
        onClick={onBack}
        style={{
          alignSelf: "flex-start",
          display: "inline-flex",
          alignItems: "center",
          gap: 6,
          border: 0,
          background: "none",
          cursor: "pointer",
          fontFamily: "inherit",
          fontSize: 13,
          fontWeight: 600,
          color: "var(--adaptive-500)",
        }}
      >
        <Icon name="chevron" size={15} color="var(--adaptive-400)" style={{ transform: "rotate(180deg)" }} />
        Back to Directory
      </button>

      {/* header */}
      <div style={cardStyle}>
        <div style={{ display: "flex", gap: 18, alignItems: "flex-start", flexWrap: "wrap" }}>
          <Avatar person={person} size={72} />
          <div style={{ flex: 1, minWidth: 220 }}>
            <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
              <h1 style={{ margin: 0, fontSize: 26, fontWeight: 700, letterSpacing: "-0.02em", color: "var(--adaptive-900)" }}>
                {employee.full_name}
              </h1>
              {liveStatus && <StatusChip status={liveStatus} />}
            </div>
            <div style={{ display: "flex", alignItems: "center", gap: 10, marginTop: 6, flexWrap: "wrap", fontSize: 14, color: "var(--adaptive-600)" }}>
              <span style={{ fontWeight: 600, color: "var(--adaptive-800)" }}>{roleName ?? "No role"}</span>
              <span style={{ color: "var(--adaptive-300)" }}>·</span>
              <span>{departmentName ?? "Unassigned"}</span>
              {employee.title && (
                <>
                  <span style={{ color: "var(--adaptive-300)" }}>·</span>
                  <span>{employee.title}</span>
                </>
              )}
            </div>
            <div style={{ display: "flex", gap: 7, marginTop: 12, flexWrap: "wrap" }}>
              <Chip tone={TYPE_TONE[employee.employment_type]}>{humanize(employee.employment_type)}</Chip>
              <Chip tone={employee.status === "active" ? "blue" : "neutral"}>{employee.status}</Chip>
              {employee.hired_at && (
                <Chip>
                  <Icon name="clock" size={12} />
                  Since {new Intl.DateTimeFormat(undefined, { month: "short", year: "numeric" }).format(new Date(employee.hired_at))}
                </Chip>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* live banner */}
      {liveEntry && (liveStatus === "working" || liveStatus === "late") && (
        <div
          style={{
            ...cardStyle,
            padding: "14px 18px",
            borderColor: liveStatus === "late" ? "var(--red-200)" : "var(--green-200)",
            background: liveStatus === "late" ? "var(--red-50)" : "var(--green-50)",
            display: "flex",
            alignItems: "center",
            gap: 14,
            flexWrap: "wrap",
          }}
        >
          <span
            className="opero-pulse"
            style={{ width: 9, height: 9, borderRadius: "50%", background: liveStatus === "late" ? "var(--red-500)" : "var(--green-500)" }}
          />
          <div style={{ flex: 1, minWidth: 200 }}>
            <div style={{ fontSize: 14, fontWeight: 600, color: "var(--adaptive-900)" }}>
              {liveStatus === "late" ? "Not checked in" : "On shift now"} · {locName(liveEntry.shift.location_id)}
            </div>
            <div style={{ fontSize: 12.5, color: "var(--adaptive-600)", marginTop: 2 }}>
              {liveEntry.check_in_at ? `Checked in ${formatTime(liveEntry.check_in_at)}` : `Scheduled ${formatTime(liveEntry.shift.starts_at)}`}
            </div>
          </div>
        </div>
      )}

      {/* two-column grid */}
      <div style={{ display: "grid", gridTemplateColumns: "minmax(0, 1.55fr) minmax(0, 1fr)", gap: 16, alignItems: "start" }}>
        {/* left: week schedule */}
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <div style={cardStyle}>
            <SectionTitle>This Week's Schedule</SectionTitle>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(7, 1fr)", gap: 8 }}>
              {days.map((d, i) => {
                const isToday = sameDay(d, new Date());
                const dayShifts = myShifts.filter((s) => sameDay(new Date(s.starts_at), d));
                return (
                  <div
                    key={i}
                    style={{
                      borderRadius: 8,
                      border: `1px solid ${isToday ? "var(--primary-300)" : "var(--adaptive-200)"}`,
                      background: isToday ? "var(--primary-50)" : "var(--card)",
                      overflow: "hidden",
                      minHeight: 92,
                    }}
                  >
                    <div style={{ textAlign: "center", padding: "6px 0", borderBottom: "1px solid var(--adaptive-100)" }}>
                      <div
                        style={{
                          fontSize: 10.5,
                          fontWeight: 600,
                          textTransform: "uppercase",
                          letterSpacing: ".04em",
                          color: isToday ? "var(--primary-600)" : "var(--adaptive-500)",
                        }}
                      >
                        {new Intl.DateTimeFormat(undefined, { weekday: "short" }).format(d)}
                      </div>
                      <div style={{ fontSize: 14, fontWeight: 700, color: isToday ? "var(--primary-700)" : "var(--adaptive-800)" }}>
                        {d.getDate()}
                      </div>
                    </div>
                    <div style={{ padding: 6 }}>
                      {dayShifts.length === 0 ? (
                        <div style={{ textAlign: "center", fontSize: 11, color: "var(--adaptive-300)", padding: "14px 0" }}>Off</div>
                      ) : (
                        dayShifts.map((s) => {
                          const color = colorForId(s.location_id ?? s.id);
                          return (
                            <div
                              key={s.id}
                              style={{
                                borderLeft: `3px solid ${color}`,
                                background: `color-mix(in srgb, ${color} 10%, transparent)`,
                                borderRadius: 5,
                                padding: "5px 6px",
                                marginBottom: 4,
                              }}
                            >
                              <div style={{ fontSize: 10.5, fontWeight: 600, color: "var(--adaptive-900)", lineHeight: 1.25 }}>
                                {locName(s.location_id)}
                              </div>
                              <div style={{ fontSize: 10, color: "var(--adaptive-500)", marginTop: 2, fontFeatureSettings: "'tnum'" }}>
                                {formatTime(s.starts_at)}–{formatTime(s.ends_at)}
                              </div>
                              {s.status === "draft" && (
                                <div style={{ fontSize: 8.5, fontWeight: 700, color, marginTop: 2 }}>DRAFT</div>
                              )}
                            </div>
                          );
                        })
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* right: contact + employment + role */}
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <div style={cardStyle}>
            <SectionTitle>Contact</SectionTitle>
            <div>
              <FieldRow icon="send" label="Email" value={employee.email ?? "—"} />
              <FieldRow icon="phone" label="Phone" value={employee.phone ?? "—"} />
            </div>
          </div>

          <div style={cardStyle}>
            <SectionTitle>Employment</SectionTitle>
            <div>
              <FieldRow icon="briefcase" label="Type" value={humanize(employee.employment_type)} />
              <FieldRow
                icon="calendar"
                label="Joined"
                value={
                  employee.hired_at
                    ? new Intl.DateTimeFormat(undefined, { dateStyle: "medium" }).format(new Date(employee.hired_at))
                    : "—"
                }
              />
              <FieldRow icon="briefcase" label="Department" value={departmentName ?? "Unassigned"} />
              <FieldRow icon="route" label="Role" value={roleName ?? "None"} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
