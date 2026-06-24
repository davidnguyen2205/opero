import { useState } from "react";
import type { CSSProperties } from "react";
import { authApi } from "../api/resources";
import type { AuthResponse } from "../api/resources";
import { Btn, Field, Icon, controlStyle } from "../ui";

type Mode = "login" | "signup";

function OperoMark({ size = 34 }: { size?: number }) {
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: 9,
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

export function AuthScreen({ onAuthenticated }: { onAuthenticated: (auth: AuthResponse) => void }) {
  const [mode, setMode] = useState<Mode>("login");
  const [tenantSlug, setTenantSlug] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [companyName, setCompanyName] = useState("");
  const [slug, setSlug] = useState("");
  const [adminName, setAdminName] = useState("");
  const [adminEmail, setAdminEmail] = useState("");
  const [adminPassword, setAdminPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit() {
    setSubmitting(true);
    setError(null);
    try {
      const result =
        mode === "login"
          ? await authApi.login({ email, password, tenant_slug: tenantSlug })
          : await authApi.signup({
              company_name: companyName,
              slug: slug.trim() || undefined,
              admin_full_name: adminName.trim() || undefined,
              admin_email: adminEmail,
              admin_password: adminPassword,
            });
      onAuthenticated(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed.");
    } finally {
      setSubmitting(false);
    }
  }

  const tab = (m: Mode): CSSProperties => ({
    flex: 1,
    minHeight: 36,
    padding: "0 14px",
    borderRadius: 6,
    border: 0,
    cursor: "pointer",
    fontFamily: "inherit",
    fontWeight: 600,
    fontSize: 13,
    color: mode === m ? "var(--primary-700)" : "var(--adaptive-600)",
    background: mode === m ? "var(--card)" : "transparent",
    boxShadow: mode === m ? "var(--shadow-xs)" : "none",
  });

  return (
    <div style={{ display: "grid", minHeight: "100vh", gridTemplateColumns: "minmax(320px, 460px) 1fr", background: "var(--background)" }}>
      <section
        style={{
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          gap: 22,
          padding: "48px clamp(28px, 5vw, 56px)",
          borderRight: "1px solid var(--adaptive-200)",
          background: "var(--card)",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <OperoMark />
          <div>
            <div style={{ fontSize: 22, fontWeight: 700, letterSpacing: "-0.02em", color: "var(--adaptive-900)" }}>
              Opero
            </div>
            <div style={{ fontSize: 13, color: "var(--adaptive-500)" }}>
              Manager console for people, roster & field ops
            </div>
          </div>
        </div>

        <div
          style={{
            display: "flex",
            gap: 4,
            padding: 4,
            border: "1px solid var(--adaptive-200)",
            borderRadius: 8,
            background: "var(--adaptive-100)",
          }}
        >
          <button onClick={() => setMode("login")} style={tab("login")}>
            Log in
          </button>
          <button onClick={() => setMode("signup")} style={tab("signup")}>
            Sign up
          </button>
        </div>

        {error && (
          <div
            style={{
              border: "1px solid var(--red-200)",
              color: "var(--red-700)",
              background: "var(--red-50)",
              borderRadius: 8,
              padding: "11px 14px",
              fontSize: 13.5,
            }}
          >
            {error}
          </div>
        )}

        <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          {mode === "login" ? (
            <>
              <Field label="Tenant slug">
                <input value={tenantSlug} onChange={(e) => setTenantSlug(e.target.value)} style={controlStyle} autoComplete="organization" />
              </Field>
              <Field label="Email">
                <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} style={controlStyle} autoComplete="email" />
              </Field>
              <Field label="Password">
                <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} style={controlStyle} autoComplete="current-password" />
              </Field>
            </>
          ) : (
            <>
              <Field label="Company name">
                <input value={companyName} onChange={(e) => setCompanyName(e.target.value)} style={controlStyle} />
              </Field>
              <Field label="Tenant slug (optional)">
                <input value={slug} onChange={(e) => setSlug(e.target.value)} style={controlStyle} pattern="^[a-z0-9]+(?:-[a-z0-9]+)*$" />
              </Field>
              <Field label="Admin full name">
                <input value={adminName} onChange={(e) => setAdminName(e.target.value)} style={controlStyle} />
              </Field>
              <Field label="Admin email">
                <input type="email" value={adminEmail} onChange={(e) => setAdminEmail(e.target.value)} style={controlStyle} autoComplete="email" />
              </Field>
              <Field label="Admin password">
                <input type="password" value={adminPassword} onChange={(e) => setAdminPassword(e.target.value)} style={controlStyle} autoComplete="new-password" minLength={8} />
              </Field>
            </>
          )}
          <Btn variant="primary" size="lg" disabled={submitting} onClick={() => void submit()}>
            {submitting ? "Working…" : mode === "login" ? "Log in" : "Create tenant"}
          </Btn>
        </div>
      </section>

      <section
        style={{
          display: "flex",
          flexDirection: "column",
          justifyContent: "flex-end",
          gap: 18,
          padding: "clamp(32px, 5vw, 64px)",
          color: "#fff",
          background:
            "linear-gradient(160deg, var(--primary-700), var(--adaptive-950))",
        }}
      >
        <div
          style={{
            display: "inline-flex",
            alignItems: "center",
            gap: 8,
            alignSelf: "flex-start",
            padding: "6px 12px",
            borderRadius: 9999,
            background: "rgba(255,255,255,0.12)",
            fontSize: 12.5,
            fontWeight: 600,
          }}
        >
          <Icon name="activity" size={15} /> Real-time field operations
        </div>
        <h1 style={{ margin: 0, fontSize: "clamp(1.8rem, 4vw, 3rem)", lineHeight: 1.05, fontWeight: 700 }}>
          Run the daily field-ops loop from one place.
        </h1>
        <p style={{ margin: 0, maxWidth: 560, fontSize: "1.05rem", color: "rgba(255,255,255,0.82)" }}>
          Build the roster, maintain the people core, and publish assignments for guides, drivers,
          operators, and office staff — then watch who's working in real time.
        </p>
      </section>
    </div>
  );
}
