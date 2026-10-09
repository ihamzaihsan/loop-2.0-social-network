"use client";
import { FormEvent, useEffect, useState } from "react";
import FeatureShell from "../../components/FeatureShell";
import { api } from "../../utils/api";
import { useRealtimeRefresh } from "../webscoket/useRealtimeRefresh";
type Report = {
  id: number;
  type: string;
  targetId: number;
  reason: string;
  status: string;
  resolution: string;
  reporter: string;
  preview?: string;
};
export default function Moderation() {
  const [reports, setReports] = useState<Report[]>([]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  async function load() {
    try {
      setReports(await api<Report[]>("/moderation"));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unable to load reports");
    }
  }
  useEffect(() => {
    load();
  }, []);
  useRealtimeRefresh(["reports"], load);
  async function act(e: FormEvent<HTMLFormElement>, id: number) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const data = new FormData(e.currentTarget);
    try {
      await api("/moderation", {
        method: "POST",
        body: JSON.stringify({
          id,
          action: data.get("action"),
          note: data.get("note"),
        }),
      });
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Action failed");
    } finally {
      setBusy(false);
    }
  }
  return (
    <FeatureShell title="Moderation" activePage="settings">
      <p>The latest 100 reports. Actions require a moderation note.</p>
      {error && <p role="alert">{error}</p>}
      {reports.map((r) => (
        <article key={r.id} className="feature-card">
          <h2>
            {r.type.replaceAll("_", " ")} #{r.targetId}
          </h2>
          <p>Reported by {r.reporter}</p>
          <blockquote>
            {r.preview || "Content is no longer available."}
          </blockquote>
          <p>{r.reason}</p>
          <p>Status: {r.status}</p>
          {r.status === "open" ? (
            <form className="feature-form" onSubmit={(e) => act(e, r.id)}>
              <label>
                Action
                <select name="action">
                  <option value="dismiss">Dismiss report</option>
                  <option value="resolve">
                    Resolve without removing content
                  </option>
                  {r.type !== "user" && (
                    <option value="remove">Remove content</option>
                  )}
                  <option value="suspend">Suspend author</option>
                </select>
              </label>
              <label>
                Moderation note
                <textarea name="note" maxLength={1000} required />
              </label>
              <button disabled={busy}>Apply action</button>
            </form>
          ) : (
            <p>{r.resolution}</p>
          )}
        </article>
      ))}
      {!error && !reports.length && <p>No reports to review.</p>}
    </FeatureShell>
  );
}
