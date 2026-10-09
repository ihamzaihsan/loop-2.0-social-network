"use client";

import { FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import AuthShowcase from "../../../../components/AuthShowcase";
import BrandMark from "../../../../components/BrandMark";
import GoogleSignIn from "../../../../components/GoogleSignIn";
import { api } from "../../../../utils/api";
import "../../../register/register.css";

type Profile = { email: string; firstName: string; lastName: string };

export default function CompleteGoogleProfile() {
  const router = useRouter();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  useEffect(() => {
    api<Profile>("/auth/google/registration").then(setProfile).catch(e => setError(e.message));
  }, []);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const fields = Object.fromEntries(new FormData(e.currentTarget));
    try {
      await api("/auth/google/registration", { method: "POST", body: JSON.stringify(fields) });
      localStorage.removeItem("sessionToken");
      router.replace("/home");
    } catch (e) { setError(e instanceof Error ? e.message : "Unable to create account"); }
    finally { setBusy(false); }
  }
  return (
    <div className="login-container">
      <AuthShowcase mode="register" />
      <div className="login-form-wrapper">
        <div className="auth-mobile-brand"><BrandMark /></div>
        <div className="auth-form-heading">
          <span className="eyebrow">One last step</span>
          <h2 className="form-title">Complete your profile</h2>
          <p>Your new account starts private. You can change this in Settings.</p>
        </div>
        {error && <p role="alert" className="feature-error">{error}</p>}
        {profile ? (
          <form className="feature-form" onSubmit={submit}>
            <p>Google account: {profile.email}</p>
            <label>First name<input name="firstName" defaultValue={profile.firstName} required maxLength={100} autoComplete="given-name" /></label>
            <label>Last name<input name="lastName" defaultValue={profile.lastName} required maxLength={100} autoComplete="family-name" /></label>
            <label>Date of birth<input name="dob" type="date" required min="1900-01-01" max={new Date().toISOString().slice(0, 10)} autoComplete="bday" /></label>
            <label>Nickname (optional)<input name="nickname" maxLength={100} autoComplete="nickname" /></label>
            <button className="submit-button" disabled={busy}>{busy ? "Creating account…" : "Create my Loop account"}</button>
          </form>
        ) : error ? <GoogleSignIn /> : <p role="status">Loading your Google profile…</p>}
        <p><Link href="/login">Back to sign in</Link></p>
      </div>
    </div>
  );
}
