#!/usr/bin/env python3
"""Daily metrics rollup for MeshDNS.

Reads the events table from the app's SQLite DB, computes KPIs per SPEC §9-10,
writes metrics.json. Idempotent — overwrites on each run.

Usage: python3 scripts/rollup_metrics.py [db_path]
  Default: meshdns.db in current directory.
"""

import json, os, sqlite3, sys, datetime

def rollup(db_path: str) -> dict:
    db = sqlite3.connect(db_path)
    db.row_factory = sqlite3.Row

    now = datetime.datetime.utcnow()
    today = now.strftime("%Y-%m-%d")
    since_24h = (now - datetime.timedelta(hours=24)).isoformat() + "Z"

    registrations = db.execute("SELECT count(*) FROM events WHERE type='register'").fetchone()[0]
    resolutions_24h = db.execute("SELECT count(*) FROM events WHERE type='resolve' AND ts >= ?", (since_24h,)).fetchone()[0]
    distinct_sources_24h = len({r[0] for r in db.execute(
        "SELECT json_extract(payload, '$.source_hash') FROM events WHERE type='resolve' AND ts >= ?",
        (since_24h,)).fetchall() if r[0]})

    probes_24h = db.execute("SELECT count(*) FROM events WHERE type='probe' AND ts >= ?", (since_24h,)).fetchone()[0]

    # false-down estimate: probes that were down but server came back within 5 min
    false_down = db.execute("""
        SELECT count(*) FROM probes p1 JOIN probes p2 ON p1.server_id=p2.server_id
        WHERE p1.up=0 AND p2.up=1 AND p1.id < p2.id
        AND strftime('%s', p2.ts) - strftime('%s', p1.ts) BETWEEN 0 AND 300
    """).fetchone()[0]

    servers_active = db.execute("SELECT count(*) FROM servers WHERE status='active'").fetchone()[0]

    db.close()

    return {
        "date": today,
        "generated_at": now.isoformat() + "Z",
        "registrations_total": registrations,
        "resolutions_24h": resolutions_24h,
        "distinct_resolving_sources_24h": distinct_sources_24h,
        "probes_24h": probes_24h,
        "false_down_estimate": false_down,
        "servers_active": servers_active,
    }

def main():
    db_path = sys.argv[1] if len(sys.argv) > 1 else "meshdns.db"
    if not os.path.exists(db_path):
        print(f"metrics: db not found at {db_path}, skipping")
        return

    metrics = rollup(db_path)
    out_path = os.path.join(os.path.dirname(db_path) or ".", "metrics.json")
    with open(out_path, "w") as f:
        json.dump(metrics, f, indent=2)
    print(f"metrics: {json.dumps(metrics)}")

if __name__ == "__main__":
    main()