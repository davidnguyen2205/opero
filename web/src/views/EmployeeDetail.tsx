import { useEffect, useState } from "react";
import type { CSSProperties, ReactNode } from "react";
import {
  attendanceApi,
  type AttendanceRecord,
  type Employee,
  type LeaveRequest,
  type LiveViewEntry,
  type Shift,
} from "../api/resources";
import { Avatar, Btn, Chip, Icon, StatusChip, colorForId, formatTime, humanize } from "../ui";
import type { ChipTone, IconName, LiveStatus } from "../ui";

const TYPE_TONE: Record<Employee["employment_type"], ChipTone> = {
  full_time: "blue",
  part_time: "neutral",
  freelance: "neutral",
  seasonal: "orange",
};

function weekStart(now = new Date()): Date {
  const d = new Date(now);
  const day = d.getDay();
  d.setDate(d.getDate() + (day === 0 ? -6 : 1 - day));
  d.setHours(0, 0, 0, 0);
  return d;
}

function weekDays(): Date[] {
  const start = weekStart();
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(start);
    d.setDate(start.getDate() + i);
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

const cardStyle: CSSProperties = {
  background: "var(--card)",
  border: "1px solid var(--adaptive-200)",
  borderRadius: 8,
  padding: 18,
};

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

function StatCard({ label, value, sub, accent }: { label: string; value: string; sub: string; accent?: string }) {
  return (
    <div style={{ ...cardStyle, flex: 1, minWidth: 150 }}>
      <div style={{ fontSize: 12, fontWeight: 500, color: "var(--adaptive-600)" }}>{label}</div>
      <div
        style={{
          fontSize: 26,
          fontWeight: 700,
          color: accent ?? "var(--adaptive-900)",
          letterSpacing: "-0.02em",
          fontFeatureSettings: "'tnum'",
          marginTop: 6,
        }}
      >
        {value}
      </div>
      <div style={{ fontSize: 11.5, color: "var(--adaptive-500)", marginTop: 2 }}>{sub}</div>
    </div>
  );
}

function FieldRow({ icon, label, value }: { icon: IconName; label: string; value: string }) {
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 11, padding: "10px 0", borderBottom: "1px solid var(--adaptive-100)" }}>
      <Icon name={icon} size={17} color="var(--adaptive-400)" />
      <span style={{ fontSize: 13, color: "var(--adaptive-500)", width: 116, flexShrink: 0 }}>{label}</span>
      <span style={{ fontSize: 13.5, fontWeight: 500, color: "var(--adaptive-900)", textAlign: "right", marginLeft: "auto" }}>{value}</span>
    </div>
  );
}

type Activity = { at: number; label: string; sub: string; dot: string };

