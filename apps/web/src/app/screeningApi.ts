type APIError = { message?: string; details?: { reason?: string } };

export async function requestJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  if (!response.ok) {
    const error = await response.json().catch(() => null) as APIError | null;
    const message = error?.message || `Request failed (${response.status})`;
    throw new Error(error?.details?.reason ? `${message} (${error.details.reason})` : message);
  }
  return response.json() as Promise<T>;
}
