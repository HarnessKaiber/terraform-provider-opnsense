#!/usr/bin/env python3
"""Install pinned Zenarmor packages in the credential-free CI image."""

import base64
import json
import os
import select
import socket
import sys
import time

QEMU_GA_SOCKET = os.environ.get("QEMU_GA_SOCKET", "/tmp/qemu-virtserialport.sock")


def send(command, timeout=10):
    with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as client:
        client.settimeout(timeout)
        client.connect(QEMU_GA_SOCKET)
        client.sendall((json.dumps(command) + "\n").encode())
        chunks = b""
        while True:
            ready, _, _ = select.select([client], [], [], timeout)
            if not ready:
                break
            chunk = client.recv(8192)
            if not chunk:
                break
            chunks += chunk
            if b"\n" in chunk:
                break
        return chunks.decode().strip()


def guest_exec(path, args, max_wait_s=900):
    response = send({
        "execute": "guest-exec",
        "arguments": {"path": path, "arg": args, "capture-output": True},
    })
    pid = json.loads(response)["return"]["pid"]
    deadline = time.monotonic() + max_wait_s
    data = {}
    while time.monotonic() < deadline:
        time.sleep(5)
        data = json.loads(send({
            "execute": "guest-exec-status",
            "arguments": {"pid": pid},
        })).get("return", {})
        if data.get("exited"):
            break
    stdout = base64.b64decode(data.get("out-data", "")).decode("utf-8", errors="replace")
    stderr = base64.b64decode(data.get("err-data", "")).decode("utf-8", errors="replace")
    return data.get("exitcode", -1), stdout, stderr


def run_pkg(args, description, max_wait_s=900):
    print(description, flush=True)
    code, stdout, stderr = guest_exec("/usr/sbin/pkg", args, max_wait_s=max_wait_s)
    if stdout:
        print(stdout)
    if stderr:
        print(stderr, file=sys.stderr)
    if code != 0:
        sys.exit(f"{description} failed with exit code {code}")


def main():
    run_pkg(["install", "-y", "os-sunnyvalley"], "Installing Sunny Valley repository")
    run_pkg(["update", "-f"], "Refreshing repositories", max_wait_s=300)
    run_pkg(
        ["install", "-y", "os-sensei-2.6.2", "os-sensei-agent-2.6.1", "os-sensei-updater-2.0"],
        "Installing pinned Zenarmor packages",
    )


if __name__ == "__main__":
    main()
