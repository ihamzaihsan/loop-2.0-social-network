"use client";

import { useEffect, useState } from "react";
import { API, api } from "../utils/api";

const errors: Record<string, string> = {
  expired: "Your Google sign-in expired. Please try again.",
  cancelled: "Google sign-in was cancelled. You can try again.",
  failed: "Google sign-in could not be completed. Please try again.",
  unavailable: "This account is unavailable.",
  existing_account: "This email already has a Loop account. Sign in with your password, then connect Google in Settings.",
  linked_elsewhere: "This Google account is already connected to another Loop account.",
  wrong_account: "Choose the Google account already connected to your Loop account.",
};

export default function GoogleSignIn({ intent = "signin" }: { intent?: "signin" | "link" | "reauth" }) {
  const [enabled, setEnabled] = useState<boolean | null>(null);
  const [error, setError] = useState("");
  const [connected, setConnected] = useState(false);
  useEffect(() => {
    let disposed = false;
    api<{ enabled: boolean }>("/auth/google/config")
      .then(data => { if (!disposed) setEnabled(data.enabled); })
      .catch(() => { if (!disposed) setEnabled(false); });
    const params = new URLSearchParams(window.location.search);
    setError(errors[params.get("google_error") || ""] || "");
    setConnected(params.get("google_connected") === "1");
    return () => { disposed = true; };
  }, []);
  return (
    <div className="google-auth-section">
      <form action={`${API}/auth/google/start`} method="get">
      <button
        type="submit"
        name="intent"
        value={intent}
        className="google-auth-button"
        disabled={!enabled}
      >
        <svg viewBox="0 0 48 48" width="20" height="20" aria-hidden="true">
          <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z" />
          <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6C44.4 38.02 46.98 31.84 46.98 24.55z" />
          <path fill="#FBBC05" d="M10.53 28.59A14.41 14.41 0 0 1 9.75 24c0-1.59.27-3.13.76-4.59l-7.98-6.19A23.87 23.87 0 0 0 0 24c0 3.87.93 7.53 2.56 10.78l7.97-6.19z" />
          <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.91-5.8l-7.73-6c-2.15 1.45-4.92 2.3-8.18 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z" />
        </svg>
        {intent === "link" ? "Connect Google" : intent === "reauth" ? "Confirm with Google" : "Continue with Google"}
      </button>
      </form>
      {enabled === false && <p className="google-auth-note">Google sign-in is currently unavailable.</p>}
      {error && <p role="alert" className="feature-error">{error}</p>}
      {connected && <p role="status">Google account confirmed.</p>}
    </div>
  );
}
