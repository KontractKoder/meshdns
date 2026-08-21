export interface ServerInfo {
  name: string;
  server_url: string;
  capabilities: string[];
  uptime_30d: number;
  last_checked_at: string;
}

interface ServerListResponse {
  servers: ServerInfo[];
  next_cursor: string;
}

interface ErrorResponse {
  error: {
    code: string;
    detail: string | Record<string, string>;
  };
}

const DEFAULT_TIMEOUT_MS = 5000;

export class MeshDNSError extends Error {
  statusCode: number;
  detail: string | Record<string, string>;

  constructor(statusCode: number, detail: string | Record<string, string>) {
    const detailStr = typeof detail === "string" ? detail : JSON.stringify(detail);
    super(`MeshDNS error ${statusCode}: ${detailStr}`);
    this.name = "MeshDNSError";
    this.statusCode = statusCode;
    this.detail = detail;
  }
}

function stripTrailingSlash(url: string): string {
  return url.endsWith("/") ? url.slice(0, -1) : url;
}

async function fetchWithTimeout(
  url: string,
  options: RequestInit = {},
  timeoutMs: number = DEFAULT_TIMEOUT_MS
): Promise<Response> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
    });
    return response;
  } finally {
    clearTimeout(timeout);
  }
}

async function checkError(response: Response): Promise<void> {
  if (!response.ok) {
    let detail: string | Record<string, string> = response.statusText;
    try {
      const body = (await response.json()) as ErrorResponse;
      if (body?.error?.detail) {
        detail = body.error.detail;
      }
    } catch {
      // use status text if body is not JSON
    }
    throw new MeshDNSError(response.status, detail);
  }
}

export class MeshDNS {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = stripTrailingSlash(baseUrl);
  }

  /**
   * Resolve servers that provide the given capability.
   * Returns active (up) servers ordered by uptime descending.
   */
  async resolve(capability: string): Promise<ServerInfo[]> {
    const url = `${this.baseUrl}/v0/resolve?capability=${encodeURIComponent(capability)}`;
    const response = await fetchWithTimeout(url);
    await checkError(response);
    return (await response.json()) as ServerInfo[];
  }

  /**
   * Resolve servers for a capability, excluding any whose name is in the skip set.
   * Useful for iterating through fallback servers.
   */
  async resolveNext(
    capability: string,
    skip: Set<string>
  ): Promise<ServerInfo[]> {
    const all = await this.resolve(capability);
    return all.filter((s) => !skip.has(s.name));
  }

  /**
   * List servers with optional filters and cursor-based pagination.
   */
  async listServers(opts?: {
    query?: string;
    capability?: string;
    status?: string;
    cursor?: string;
    limit?: number;
  }): Promise<{ servers: ServerInfo[]; nextCursor: string | null }> {
    const params = new URLSearchParams();
    if (opts?.query) params.set("query", opts.query);
    if (opts?.capability) params.set("capability", opts.capability);
    if (opts?.status) params.set("status", opts.status);
    if (opts?.cursor) params.set("cursor", opts.cursor);
    if (opts?.limit != null) params.set("limit", String(opts.limit));

    const qs = params.toString();
    const url = qs
      ? `${this.baseUrl}/v0/servers?${qs}`
      : `${this.baseUrl}/v0/servers`;

    const response = await fetchWithTimeout(url);
    await checkError(response);
    const body = (await response.json()) as ServerListResponse;
    return {
      servers: body.servers,
      nextCursor: body.next_cursor || null,
    };
  }
}