import type { CSSProperties, ReactNode } from "react";

// Platform accent = blue (distinct from the tenant orange primary), so the
// internal console never reads as a tenant screen.
export const ACCENT = "var(--blue-600)";

type PillTone = {
  fg: string;
  bg: string;
  bd: string;
  dot: string;
};

const TONES: Record<string, PillTone> = {
  green: { fg: "var(--green-700)", bg: "var(--green-50)", bd: "var(--green-200)", dot: "var(--green-500)" },
  amber: { fg: "var(--amber-700)", bg: "var(--amber-50)", bd: "var(--amber-200)", dot: "var(--amber-500)" },
  red: { fg: "var(--red-700)", bg: "var(--red-50)", bd: "var(--red-200)", dot: "var(--red-500)" },
  blue: { fg: "var(--blue-700)", bg: "var(--blue-50)", bd: "var(--blue-200)", dot: "var(--blue-500)" },
  gray: { fg: "var(--adaptive-600)", bg: "var(--adaptive-100)", bd: "var(--adaptive-200)", dot: "var(--adaptive-400)" },
};

export function StatusPill({ tone, label }: { tone: keyof typeof TONES; label: string }) {
  const t = TONES[tone];
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 6,
        padding: "3px 10px",
        borderRadius: 9999,
        background: t.bg,
        border: `1px solid ${t.bd}`,
        fontSize: 12,
        fontWeight: 600,
        color: t.fg,
        whiteSpace: "nowrap",
      }}
    >
      <span style={{ width: 7, height: 7, borderRadius: "50%", background: t.dot }} />
      {label}
    </span>
  );
}

export function tenantTone(status: string): keyof typeof TONES {
  if (status === "active") return "green";
  if (status === "suspended") return "red";
  if (status === "provisioning") return "amber";
  return "gray";
}

export function userTone(status: string): keyof typeof TONES {
  return status === "active" ? "green" : "gray";
}

// Read-only monospace-ish key/value used across detail views.
export function KV({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <div style={{ fontSize: 11, fontWeight: 600, letterSpacing: ".04em", textTransform: "uppercase", color: "var(--adaptive-500)" }}>
        {label}
      </div>
      <div style={{ fontSize: 13.5, color: "var(--adaptive-900)", wordBreak: "break-word" }}>{value}</div>
    </div>
  );
}

export const tdStyle: CSSProperties = {
  padding: "11px 16px",
  fontSize: 13.5,
  color: "var(--adaptive-800)",
  borderBottom: "1px solid var(--adaptive-100)",
  whiteSpace: "nowrap",
};

export function TableShell({ children }: { children: ReactNode }) {
  return (
    <div
      style={{
        border: "1px solid var(--adaptive-200)",
        borderRadius: 10,
        overflow: "hidden",
        background: "var(--card)",
      }}
    >
      <div style={{ overflowX: "auto" }}>
        <table style={{ width: "100%", borderCollapse: "collapse", minWidth: 640 }}>{children}</table>
      </div>
    </div>
  );
}

export function EmptyState({ text }: { text: string }) {
  return (
    <div
      style={{
        padding: "40px 20px",
        textAlign: "center",
        color: "var(--adaptive-500)",
        fontSize: 13.5,
      }}
    >
      {text}
    </div>
  );
}

// Small banner stating an action writes an audited platform event.
export function AuditNote({ children }: { children: ReactNode }) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 8,
        padding: "9px 12px",
        borderRadius: 8,
        background: "var(--blue-50)",
        border: "1px solid var(--blue-200)",
        color: "var(--blue-700)",
        fontSize: 12.5,
        fontWeight: 500,
      }}
    >
      {children}
    </div>
  );
}
