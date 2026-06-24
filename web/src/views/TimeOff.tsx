import { useMemo, useState } from "react";
import type { Employee, LeaveRequest, LeaveStatus } from "../api/resources";
import {
  Avatar,
  Btn,
  Card,
  Chip,
  Drawer,
  DrawerSectionLabel,
  Icon,
  PageHeader,
  humanize,
} from "../ui";
import type { ChipTone } from "../ui";

// The API uses pending | approved | rejected. (The prototype called the third
// "declined"; we surface the API's wording.)
type Tab = "pending" | "approved" | "rejected" | "all";

const TYPE_TONE: Record<string, ChipTone> = {
  holiday: "blue",
  sick: "orange",
  personal: "neutral",
};

const STATUS_STYLE: Record<LeaveStatus, { label: string; fg: string; bg: string; bd: string; dot: string }> = {
  pending: { label: "Pending", fg: "var(--amber-700)", bg: "var(--amber-50)", bd: "var(--amber-200)", dot: "var(--amber-500)" },
  approved: { label: "Approved", fg: "var(--green-700)", bg: "var(--green-50)", bd: "var(--green-200)", dot: "var(--green-500)" },
  rejected: { label: "Rejected", fg: "var(--red-700)", bg: "var(--red-50)", bd: "var(--red-200)", dot: "var(--red-500)" },
};

function daysInclusive(start: string, end: string): number {
  const a = new Date(start + "T00:00:00Z").getTime();
  const b = new Date(end + "T00:00:00Z").getTime();
  if (Number.isNaN(a) || Number.isNaN(b) || b < a) return 1;
  return Math.round((b - a) / 86_400_000) + 1;
}

function fmtDate(d: string): string {
  const date = new Date(d + "T00:00:00Z");
  return Number.isNaN(date.getTime())
    ? d
    : new Intl.DateTimeFormat(undefined, { day: "2-digit", month: "short", timeZone: "UTC" }).format(date);
}

function StatusBadge({ status }: { status: LeaveStatus }) {
  const s = STATUS_STYLE[status];
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 6,
        padding: "2px 9px",
        borderRadius: 9999,
        background: s.bg,
        border: `1px solid ${s.bd}`,
        fontSize: 11.5,
        fontWeight: 600,
        color: s.fg,
      }}
    >
      <span style={{ width: 6, height: 6, borderRadius: "50%", background: s.dot }} />
      {s.label}
    </span>
  );
}

function StatCard({ label, value, dot }: { label: string; value: number; dot?: string }) {
  return (
    <Card style={{ flex: 1, minWidth: 130, padding: "12px 14px" }}>
      <div style={{ display: "flex", alignItems: "center", gap: 7, marginBottom: 8 }}>
        {dot && <span style={{ width: 9, height: 9, borderRadius: "50%", background: dot }} />}
        <span style={{ fontSize: 12, fontWeight: 500, color: "var(--adaptive-600)" }}>{label}</span>
      </div>
      <div style={{ fontSize: 26, fontWeight: 700, color: "var(--adaptive-900)", letterSpacing: "-0.02em", fontFeatureSettings: "'tnum'" }}>
        {value}
      </div>
    </Card>
  );
}

