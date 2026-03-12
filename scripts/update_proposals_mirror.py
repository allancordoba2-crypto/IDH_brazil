#!/usr/bin/env python3
"""Build proposal summaries and an iCloud-style mirror with links."""

from __future__ import annotations

from collections import defaultdict
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
import os
import re
import shutil


ROOT = Path(__file__).resolve().parents[1]
REPORT_DIR = ROOT / "reports"
MIRROR_ROOT = ROOT / "mirror" / "icloud_files"
INDEX_FILE = MIRROR_ROOT / "INDEX.md"
SUMMARY_REPORT = REPORT_DIR / "proposals_summary.md"

TEXT_EXTENSIONS = {".md", ".markdown", ".txt"}
EXCLUDED_TOP_DIRS = {".git", ".github", "mirror", "reports"}
NAME_KEYWORDS = (
    "proposal",
    "proposta",
    "idea",
    "prevention",
    "policy",
    "plano",
    "bundle",
)
CONTENT_KEYWORDS_RE = re.compile(
    r"\b(proposal|proposta|prevention|policy|idea|preventive|population issue)\b",
    re.IGNORECASE,
)
SENTENCE_SPLIT_RE = re.compile(r"(?<=[.!?])\s+")


@dataclass(frozen=True)
class ProposalRecord:
    source_abs: Path
    source_rel: Path
    title: str
    summary: str
    mirror_summary_abs: Path
    mirror_summary_rel: Path
    mirror_group: str


def _read_text(path: Path) -> str:
    return path.read_text(encoding="utf-8", errors="ignore")


def _clean_spaces(value: str) -> str:
    return re.sub(r"\s+", " ", value).strip()


def _extract_title(text: str, fallback_name: str) -> str:
    for line in text.splitlines():
        striped = line.strip()
        if striped.startswith("#"):
            heading = striped.lstrip("#").strip()
            if heading:
                return heading
    return fallback_name


def _extract_summary(text: str) -> str:
    lines = text.splitlines()
    kept: list[str] = []
    in_code = False

    for raw in lines:
        line = raw.strip()
        if line.startswith("```"):
            in_code = not in_code
            continue
        if in_code or not line:
            continue
        if line.startswith("#"):
            continue
        kept.append(line)

    block = _clean_spaces(" ".join(kept))
    if not block:
        return "No summary content available."

    sentences = [s.strip() for s in SENTENCE_SPLIT_RE.split(block) if s.strip()]
    if not sentences:
        return block[:240]

    summary = " ".join(sentences[:2]).strip()
    if len(summary) > 280:
        summary = summary[:277].rstrip() + "..."
    return summary


def _is_proposal_candidate(path: Path) -> bool:
    if path.suffix.lower() not in TEXT_EXTENSIONS:
        return False

    rel = path.relative_to(ROOT)
    if rel.parts and rel.parts[0] in EXCLUDED_TOP_DIRS:
        return False

    lower_name = path.name.lower()
    if any(token in lower_name for token in NAME_KEYWORDS):
        return True

    try:
        text = _read_text(path)
    except OSError:
        return False
    return bool(CONTENT_KEYWORDS_RE.search(text))


def _iter_proposals() -> list[Path]:
    results: list[Path] = []
    for path in ROOT.rglob("*"):
        if not path.is_file():
            continue
        if _is_proposal_candidate(path):
            results.append(path)
    return sorted(results)


def _relative_link(from_path: Path, to_path: Path) -> str:
    return Path(os.path.relpath(to_path, start=from_path.parent)).as_posix()


def _write_mirror_summary(
    *,
    source_abs: Path,
    source_rel: Path,
    title: str,
    summary: str,
) -> tuple[Path, Path, str]:
    group_rel = source_rel.parent if str(source_rel.parent) != "." else Path("_root")
    mirror_dir = MIRROR_ROOT / group_rel
    mirror_dir.mkdir(parents=True, exist_ok=True)

    mirror_file = mirror_dir / f"{source_abs.stem}.summary.md"
    source_link = _relative_link(mirror_file, source_abs)
    generated = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S UTC")

    mirror_file.write_text(
        "\n".join(
            [
                f"# {title}",
                "",
                f"- Source file: [{source_rel.as_posix()}]({source_link})",
                f"- Updated (UTC): {generated}",
                "",
                "## Summary",
                summary,
                "",
            ]
        ),
        encoding="utf-8",
    )
    return mirror_file, mirror_file.relative_to(ROOT), group_rel.as_posix()


def _escape_table(value: str) -> str:
    return value.replace("|", "\\|")


def _write_index(records: list[ProposalRecord]) -> None:
    groups: dict[str, list[ProposalRecord]] = defaultdict(list)
    for record in records:
        groups[record.mirror_group].append(record)

    generated = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S UTC")
    lines = [
        "# Proposal Mirror Index",
        "",
        f"- Generated (UTC): {generated}",
        f"- Total proposals: {len(records)}",
        "",
        "This folder mirrors proposal files in an iCloud-like structure and links back to source files.",
        "",
    ]

    for group in sorted(groups):
        lines.append(f"## {group}")
        lines.append("")
        for record in sorted(groups[group], key=lambda item: item.source_rel.as_posix()):
            source_link = _relative_link(INDEX_FILE, record.source_abs)
            summary_link = _relative_link(INDEX_FILE, record.mirror_summary_abs)
            lines.append(
                f"- [{record.title}]({source_link}) "
                f"([summary]({summary_link}))"
            )
        lines.append("")

    INDEX_FILE.parent.mkdir(parents=True, exist_ok=True)
    INDEX_FILE.write_text("\n".join(lines), encoding="utf-8")


def _write_summary_report(records: list[ProposalRecord]) -> None:
    generated = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S UTC")
    lines = [
        "# Proposals Summary Report",
        "",
        f"- Generated (UTC): {generated}",
        f"- Total proposals: {len(records)}",
        "",
        "| Source File | Title | Summary | Mirror File |",
        "|---|---|---|---|",
    ]

    for record in records:
        source = record.source_rel.as_posix()
        mirror = record.mirror_summary_rel.as_posix()
        lines.append(
            "| "
            f"`{source}` | "
            f"{_escape_table(record.title)} | "
            f"{_escape_table(record.summary)} | "
            f"`{mirror}` |"
        )

    REPORT_DIR.mkdir(parents=True, exist_ok=True)
    SUMMARY_REPORT.write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> int:
    proposals = _iter_proposals()

    if MIRROR_ROOT.exists():
        shutil.rmtree(MIRROR_ROOT)
    MIRROR_ROOT.mkdir(parents=True, exist_ok=True)

    records: list[ProposalRecord] = []
    for source_abs in proposals:
        source_rel = source_abs.relative_to(ROOT)
        text = _read_text(source_abs)
        title = _extract_title(text, source_abs.stem.replace("_", " "))
        summary = _extract_summary(text)
        mirror_abs, mirror_rel, mirror_group = _write_mirror_summary(
            source_abs=source_abs,
            source_rel=source_rel,
            title=title,
            summary=summary,
        )
        records.append(
            ProposalRecord(
                source_abs=source_abs,
                source_rel=source_rel,
                title=title,
                summary=summary,
                mirror_summary_abs=mirror_abs,
                mirror_summary_rel=mirror_rel,
                mirror_group=mirror_group,
            )
        )

    _write_index(records)
    _write_summary_report(records)

    print(f"Generated proposal mirror for {len(records)} files.")
    print(f"Index: {INDEX_FILE}")
    print(f"Summary report: {SUMMARY_REPORT}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