export function EmployeeDetail({
  employee,
  shifts,
  live,
  leaveRequests,
  locationNames,
  departmentName,
  roleName,
  onNavigate,
}: {
  employee: Employee;
  shifts: Shift[];
  live: LiveViewEntry[];
  leaveRequests: LeaveRequest[];
  locationNames: Map<string, string>;
  departmentName: string | null;
  roleName: string | null;
  onNavigate: (view: "roster" | "departments" | "roles") => void;
}) {
  const [attendance, setAttendance] = useState<AttendanceRecord[]>([]);
  const [loadingAtt, setLoadingAtt] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoadingAtt(true);
    attendanceApi
      .list({ employee_id: employee.id })
      .then((recs) => {
        if (!cancelled) setAttendance(recs);
      })
      .catch(() => {
        if (!cancelled) setAttendance([]);
      })
      .finally(() => {
        if (!cancelled) setLoadingAtt(false);
      });
    return () => {
      cancelled = true;
    };
  }, [employee.id]);

  const nowMs = Date.now();
  const days = weekDays();
  const wkStart = weekStart().getTime();
  const monthStart = new Date(new Date().getFullYear(), new Date().getMonth(), 1).getTime();

  const myShifts = shifts.filter((s) => s.employee_id === employee.id);
  const shiftById = new Map(shifts.map((s) => [s.id, s]));
  const liveEntry = live.find((e) => e.employee_id === employee.id) ?? null;
  const liveStatus = liveEntry ? deriveLiveStatus(liveEntry, nowMs) : null;
  const person = { id: employee.id, name: employee.full_name };
  const locName = (id?: string | null) => (id ? locationNames.get(id) ?? "Assigned" : "Assigned");

  // ── Real stats from attendance + shifts ──────────────────────────────────
  let onTimeConsidered = 0;
  let onTime = 0;
  let weekMinutes = 0;
  let completedThisMonth = 0;
  for (const r of attendance) {
    if (r.check_in_at && r.shift_id) {
      const sh = shiftById.get(r.shift_id);
      if (sh) {
        onTimeConsidered++;
        if (new Date(r.check_in_at).getTime() <= new Date(sh.starts_at).getTime()) onTime++;
      }
    }
    if (r.check_in_at && r.check_out_at) {
      const ci = new Date(r.check_in_at).getTime();
      if (ci >= wkStart) {
        weekMinutes += Math.max(0, (new Date(r.check_out_at).getTime() - ci) / 60000);
      }
      if (new Date(r.check_out_at).getTime() >= monthStart) completedThisMonth++;
    }
  }
  const onTimePct = onTimeConsidered ? Math.round((onTime / onTimeConsidered) * 100) : null;
  const hoursThisWeek = Math.round((weekMinutes / 60) * 10) / 10;
  const shiftsThisMonth = myShifts.filter((s) => new Date(s.starts_at).getTime() >= monthStart).length;

  // ── Recent activity from real attendance + leave + upcoming shift ─────────
  const activity: Activity[] = [];
  for (const r of attendance) {
    if (r.check_out_at) {
      const sh = r.shift_id ? shiftById.get(r.shift_id) : undefined;
      activity.push({
        at: new Date(r.check_out_at).getTime(),
        label: `Checked out — ${locName(sh?.location_id)}`,
        sub: relTime(r.check_out_at),
        dot: "var(--blue-500)",
      });
    }
    if (r.check_in_at) {
      const sh = r.shift_id ? shiftById.get(r.shift_id) : undefined;
      const late = sh ? new Date(r.check_in_at).getTime() > new Date(sh.starts_at).getTime() : false;
      activity.push({
        at: new Date(r.check_in_at).getTime(),
        label: `Checked in — ${late ? "late" : "on time"}`,
        sub: relTime(r.check_in_at),
        dot: late ? "var(--red-500)" : "var(--green-500)",
      });
    }
  }
  for (const lv of leaveRequests.filter((l) => l.employee_id === employee.id)) {
    const when = lv.reviewed_at ?? lv.created_at;
    const days_ = leaveDays(lv.start_date, lv.end_date);
    activity.push({
      at: new Date(when).getTime(),
      label: `Time-off ${lv.status} · ${days_} day${days_ === 1 ? "" : "s"}`,
      sub: relTime(when),
      dot:
        lv.status === "approved"
          ? "var(--green-500)"
          : lv.status === "rejected"
            ? "var(--red-500)"
            : "var(--adaptive-400)",
    });
  }
  const upcoming = myShifts
    .filter((s) => new Date(s.starts_at).getTime() > nowMs)
    .sort((a, b) => new Date(a.starts_at).getTime() - new Date(b.starts_at).getTime())[0];
  if (upcoming) {
    activity.push({
      at: new Date(upcoming.starts_at).getTime(),
      label: "Upcoming shift scheduled",
      sub: relTime(upcoming.starts_at),
      dot: "var(--adaptive-300)",
    });
  }
  activity.sort((a, b) => b.at - a.at);
  const recentActivity = activity.slice(0, 6);

  const tenureLabel = (() => {
    if (!employee.hired_at) return null;
    const hired = new Date(employee.hired_at);
    const since = new Intl.DateTimeFormat(undefined, { month: "short", year: "numeric" }).format(hired);
    const yrs = Math.floor((nowMs - hired.getTime()) / (365.25 * 86_400_000));
    return yrs >= 1 ? `${yrs} yr${yrs === 1 ? "" : "s"} · since ${since}` : `since ${since}`;
  })();

  return (
    <div style={{ padding: "20px 24px 40px", display: "flex", flexDirection: "column", gap: 18 }}>
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
            </div>
            <div style={{ display: "flex", gap: 7, marginTop: 12, flexWrap: "wrap" }}>
              <Chip tone={TYPE_TONE[employee.employment_type]}>{humanize(employee.employment_type)}</Chip>
              <Chip tone={employee.status === "active" ? "blue" : "neutral"}>{employee.status}</Chip>
              {tenureLabel && (
                <Chip>
                  <Icon name="clock" size={12} />
                  {tenureLabel}
                </Chip>
              )}
              {employee.title && <Chip>{employee.title}</Chip>}
            </div>
          </div>
          <div style={{ display: "flex", gap: 8, flexShrink: 0, flexWrap: "wrap" }}>
            <Btn
              variant="secondary"
              icon="send"
              disabled={!employee.email}
              onClick={() => employee.email && window.open(`mailto:${employee.email}`)}
            >
              Message
            </Btn>
            <Btn
              variant="secondary"
              icon="phone"
              disabled={!employee.phone}
              onClick={() => employee.phone && window.open(`tel:${employee.phone.replace(/\s+/g, "")}`)}
            >
              Call
            </Btn>
            <Btn variant="primary" icon="calendar" onClick={() => onNavigate("roster")}>
              Assign Shift
            </Btn>
          </div>
        </div>
      </div>

      {/* stat cards */}
      <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
        <StatCard
          label="On-time rate"
          value={onTimePct == null ? "—" : `${onTimePct}%`}
          sub={loadingAtt ? "loading…" : onTimePct == null ? "no check-ins yet" : "from check-ins"}
          accent={onTimePct != null && onTimePct >= 95 ? "var(--green-600)" : undefined}
        />
        <StatCard label="Hours this week" value={`${hoursThisWeek}h`} sub={loadingAtt ? "loading…" : "from completed shifts"} />
        <StatCard label="Completed" value={`${completedThisMonth}`} sub="this month" />
        <StatCard label="Shifts" value={`${shiftsThisMonth}`} sub="scheduled this month" />
      </div>

      {/* two-column grid */}
      <div style={{ display: "grid", gridTemplateColumns: "minmax(0, 1.55fr) minmax(0, 1fr)", gap: 16, alignItems: "start" }}>
        {/* left: schedule + activity */}
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
                              {s.status === "draft" && <div style={{ fontSize: 8.5, fontWeight: 700, color, marginTop: 2 }}>DRAFT</div>}
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

          <div style={cardStyle}>
            <SectionTitle>Recent Activity</SectionTitle>
            {loadingAtt ? (
              <div style={{ fontSize: 13, color: "var(--adaptive-400)" }}>Loading…</div>
            ) : recentActivity.length === 0 ? (
              <div style={{ fontSize: 13, color: "var(--adaptive-400)" }}>No recent activity.</div>
            ) : (
              <div>
                {recentActivity.map((a, i) => (
                  <div key={i} style={{ display: "flex", gap: 12, alignItems: "flex-start" }}>
                    <div style={{ display: "flex", flexDirection: "column", alignItems: "center" }}>
                      <div style={{ width: 10, height: 10, borderRadius: "50%", background: a.dot, boxShadow: "0 0 0 3px var(--card)", marginTop: 4 }} />
                      {i < recentActivity.length - 1 && <div style={{ width: 2, flex: 1, minHeight: 24, background: "var(--adaptive-200)" }} />}
                    </div>
                    <div style={{ flex: 1, paddingBottom: 16 }}>
                      <div style={{ fontSize: 13.5, color: "var(--adaptive-800)" }}>{a.label}</div>
                      <div style={{ fontSize: 11.5, color: "var(--adaptive-400)", marginTop: 1 }}>{a.sub}</div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* right: contact + employment + role & access */}
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
            </div>
          </div>

          <div style={cardStyle}>
            <SectionTitle>Role &amp; Access</SectionTitle>
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              <NavRow
                icon="briefcase"
                label="Department"
                value={departmentName ?? "Unassigned"}
                onClick={() => onNavigate("departments")}
              />
              <NavRow icon="route" label="Role" value={roleName ?? "None"} onClick={() => onNavigate("roles")} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function NavRow({ icon, label, value, onClick }: { icon: IconName; label: string; value: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      style={{
        display: "flex",
        alignItems: "center",
        gap: 10,
        padding: "10px 12px",
        borderRadius: 8,
        border: "1px solid var(--adaptive-200)",
        background: "var(--card)",
        cursor: "pointer",
        fontFamily: "inherit",
        textAlign: "left",
        width: "100%",
      }}
    >
      <div
        style={{
          width: 30,
          height: 30,
          borderRadius: 7,
          background: "var(--adaptive-100)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Icon name={icon} size={16} color="var(--adaptive-600)" />
      </div>
      <div style={{ flex: 1 }}>
        <div style={{ fontSize: 11, color: "var(--adaptive-500)" }}>{label}</div>
        <div style={{ fontSize: 13.5, fontWeight: 600, color: "var(--adaptive-900)" }}>{value}</div>
      </div>
      <Icon name="chevron" size={15} color="var(--adaptive-300)" />
    </button>
  );
}

function leaveDays(start: string, end: string): number {
  const a = new Date(start + "T00:00:00Z").getTime();
  const b = new Date(end + "T00:00:00Z").getTime();
  if (Number.isNaN(a) || Number.isNaN(b) || b < a) return 1;
  return Math.round((b - a) / 86_400_000) + 1;
}

function relTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const now = new Date();
  const sameDate = (x: Date, y: Date) =>
    x.getFullYear() === y.getFullYear() && x.getMonth() === y.getMonth() && x.getDate() === y.getDate();
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  const time = new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit" }).format(d);
  if (sameDate(d, now)) return `Today · ${time}`;
  if (sameDate(d, yesterday)) return `Yesterday · ${time}`;
  return `${new Intl.DateTimeFormat(undefined, { day: "2-digit", month: "short" }).format(d)} · ${time}`;
}