function ReviewDrawer({
  req,
  employeeName,
  onClose,
  onApprove,
  onReject,
}: {
  req: LeaveRequest;
  employeeName: string;
  onClose: () => void;
  onApprove: () => void;
  onReject: () => void;
}) {
  const days = daysInclusive(req.start_date, req.end_date);
  return (
    <Drawer
      onClose={onClose}
      width={410}
      header={
        <>
          <Avatar person={{ id: req.employee_id, name: employeeName }} size={42} />
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 16, fontWeight: 600, color: "var(--adaptive-900)" }}>{employeeName}</div>
            <div style={{ fontSize: 12.5, color: "var(--adaptive-500)" }}>{humanize(req.type)} leave</div>
          </div>
        </>
      }
      footer={
        req.status === "pending" ? (
          <>
            <Btn variant="secondary" icon="x" style={{ flex: 1 }} onClick={onReject}>
              Reject
            </Btn>
            <Btn variant="primary" icon="check" style={{ flex: 1 }} onClick={onApprove}>
              Approve
            </Btn>
          </>
        ) : (
          <div style={{ flex: 1, textAlign: "center", fontSize: 13, color: "var(--adaptive-500)" }}>
            This request was {STATUS_STYLE[req.status].label.toLowerCase()}.
          </div>
        )
      }
    >
      <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
        <StatusBadge status={req.status} />
        <Chip tone={TYPE_TONE[req.type] ?? "neutral"}>{humanize(req.type)}</Chip>
      </div>

      <Card style={{ padding: 14 }}>
        <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "9px 14px", fontSize: 13 }}>
          <span style={{ color: "var(--adaptive-500)" }}>Dates</span>
          <span style={{ color: "var(--adaptive-900)", fontWeight: 600, textAlign: "right" }}>
            {fmtDate(req.start_date)} – {fmtDate(req.end_date)}
          </span>
          <span style={{ color: "var(--adaptive-500)" }}>Duration</span>
          <span style={{ color: "var(--adaptive-900)", fontWeight: 600, textAlign: "right" }}>
            {days} day{days === 1 ? "" : "s"}
          </span>
          {req.reviewed_at && (
            <>
              <span style={{ color: "var(--adaptive-500)" }}>Reviewed</span>
              <span style={{ color: "var(--adaptive-800)", textAlign: "right" }}>
                {new Intl.DateTimeFormat(undefined, { dateStyle: "medium" }).format(new Date(req.reviewed_at))}
              </span>
            </>
          )}
        </div>
      </Card>

      {req.note && (
        <div>
          <DrawerSectionLabel>Note</DrawerSectionLabel>
          <div
            style={{
              fontSize: 13.5,
              color: "var(--adaptive-700)",
              lineHeight: 1.55,
              padding: "11px 13px",
              background: "var(--adaptive-50)",
              border: "1px solid var(--adaptive-200)",
              borderRadius: 8,
            }}
          >
            {req.note}
          </div>
        </div>
      )}
    </Drawer>
  );
}

