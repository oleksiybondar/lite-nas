#!/usr/bin/python3
"""Send a fixed local port scan through UFW from an isolated test namespace.

This root-owned test-host helper has no caller-supplied network or command
operands. It creates one temporary veth pair, probes the host, then removes
the namespace and link even when a probe fails.
"""

import secrets
import socket
import subprocess
import sys
import time
from ipaddress import IPv4Address

IP_COMMAND = "/usr/sbin/ip"
PYTHON_COMMAND = "/usr/bin/python3"
HELPER_PATH = "/usr/local/libexec/lite-nas-test-port-scan"
NAMESPACE = "litenas-scan-test"
HOST_LINK = "ltnas-scan-host"
PEER_LINK = "ltnas-scan-peer"
HOST_ADDRESS = "198.18.0.1"
FIRST_PORT = 39001
PORT_COUNT = 14
PROBES_BEFORE_REFILL = 10
UFW_REFILL_SECONDS = 45


def run_ip(*arguments: str) -> None:
    """Execute one helper-owned iproute2 operation without a shell."""
    subprocess.run(  # noqa: S603 - fixed executable and helper-owned operands
        [IP_COMMAND, *arguments],
        check=True,
        capture_output=True,
        text=True,
        timeout=150,
    )


def probe(mode: str) -> None:
    """Send a fixed blocked TCP pattern from inside the temporary namespace."""
    for offset in range(PORT_COUNT):
        if offset in (PROBES_BEFORE_REFILL, PROBES_BEFORE_REFILL + 2):
            # UFW's low-volume logging can rate limit bursts. Keep the last
            # four final probes in the detector's two-minute window after refills.
            time.sleep(UFW_REFILL_SECONDS)
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as connection:
            connection.settimeout(0.5)
            host = f"198.18.0.{offset + 1}" if mode == "horizontal" else HOST_ADDRESS
            port = FIRST_PORT + offset if mode == "ports" else FIRST_PORT
            connection.connect_ex((host, port))


def scan(mode: str) -> None:
    """Create local network isolation, run one fixed pattern, then clean up."""
    third_octet = secrets.randbelow(254) + 1
    fourth_octet = secrets.randbelow(254) + 1
    peer_address = str(IPv4Address(f"198.18.{third_octet}.{fourth_octet}"))
    namespace_created = False
    link_created = False
    try:
        run_ip("netns", "add", NAMESPACE)
        namespace_created = True
        run_ip("link", "add", HOST_LINK, "type", "veth", "peer", "name", PEER_LINK)
        link_created = True
        run_ip("link", "set", PEER_LINK, "netns", NAMESPACE)
        run_ip("addr", "add", f"{HOST_ADDRESS}/16", "dev", HOST_LINK)
        if mode == "horizontal":
            for offset in range(1, PORT_COUNT):
                run_ip("addr", "add", f"198.18.0.{offset + 1}/16", "dev", HOST_LINK)
        run_ip("link", "set", HOST_LINK, "up")
        run_ip(
            "netns",
            "exec",
            NAMESPACE,
            IP_COMMAND,
            "addr",
            "add",
            f"{peer_address}/16",
            "dev",
            PEER_LINK,
        )
        run_ip("netns", "exec", NAMESPACE, IP_COMMAND, "link", "set", PEER_LINK, "up")
        run_ip(
            "netns", "exec", NAMESPACE, PYTHON_COMMAND, HELPER_PATH, f"--probe-{mode}"
        )
        sys.stdout.write(f"source_ip={peer_address}\n")
    finally:
        try:
            if link_created:
                run_ip("link", "delete", HOST_LINK)
        finally:
            if namespace_created:
                run_ip("netns", "delete", NAMESPACE)


def main() -> None:
    """Accept only the test runner's scan request or an internal probe call."""
    if sys.argv[1:] == []:
        scan("ports")
    elif sys.argv[1:] in (["--ports"], ["--horizontal"], ["--repeated"]):
        scan(sys.argv[1][2:])
    elif sys.argv[1:] in (
        ["--probe-ports"],
        ["--probe-horizontal"],
        ["--probe-repeated"],
    ):
        probe(sys.argv[1][8:])
    else:
        sys.stderr.write("unsupported test-port-scan arguments\n")
        raise SystemExit(2)


if __name__ == "__main__":
    main()
