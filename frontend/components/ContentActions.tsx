"use client";

import { useState } from "react";
import { FlagIcon } from "@heroicons/react/24/outline";
import ActionDialog from "./ActionDialog";
import { api, refreshResources } from "../utils/api";

export type ContentKind =
  "user" | "post" | "comment" | "group_post" | "group_comment" | "group_event";
type EventValues = { title: string; description: string; eventTime: string };

export default function ContentActions({
  kind,
  id,
  canManage = false,
  content = "",
  eventValues,
  reportable = true,
}: {
  kind: ContentKind;
  id: number;
  canManage?: boolean;
  content?: string;
  eventValues?: EventValues;
  reportable?: boolean;
}) {
  const [action, setAction] = useState<"edit" | "delete" | "report" | null>(
    null,
  );
  const [text, setText] = useState("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [date, setDate] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const open = (value: typeof action) => {
    setAction(value);
    setText(value === "edit" ? content : "");
    setError("");
    setMessage("");
    setTitle(eventValues?.title ?? "");
    setDescription(eventValues?.description ?? "");
    const time = eventValues?.eventTime
      ? new Date(eventValues.eventTime)
      : null;
    setDate(
      time && !Number.isNaN(time.getTime())
        ? new Date(time.getTime() - time.getTimezoneOffset() * 60000)
            .toISOString()
            .slice(0, 16)
        : "",
    );
  };
  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      if (action === "report") {
        await api("/reports", {
          method: "POST",
          body: JSON.stringify({ type: kind, id, reason: text }),
        });
        setMessage("Report submitted for review.");
      } else {
        await api(`/content/manage?type=${kind}&id=${id}`, {
          method: action === "delete" ? "DELETE" : "PUT",
          ...(action === "edit"
            ? {
                body: JSON.stringify(
                  kind === "group_event"
                    ? {
                        title,
                        description,
                        eventTime: new Date(date).toISOString(),
                      }
                    : { content: text },
                ),
              }
            : {}),
        });
        refreshResources("posts", "comments", "groups");
      }
      setAction(null);
    } catch (error) {
      setError(
        error instanceof Error
          ? error.message
          : "Unable to complete the action.",
      );
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="content-actions">
      {canManage && kind !== "post" && kind !== "user" && (
        <>
          <button type="button" onClick={() => open("edit")}>
            Edit
          </button>
          <button type="button" onClick={() => open("delete")}>
            Delete
          </button>
        </>
      )}
      {reportable && (
        <button type="button" className="content-report-button" onClick={() => open("report")}>
          <FlagIcon aria-hidden="true" />
          Report
        </button>
      )}
      {message && <span role="status">{message}</span>}
      {action && (
        <ActionDialog
          title={`${action === "report" ? "Report" : action === "delete" ? "Delete" : "Edit"} ${kind.replaceAll("_", " ")}`}
          onClose={() => {
            if (!busy) setAction(null);
          }}
        >
          <form onSubmit={submit} className="feature-form">
            {action === "delete" ? (
              <p>
                Delete this {kind.replaceAll("_", " ")}? This cannot be undone.
              </p>
            ) : action === "edit" && kind === "group_event" ? (
              <>
                <label>
                  Title
                  <input
                    required
                    maxLength={100}
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                  />
                </label>
                <label>
                  Description
                  <textarea
                    maxLength={500}
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                  />
                </label>
                <label>
                  Date and time
                  <input
                    type="datetime-local"
                    required
                    value={date}
                    onChange={(e) => setDate(e.target.value)}
                  />
                </label>
              </>
            ) : (
              <label>
                {action === "report" ? "Reason for reporting" : "Content"}
                <textarea
                  required
                  maxLength={action === "report" ? 1000 : 100}
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                />
              </label>
            )}
            {error && (
              <p role="alert" className="feature-error">
                {error}
              </p>
            )}
            <button className="primary-button" disabled={busy}>
              {busy
                ? "Saving…"
                : action === "report"
                  ? "Submit report"
                  : action === "delete"
                    ? "Confirm deletion"
                    : "Save changes"}
            </button>
          </form>
        </ActionDialog>
      )}
    </div>
  );
}
