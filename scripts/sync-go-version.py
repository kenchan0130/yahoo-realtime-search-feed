#!/usr/bin/env python3
"""Keep the Go language version in go.mod aligned with the Docker builder."""

import argparse
import pathlib
import re
import sys


DOCKER_GO = re.compile(r"^FROM golang:(\d+)\.(\d+)(?:\.\d+)?(?:\s|$)", re.MULTILINE | re.IGNORECASE)
MODULE_GO = re.compile(r"^go (\d+)\.(\d+)(?:\.\d+)?$", re.MULTILINE)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", nargs="?", type=pathlib.Path, default=pathlib.Path("."))
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--check", action="store_true")
    mode.add_argument("--write", action="store_true")
    args = parser.parse_args()

    docker = (args.directory / "Dockerfile").read_text()
    module_path = args.directory / "go.mod"
    module = module_path.read_text()
    docker_match = DOCKER_GO.search(docker)
    module_match = MODULE_GO.search(module)
    if not docker_match or not module_match:
        print("Could not find Go versions in Dockerfile and go.mod", file=sys.stderr)
        return 1

    docker_version = tuple(int(part) for part in docker_match.groups())
    module_version = tuple(int(part) for part in module_match.groups())
    if docker_version == module_version:
        return 0

    expected = ".".join(map(str, docker_version))
    if args.check:
        print(f"go.mod must declare go {expected} to match Dockerfile", file=sys.stderr)
        return 1
    if docker_version < module_version:
        print("Docker builder is older than go.mod; refusing to lower the required Go version", file=sys.stderr)
        return 1

    module_path.write_text(module[: module_match.start()] + f"go {expected}" + module[module_match.end() :])
    print(f"Updated go.mod to go {expected}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
