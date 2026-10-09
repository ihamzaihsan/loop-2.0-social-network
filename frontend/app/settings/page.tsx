"use client";
import { imageSizeError } from "../../utils/images";
import { FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import GoogleSignIn from "../../components/GoogleSignIn";
import FeatureShell from "../../components/FeatureShell";
import { useRouter } from "next/navigation";
import { api, refreshResources } from "../../utils/api";
import { useRealtimeRefresh } from "../webscoket/useRealtimeRefresh";
import { WebSocketClient } from "../webscoket/websocket";
type Account = {
  user: {
    firstName: string;
    lastName: string;
    nickname: string;
    aboutMe: string;
    email: string;
  };
  isModerator: boolean;
  hasPassword: boolean;
  googleConnected: boolean;
};
type Block = { id: number; name: string };
export default function Settings() {
  const router = useRouter();
  const [account, setAccount] = useState<Account | null>(null);
  const [blocks, setBlocks] = useState<Block[]>([]);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  async function loadBlocks() {
    const data = await api<Block[]>("/blocks");
    setBlocks(data);
  }
  useEffect(() => {
    api<Account>("/account")
      .then(setAccount)
      .catch((e) => setError(e.message));
    loadBlocks().catch((e) => setError(e.message));
  }, []);
  useRealtimeRefresh(["social"], loadBlocks);
  function logout() {
    localStorage.removeItem("sessionToken");
    WebSocketClient.resetInstance();
    router.push("/login");
  }
  async function submit(e: FormEvent<HTMLFormElement>, method: string) {
    e.preventDefault();
    const data = new FormData(e.currentTarget);
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const avatar = data.get("avatar");
      if (avatar instanceof File) {const sizeError = imageSizeError(avatar);if (sizeError) throw new Error(sizeError);}
      await api("/account", {
        method,
        body:
          method === "PUT" ? data : JSON.stringify(Object.fromEntries(data)),
      });
      if (method === "PUT") {
        setMessage("Profile saved.");
        refreshResources("profiles", "users", "posts");
      } else logout();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unable to save");
    } finally {
      setBusy(false);
    }
  }
  return (
    <FeatureShell title="Settings" activePage="settings">
      {error && (
        <p role="alert" className="feature-error">
          {error}
        </p>
      )}
      {message && <p role="status">{message}</p>}
      {account ? (
        <>
          <section className="feature-card">
            <h2>Edit profile</h2>
            <form className="feature-form" onSubmit={(e) => submit(e, "PUT")}>
              <label>
                First name
                <input
                  name="firstName"
                  defaultValue={account.user.firstName}
                  maxLength={100}
                  required
                />
              </label>
              <label>
                Last name
                <input
                  name="lastName"
                  defaultValue={account.user.lastName}
                  maxLength={100}
                  required
                />
              </label>
              <label>
                Nickname
                <input
                  name="nickname"
                  defaultValue={account.user.nickname || ""}
                  maxLength={100}
                />
              </label>
              <label>
                Biography
                <textarea
                  name="aboutMe"
                  defaultValue={account.user.aboutMe || ""}
                  maxLength={500}
                />
              </label>
              <label>
                Avatar
                <input
                  name="avatar"
                  type="file"
                  accept="image/png,image/jpeg,image/gif"
                />
              </label>
              <label>
                Remove avatar
                <select name="removeAvatar">
                  <option value="false">Keep avatar</option>
                  <option value="true">Remove avatar</option>
                </select>
              </label>
              <button className="primary-button" disabled={busy}>
                Save profile
              </button>
            </form>
          </section>
          <section className="feature-card">
            <h2>Google sign-in</h2>
            <p>{account.googleConnected ? "Google is connected to your account." : "Connect Google to sign in without entering your Loop password."}</p>
            <GoogleSignIn intent={account.googleConnected ? "reauth" : "link"} />
            {!account.hasPassword && <p>Confirm with Google before changing your email, setting a password or deleting your account. Confirmation lasts ten minutes.</p>}
          </section>
          <section className="feature-card">
            <h2>Email and password</h2>
            <p>
              Saving signs you out on all devices. Leave the new password empty
              to keep it.
            </p>
            <form className="feature-form" onSubmit={(e) => submit(e, "POST")}>
              <label>
                Email
                <input
                  name="email"
                  type="email"
                  defaultValue={account.user.email}
                  required
                />
              </label>
              {account.hasPassword && <label>
                Current password
                <input
                  name="currentPassword"
                  type="password"
                  autoComplete="current-password"
                  required
                />
              </label>}
              <label>
                New password
                <input
                  name="password"
                  type="password"
                  autoComplete="new-password"
                  minLength={8}
                  maxLength={72}
                />
              </label>
              <button className="primary-button" disabled={busy}>
                Update account
              </button>
            </form>
          </section>
          <section className="feature-card">
            <h2>Blocked accounts</h2>
            {blocks.length ? (
              blocks.map((b) => (
                <div className="feature-result" key={b.id}>
                  <span>{b.name}</span>
                  <button
                    disabled={busy}
                    onClick={async () => {
                      setBusy(true);
                      try {
                        await api("/blocks?id=" + b.id, { method: "DELETE" });
                        await loadBlocks();
                        refreshResources("social", "users", "posts");
                      } catch (e) {
                        setError(
                          e instanceof Error ? e.message : "Unable to unblock",
                        );
                      } finally {
                        setBusy(false);
                      }
                    }}
                  >
                    Unblock
                  </button>
                </div>
              ))
            ) : (
              <p>No blocked accounts.</p>
            )}
          </section>
          {account.isModerator && (
            <section className="feature-card">
              <Link href="/moderation">Open moderation</Link>
            </section>
          )}
          <section className="feature-card">
            <h2>Delete account</h2>
            <p>
              This permanently removes your account, posts, messages and groups
              you own.
            </p>
            <form
              className="feature-form"
              onSubmit={(e) => submit(e, "DELETE")}
            >
              {account.hasPassword && <label>
                Current password
                <input name="currentPassword" type="password" required />
              </label>}
              <label>
                Type DELETE to confirm
                <input name="confirmation" required pattern="DELETE" />
              </label>
              <button disabled={busy}>Permanently delete account</button>
            </form>
          </section>
        </>
      ) : (
        <p>Loading account?</p>
      )}
    </FeatureShell>
  );
}
