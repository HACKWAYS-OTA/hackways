#!/usr/bin/env python3
"""Run the existing migration CLI from a Compose model without logging secrets.

Operations utility only: Go services retain all application/database behavior.
Input is Compose's JSON on stdin. Never print the model or password-bearing URI.
"""

import json
import os
from pathlib import Path
import subprocess
import sys
from urllib.parse import parse_qsl, quote, unquote, urlsplit, urlunsplit


def main() -> int:
    validate_only = len(sys.argv) == 3 and sys.argv[2] == "--validate-only"
    if (len(sys.argv) != 2 and not validate_only) or not Path(sys.argv[1]).is_dir():
        print("Migration directory is missing.", file=sys.stderr)
        return 2
    try:
        model = json.load(sys.stdin)
        services = model["services"]
        raw = services["core-backend"]["environment"]["DATABASE_URL"]
        if raw != services["global-backend"]["environment"]["DATABASE_URL"]:
            raise ValueError("Database targets differ")
        parsed = urlsplit(raw)
        if (
            parsed.scheme not in ("postgres", "postgresql")
            or not parsed.hostname
            or not parsed.username
            or not parsed.path.strip("/")
            or parsed.fragment
            or any(key.lower() == "password" for key, _ in parse_qsl(parsed.query))
        ):
            raise ValueError("Invalid database configuration")
        host = parsed.hostname
        if ":" in host:
            host = f"[{host}]"
        if parsed.port:
            host += f":{parsed.port}"
        username = quote(unquote(parsed.username), safe="")
        passwordless = urlunsplit(
            (parsed.scheme, f"{username}@{host}", parsed.path, parsed.query, "")
        )
        environment = os.environ.copy()
        if parsed.password is not None:
            environment["PGPASSWORD"] = unquote(parsed.password)
        if validate_only:
            print("Migration configuration validated; no connection or migration attempted.")
            return 0
        # golang-migrate's PostgreSQL driver supports the lib/pq PG environment.
        # Keep the credential out of process arguments and diagnostic output.
        result = subprocess.run(
            ["migrate", "-path", sys.argv[1], "-database", passwordless, "up"],
            env=environment,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            timeout=600,
            check=False,
        )
    except (ValueError, KeyError, TypeError, OSError, subprocess.TimeoutExpired):
        print("Migration configuration or execution failed; release stopped.", file=sys.stderr)
        return 1
    if result.returncode:
        # Tool diagnostics can include a URI or data values; do not forward them.
        print(
            f"Migration failed (exit {result.returncode}); inspect migration state before retrying.",
            file=sys.stderr,
        )
        return 1
    print("Pending migrations completed successfully.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