export function TimeOff({
  requests,
  employees,
  onApprove,
  onReject,
}: {
  requests: LeaveRequest[];
  employees: Employee[];
  onApprove: (id: string) => void;
  onReject: (id: string) => void;
}) {
  const [tab, setTab] = useState<Tab>("pending");
  const [sel, setSel] = useState<LeaveRequest | null>(null);

  const nameById = useMemo(
    () => new Map(employees.map((e) => [e.id, e.full_name])),
    [employees],
  );
  const nameFor = (id: string) => nameById.get(id) ?? "Unknown employee";

  const counts = {
    pending: requests.filter((r) => r.status === "pending").length,
    approved: requests.filter((r) => r.status === "approved").length,
    rejected: requests.filter((r) => r.status === "rejected").length,
  };
  const daysBooked = requests
    .filter((r) => r.status === "approved")
    .reduce((n, r) => n + daysInclusive(r.start_date, r.end_date), 0);

  const shown = tab === "all" ? requests : requests.filter((r) => r.status === tab);
  const tabs: [Tab, string, number][] = [
    ["pending", "Pending", counts.pending],
    ["approved", "Approved", counts.approved],
    ["rejected", "Rejected", counts.rejected],
    ["all", "All", requests.length],
  ];

  // Keep the open drawer in sync with refreshed data (status changes after approve/reject).
  const selLive = sel ? requests.find((r) => r.id === sel.id) ?? null : null;

  return (
    <div style={{ padding: "20px 24px 32px", display: "flex", flexDirection: "column", gap: 18 }}>
      <PageHeader
        title="Time Off"
        subtitle={`${counts.pending} awaiting your review`}
      />

      <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
        <StatCard label="Pending" value={counts.pending} dot="var(--amber-500)" />
        <StatCard label="Approved" value={counts.approved} dot="var(--green-500)" />
        <StatCard label="Rejected" value={counts.rejected} dot="var(--red-500)" />
        <StatCard label="Days off booked" value={daysBooked} />
      </div>

      <div style={{ display: "flex", gap: 4, borderBottom: "1px solid var(--adaptive-200)", flexWrap: "wrap" }}>
        {tabs.map(([id, label, n]) => {
          const on = id === tab;
          return (
            <button
              key={id}
              onClick={() => setTab(id)}
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
              {label}
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
                {n}
              </span>
            </button>
          );
        })}
      </div>

      <Card style={{ overflow: "hidden" }}>
        {shown.length === 0 ? (
          <div style={{ padding: "48px 0", textAlign: "center", color: "var(--adaptive-400)" }}>
            <Icon name="check" size={26} color="var(--adaptive-300)" style={{ margin: "0 auto 10px", display: "block" }} />
            <div style={{ fontSize: 14, fontWeight: 600, color: "var(--adaptive-600)" }}>Nothing here</div>
            <div style={{ fontSize: 13, marginTop: 3 }}>No {tab === "all" ? "" : tab} requests.</div>
          </div>
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13, minWidth: 720 }}>
              <thead>
                <tr style={{ background: "var(--adaptive-50)", textAlign: "left" }}>
                  {["Employee", "Type", "Dates", "Days", "Status", ""].map((h, i) => (
                    <th
                      key={i}
                      style={{
                        padding: "11px 16px",
                        fontWeight: 600,
                        fontSize: 12,
                        color: "var(--adaptive-500)",
                        borderBottom: "1px solid var(--adaptive-200)",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {shown.map((r, i) => {
                  const name = nameFor(r.employee_id);
                  const days = daysInclusive(r.start_date, r.end_date);
                  return (
                    <tr
                      key={r.id}
                      onClick={() => setSel(r)}
                      style={{ cursor: "pointer", borderBottom: i < shown.length - 1 ? "1px solid var(--adaptive-100)" : "none" }}
                      onMouseEnter={(e) => (e.currentTarget.style.background = "var(--adaptive-50)")}
                      onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                    >
                      <td style={{ padding: "11px 16px" }}>
                        <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
                          <Avatar person={{ id: r.employee_id, name }} size={30} />
                          <span style={{ fontWeight: 600, color: "var(--adaptive-900)" }}>{name}</span>
                        </div>
                      </td>
                      <td style={{ padding: "11px 16px" }}>
                        <Chip tone={TYPE_TONE[r.type] ?? "neutral"}>{humanize(r.type)}</Chip>
                      </td>
                      <td style={{ padding: "11px 16px", color: "var(--adaptive-700)" }}>
                        {fmtDate(r.start_date)} – {fmtDate(r.end_date)}
                      </td>
                      <td style={{ padding: "11px 16px", color: "var(--adaptive-700)", fontFeatureSettings: "'tnum'" }}>{days}</td>
                      <td style={{ padding: "11px 16px" }}>
                        <StatusBadge status={r.status} />
                      </td>
                      <td style={{ padding: "11px 16px", textAlign: "right" }}>
                        {r.status === "pending" ? (
                          <div style={{ display: "inline-flex", gap: 6 }} onClick={(e) => e.stopPropagation()}>
                            <button
                              onClick={() => onReject(r.id)}
                              title="Reject"
                              style={{ width: 30, height: 30, borderRadius: 6, border: "1px solid var(--adaptive-200)", background: "var(--card)", cursor: "pointer", display: "inline-flex", alignItems: "center", justifyContent: "center" }}
                            >
                              <Icon name="x" size={14} color="var(--red-500)" />
                            </button>
                            <button
                              onClick={() => onApprove(r.id)}
                              title="Approve"
                              style={{ width: 30, height: 30, borderRadius: 6, border: "1px solid var(--green-200)", background: "var(--green-50)", cursor: "pointer", display: "inline-flex", alignItems: "center", justifyContent: "center" }}
                            >
                              <Icon name="check" size={14} color="var(--green-600)" />
                            </button>
                          </div>
                        ) : (
                          <Icon name="chevron" size={15} color="var(--adaptive-300)" />
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {selLive && (
        <ReviewDrawer
          req={selLive}
          employeeName={nameFor(selLive.employee_id)}
          onClose={() => setSel(null)}
          onApprove={() => onApprove(selLive.id)}
          onReject={() => onReject(selLive.id)}
        />
      )}
    </div>
  );
}
