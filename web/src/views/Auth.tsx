import { useState } from "react";
import type { ReactNode } from "react";
import { authApi } from "../api/resources";
import type { AuthResponse } from "../api/resources";
import { useTypewriter } from "../hooks/useTypewriter";
import { Btn, Icon, OperoMark, controlStyle } from "../ui";
import type { IconName } from "../ui";

type Mode = "login" | "signup";

// Rotating sign-in headlines (pattern ported from the Blazeup Super Admin
// login: type each phrase out, hold it, then start the next at random).
const SIGN_IN_TITLES = [
  "Welcome to Opero!",
  "Have a nice day!",
  "Let's get to work",
  "Ready when you are",
  "Make today count",
  "Your crew is waiting",
];

const REDUCED_MOTION =
  typeof window !== "undefined" &&
  window.matchMedia("(prefers-reduced-motion: reduce)").matches;

// Each character is its own span that mounts once and fades in, so the fade
// always plays to completion instead of snapping when the next char arrives.
// The leading character is tinted primary while typing and eases back to the
// title color once the phrase settles. The caret is solid while typing and
// blinks when the phrase is held.
function TypewriterTitle() {
  const { text, phase } = useTypewriter(SIGN_IN_TITLES, {
    typingSpeed: 50,
    pauseDuration: 3000,
  });
  if (REDUCED_MOTION) {
    return <>{SIGN_IN_TITLES[0]}</>;
  }
  const isTyping = phase === "typing";
  const lastIndex = text.length - 1;
  return (
    <>
      {text.split("").map((char, i) => (
        <span
          key={i}
          className="opero-char-in"
          style={isTyping && i === lastIndex ? { color: "var(--primary-600)" } : undefined}
        >
          {char}
        </span>
      ))}
      <span
        aria-hidden="true"
        className={isTyping ? "" : "opero-caret-blink"}
        style={{
          marginLeft: 4,
          display: "inline-block",
          width: 2,
          height: "0.8em",
          transform: "translateY(0.12em)",
          background: "var(--primary-300)",
        }}
      />
    </>
  );
}

// Label with the reference's required-asterisk treatment.
function FieldLabel({ children, required }: { children: ReactNode; required?: boolean }) {
  return (
    <label
      style={{
        display: "block",
        fontSize: 13,
        fontWeight: 600,
        color: "var(--adaptive-800)",
        marginBottom: 6,
      }}
    >
      {children}
      {required && <span style={{ color: "var(--red-500)", marginLeft: 3 }}>*</span>}
    </label>
  );
}

