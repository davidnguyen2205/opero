import { useState } from "react";
import { platformAuthApi } from "../api/platform";
import type { PlatformAuthResponse } from "../api/platform";
import { Btn, Field, Icon, controlStyle } from "../ui";

// Standalone platform sign-in. Deliberately NOT the tenant AuthScreen: distinct
// dark "Internal" chrome so it is unmistakably the Opero staff console.
export function PlatformLogin({
  onAuthenticated,
}: {
  onAuthenticated: (auth: PlatformAuthResponse) => void;
}) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit() {
    setSubmitting(true);
    setError(null);
    try {
      const result = await platformAuthApi.login({ email, password });
      onAuthenticated(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Sign-in failed.");
    }
    setSubmitting(false);
  }

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 24,
        background:
          "radial-gradient(1200px 600px at 50% -10%, #1e293b, var(--adaptive-950))",
        fontFamily: "var(--font-sans)",
      }}
    >
      <div
        style={{
          width: "100%",
          maxWidth: 400,
          background: "var(--card)",
          border: "1px solid var(--adaptive-200)",
          borderRadius: 12,
          boxShadow: "var(--shadow-lg)",
          overflow: "hidden",
        }}
      >
        <div
          style={{
            padding: "20px 24px",
            background: "var(--adaptive-950)",
            color: "#fff",
            display: "flex",
            alignItems: "center",
            gap: 12,
          }}
        >
          <div
            style={{
              width: 34,
              height: 34,
              borderRadius: 8,
              background: "var(--blue-600)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              flexShrink: 0,
            }}
          >
            <Icon name="activity" size={19} color="#fff" />
          </div>
          <div>
            <div style={{ fontSize: 16, fontWeight: 700, letterSpacing: "-0.01em" }}>
              Opero Platform
            </div>
            <div style={{ fontSize: 12, color: "rgba(255,255,255,0.6)" }}>
              Internal Super Admin console
            </div>
          </div>
          <span
            style={{
              marginLeft: "auto",
              fontSize: 9,
              fontWeight: 700,
              letterSpacing: ".08em",
              padding: "3px 7px",
              borderRadius: 5,
              background: "rgba(59,130,246,0.2)",
              color: "#93c5fd",
              border: "1px solid rgba(59,130,246,0.4)",
            }}
          >
            INTERNAL
          </span>
        </div>

        <div style={{ padding: 24, display: "flex", flexDirection: "column", gap: 14 }}>
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
          <Field label="Staff email">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              style={controlStyle}
              autoComplete="email"
              onKeyDown={(e) => e.key === "Enter" && void submit()}
            />
          </Field>
          <Field label="Password">
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              style={controlStyle}
              autoComplete="current-password"
              onKeyDown={(e) => e.key === "Enter" && void submit()}
            />
          </Field>
          <Btn variant="primary" size="lg" disabled={submitting} onClick={() => void submit()}>
            {submitting ? "Signing in…" : "Sign in"}
          </Btn>
          <div style={{ fontSize: 11.5, color: "var(--adaptive-500)", textAlign: "center" }}>
            Opero staff only. Access is audited.
          </div>
        </div>
      </div>
    </div>
  );
}
