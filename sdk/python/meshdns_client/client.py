"""MeshDNS HTTP client with resolve, resolve_next, and list_servers."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Optional

import httpx


@dataclass
class ServerInfo:
    """A server returned by MeshDNS resolution or listing."""

    name: str
    server_url: str
    capabilities: list[str] = field(default_factory=list)
    uptime_30d: float = 0.0
    last_checked_at: str = ""


class MeshDNSError(Exception):
    """Raised on 4xx/5xx responses from the MeshDNS server."""

    def __init__(self, status_code: int, detail: str) -> None:
        self.status_code = status_code
        self.detail = detail
        super().__init__(f"MeshDNS {status_code}: {detail}")


class MeshDNS:
    """Synchronous HTTP client for MeshDNS.

    Args:
        base_url: MeshDNS server base URL (e.g. ``http://localhost:8080``).
    """

    def __init__(self, base_url: str) -> None:
        self._base = base_url.rstrip("/")
        self._client = httpx.Client(
            base_url=self._base,
            timeout=httpx.Timeout(5.0),
        )

    def resolve(self, capability: str) -> list[ServerInfo]:
        """Resolve servers advertising *capability*.

        Returns only servers that are currently up and active, ordered by
        descending 30-day uptime then by name.
        """
        r = self._client.get("/v0/resolve", params={"capability": capability})
        self._raise_for_error(r)
        return [_parse_server_json(obj) for obj in r.json()]

    def resolve_next(
        self, capability: str, skip: frozenset[str]
    ) -> list[ServerInfo]:
        """Like :meth:`resolve` but excludes servers whose **name** is in *skip*."""
        servers = self.resolve(capability)
        return [s for s in servers if s.name not in skip]

    def list_servers(
        self,
        query: str = "",
        capability: str = "",
        status: str = "",
        cursor: str = "",
        limit: int = 20,
    ) -> tuple[list[ServerInfo], Optional[str]]:
        """List servers with optional filtering and cursor-based pagination.

        Args:
            query: Free-text search across name and description.
            capability: Filter to servers with a specific capability.
            status: One of ``"active"``, ``"delisted"``, or ``"all"``.
            cursor: Opaque pagination cursor from a previous response.
            limit: Page size (capped at 100).

        Returns:
            A ``(servers, next_cursor)`` tuple.  *next_cursor* is ``None``
            when there are no more pages.
        """
        if limit > 100:
            limit = 100
        if limit < 1:
            limit = 1

        r = self._client.get(
            "/v0/servers",
            params={
                "query": query,
                "capability": capability,
                "status": status,
                "cursor": cursor,
                "limit": str(limit),
            },
        )
        self._raise_for_error(r)
        body = r.json()
        servers = [_parse_server_json(obj) for obj in body.get("servers", [])]
        next_cursor = body.get("next_cursor") or None
        return servers, next_cursor

    def close(self) -> None:
        """Close the underlying HTTP client."""
        self._client.close()

    def _raise_for_error(self, r: httpx.Response) -> None:
        if r.is_success:
            return
        detail = "unknown error"
        try:
            err_body = r.json()
            detail = err_body.get("error", {}).get("detail", str(err_body))
        except Exception:
            detail = r.text or f"HTTP {r.status_code}"
        raise MeshDNSError(r.status_code, detail)


def _parse_server_json(obj: dict) -> ServerInfo:
    return ServerInfo(
        name=obj.get("name", ""),
        server_url=obj.get("server_url", ""),
        capabilities=list(obj.get("capabilities", [])),
        uptime_30d=float(obj.get("uptime_30d", 0.0)),
        last_checked_at=str(obj.get("last_checked_at", "")),
    )