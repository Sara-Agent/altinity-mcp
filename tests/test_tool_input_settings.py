"""Run the Go acceptance checks using only the Python standard library."""

from pathlib import Path
import subprocess
import sys


ROOT = Path(__file__).resolve().parents[1]


def _run(case):
    result = subprocess.run(
        ["go", "test", "./pkg/server/", "-count=1", "-run",
         f"TestValidateToolInputSettings/{case}$"],
        cwd=ROOT,
        check=False,
    )
    if result.returncode:
        raise SystemExit(result.returncode)


def test_sql_underscore_silent():
    _run("sql_underscore_silent")


def test_sql_underscore_any_case():
    _run("sql_underscore_any_case")


def test_non_sql_prefix_warns():
    _run("non_sql_prefix_warns")


def test_custom_prefix_warns():
    _run("custom_prefix_warns")


def test_blocked_readonly_still_errors():
    _run("blocked_readonly")


if __name__ == "__main__":
    globals()[sys.argv[1]]()