// Input with a leading icon, like the reference's email field.
function IconInput({
  icon,
  ...rest
}: { icon: IconName } & React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <div style={{ position: "relative" }}>
      <span
        style={{
          position: "absolute",
          left: 13,
          top: "50%",
          transform: "translateY(-50%)",
          display: "flex",
          pointerEvents: "none",
        }}
      >
        <Icon name={icon} size={16} color="var(--adaptive-400)" />
      </span>
      <input
        {...rest}
        style={{ ...controlStyle, width: "100%", boxSizing: "border-box", paddingLeft: 38, minHeight: 42 }}
      />
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

  function switchMode(m: Mode) {
    setMode(m);
    setError(null);
  }

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        // Warm cream page ground, same value as the Blazeup login layout.
        background: "#fefbf6",
        padding: "24px 20px",
      }}
    >
      <main
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          width: "100%",
          maxWidth: 432,
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 26 }}>
          <OperoMark size={40} />
          <span style={{ fontSize: 24, fontWeight: 700, letterSpacing: "-0.02em", color: "var(--adaptive-900)" }}>
            opero
          </span>
          <span
            style={{
              border: "1px solid var(--primary-300)",
              color: "var(--primary-600)",
              background: "var(--primary-50)",
              borderRadius: 9999,
              padding: "3px 11px",
              fontSize: 11,
              fontWeight: 700,
              letterSpacing: "0.06em",
            }}
          >
            MANAGER CONSOLE
          </span>
        </div>

        <h1
          style={{
            margin: "0 0 8px",
            fontSize: 30,
            fontWeight: 800,
            letterSpacing: "-0.02em",
            color: "var(--adaptive-900)",
            minHeight: "1.15em",
            whiteSpace: "nowrap",
          }}
        >
          {mode === "login" ? <TypewriterTitle /> : "Create your workspace"}
        </h1>
        <p style={{ margin: "0 0 26px", fontSize: 15, color: "var(--adaptive-600)" }}>
          {mode === "login"
            ? "Sign in to run your roster and field operations."
            : "Set up a company workspace and its first admin account."}
        </p>

        {error && (
          <div
            style={{
              border: "1px solid var(--red-200)",
              color: "var(--red-700)",
              background: "var(--red-50)",
              borderRadius: 8,
              padding: "11px 14px",
              fontSize: 13.5,
              marginBottom: 18,
            }}
          >
            {error}
          </div>
        )}

        <form
          onSubmit={(e) => {
            e.preventDefault();
            void submit();
          }}
          style={{ display: "flex", flexDirection: "column", gap: 16 }}
        >
          {mode === "login" ? (
            <>
              <div>
                <FieldLabel required>Workspace</FieldLabel>
                <IconInput
                  icon="briefcase"
                  value={tenantSlug}
                  onChange={(e) => setTenantSlug(e.target.value)}
                  placeholder="your-company-slug"
                  autoComplete="organization"
                  required
                />
              </div>
              <div>
                <FieldLabel required>Email address</FieldLabel>
                <IconInput
                  icon="mail"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="Enter your email"
                  autoComplete="email"
                  required
                />
              </div>
              <div>
                <FieldLabel required>Password</FieldLabel>
                <IconInput
                  icon="lock"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Enter your password"
                  autoComplete="current-password"
                  required
                />
              </div>
            </>
          ) : (
            <>
              <div>
                <FieldLabel required>Company name</FieldLabel>
                <IconInput
                  icon="briefcase"
                  value={companyName}
                  onChange={(e) => setCompanyName(e.target.value)}
                  placeholder="Saigon Tours Co."
                  required
                />
              </div>
              <div>
                <FieldLabel>Workspace slug</FieldLabel>
                <IconInput
                  icon="route"
                  value={slug}
                  onChange={(e) => setSlug(e.target.value)}
                  placeholder="saigon-tours (optional)"
                  pattern="^[a-z0-9]+(?:-[a-z0-9]+)*$"
                />
              </div>
              <div>
                <FieldLabel>Admin full name</FieldLabel>
                <IconInput
                  icon="users"
                  value={adminName}
                  onChange={(e) => setAdminName(e.target.value)}
                  placeholder="Your name (optional)"
                />
              </div>
              <div>
                <FieldLabel required>Admin email</FieldLabel>
                <IconInput
                  icon="mail"
                  type="email"
                  value={adminEmail}
                  onChange={(e) => setAdminEmail(e.target.value)}
                  placeholder="you@company.com"
                  autoComplete="email"
                  required
                />
              </div>
              <div>
                <FieldLabel required>Admin password</FieldLabel>
                <IconInput
                  icon="lock"
                  type="password"
                  value={adminPassword}
                  onChange={(e) => setAdminPassword(e.target.value)}
                  placeholder="At least 8 characters"
                  autoComplete="new-password"
                  minLength={8}
                  required
                />
              </div>
            </>
          )}

          <Btn variant="primary" size="lg" type="submit" disabled={submitting} style={{ width: "100%" }}>
            {submitting ? "Working…" : mode === "login" ? "Continue" : "Create workspace"}
          </Btn>
        </form>

        <p style={{ margin: "20px 0 0", fontSize: 13.5, color: "var(--adaptive-600)", textAlign: "center" }}>
          {mode === "login" ? (
            <>
              New to Opero?{" "}
              <button onClick={() => switchMode("signup")} style={linkStyle}>
                Create a workspace
              </button>
            </>
          ) : (
            <>
              Already have a workspace?{" "}
              <button onClick={() => switchMode("login")} style={linkStyle}>
                Log in
              </button>
            </>
          )}
        </p>
      </main>

      <footer style={{ textAlign: "center", paddingTop: 24 }}>
        <div style={{ display: "flex", justifyContent: "center", gap: 18, fontSize: 13, marginBottom: 8 }}>
          {["Help", "Privacy", "Terms"].map((l) => (
            <a key={l} href="#" style={{ color: "var(--adaptive-600)", textDecoration: "none" }}>
              {l}
            </a>
          ))}
        </div>
        <div style={{ fontSize: 12, color: "var(--adaptive-400)" }}>
          Copyright © 2026 Opero. All rights reserved.
        </div>
      </footer>
    </div>
  );
}

const linkStyle = {
  border: 0,
  padding: 0,
  background: "none",
  color: "var(--primary-600)",
  fontWeight: 600,
  fontSize: 13.5,
  fontFamily: "inherit",
  cursor: "pointer",
} as const;
