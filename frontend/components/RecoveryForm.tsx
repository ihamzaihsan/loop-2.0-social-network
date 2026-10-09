"use client";
import { FormEvent, useState } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { api } from "../utils/api";
export default function RecoveryForm({ reset = false }: { reset?: boolean }) {
  const params = useSearchParams();
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const data = new FormData(e.currentTarget);
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await api(reset ? "/reset-password" : "/forgot-password", {
        method: "POST",
        body: JSON.stringify(
          reset
            ? {
                token: params.get("token") || "",
                password: data.get("password"),
              }
            : { email: data.get("email") },
        ),
      });
      setMessage(
        reset
          ? "Password updated. Sign in with your new password."
          : "If the account exists, a reset link has been sent. Check your inbox.",
      );
    } catch (e) {
      setError(e instanceof Error ? e.message : "Request failed");
    } finally {
      setBusy(false);
    }
  }
  return (
    <main className="recovery-page">
      <section className="feature-card">
        <h1>{reset ? "Reset password" : "Forgot password"}</h1>
        <p>Reset links expire after 30 minutes and can be used once.</p>
        <form className="feature-form" onSubmit={submit}>
          {reset ? (
            <label>
              New password
              <input
                name="password"
                type="password"
                minLength={8}
                maxLength={72}
                autoComplete="new-password"
                required
              />
            </label>
          ) : (
            <label>
              Email
              <input name="email" type="email" autoComplete="email" required />
            </label>
          )}
          <button className="primary-button" disabled={busy}>
            {busy
              ? "Please wait?"
              : reset
                ? "Save password"
                : "Send reset link"}
          </button>
        </form>
        {error && (
          <p role="alert" className="feature-error">
            {error}
          </p>
        )}
        {message && <p role="status">{message}</p>}
        <Link href="/login">Back to sign in</Link>
      </section>
    </main>
  );
}
