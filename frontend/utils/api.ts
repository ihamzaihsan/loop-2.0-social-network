export const API = "http://localhost:8080";

export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  if (options.body && !(options.body instanceof FormData))
    headers.set("Content-Type", "application/json");
  const response = await fetch(API + path, {
    ...options,
    headers,
    credentials: "include",
  });
  if (!response.ok) {
    const text = await response.text();
    let message = text.trim() || "The request could not be completed.";
    try {
      const error = JSON.parse(text);
      message = error.error || error.message || message;
    } catch {
      /* Plain-text API errors are supported. */
    }
    throw new Error(message);
  }
  return response.json() as Promise<T>;
}

export function refreshResources(...resources: string[]) {
  for (const resource of resources)
    window.dispatchEvent(
      new CustomEvent("realtime_changed", { detail: { resource } }),
    );
}

export function mediaURL(path: string) {
  if (/^https?:\/\//.test(path)) return path;
  const normalized = path.replace(/\\/g, "/").replace(/^\.\//, "");
  return API + "/" + normalized.replace(/^\//, "");
}
