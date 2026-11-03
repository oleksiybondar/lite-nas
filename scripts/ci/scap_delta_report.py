#!/usr/bin/env python3
"""Compare two OpenSCAP XCCDF evaluations and render release evidence."""

import argparse
import hashlib
import html
import json
import xml.etree.ElementTree as ET
from collections import Counter, defaultdict
from pathlib import Path


def local_name(tag: str) -> str:
    """Return an XML tag name without its namespace."""
    return tag.rsplit("}", 1)[-1]


def children(element: ET.Element, name: str) -> list[ET.Element]:
    """Return direct children with the requested local tag name."""
    return [child for child in element if local_name(child.tag) == name]


def child_text(element: ET.Element, name: str) -> str:
    """Read normalized text from the first direct child of an element."""
    found = children(element, name)
    return " ".join("".join(found[0].itertext()).split()) if found else ""


def benchmark_rules(path: Path) -> dict[str, dict[str, object]]:
    """Read rule titles, severity, references, and top-level XCCDF groups."""
    root = ET.parse(path).getroot()
    benchmarks = [element for element in root.iter() if local_name(element.tag) == "Benchmark"]
    if len(benchmarks) != 1:
        raise ValueError(f"Expected one benchmark in {path}, found {len(benchmarks)}")

    rules: dict[str, dict[str, object]] = {}

    def visit(element: ET.Element, groups: list[str]) -> None:
        """Carry group titles down the benchmark tree to each rule."""
        for item in element:
            tag = local_name(item.tag)
            if tag == "Group":
                title = child_text(item, "title") or item.get("id", "Unnamed group")
                visit(item, [*groups, title])
            elif tag == "Rule":
                rule_id = item.get("id", "")
                if not rule_id or rule_id in rules:
                    raise ValueError(f"Missing or duplicate benchmark rule ID: {rule_id}")
                references = [
                    {
                        "source": reference.get("href", ""),
                        "id": " ".join(reference.itertext()).strip(),
                    }
                    for reference in children(item, "reference")
                    if "62443" in reference.get("href", "")
                ]
                rules[rule_id] = {
                    "title": child_text(item, "title") or rule_id,
                    "severity": item.get("severity", "unknown"),
                    "category": groups[0] if groups else "Other",
                    "group_path": groups,
                    "iec_62443_references": references,
                }
            elif tag in {"Benchmark", "Value"}:
                visit(item, groups)

    visit(benchmarks[0], [])
    if not rules:
        raise ValueError(f"No benchmark rules found in {path}")
    return rules


def scan_results(path: Path) -> tuple[str, dict[str, str]]:
    """Read the last XCCDF TestResult and require unique rule outcomes."""
    root = ET.parse(path).getroot()
    tests = [element for element in root.iter() if local_name(element.tag) == "TestResult"]
    if not tests:
        raise ValueError(f"No XCCDF TestResult found in {path}")
    test = tests[-1]
    profile = next((element.get("idref", "") for element in children(test, "profile")), "")
    results: dict[str, str] = {}
    for item in children(test, "rule-result"):
        rule_id = item.get("idref", "")
        result = child_text(item, "result").lower()
        if not rule_id or not result or rule_id in results:
            raise ValueError(f"Missing or duplicate result for rule {rule_id} in {path}")
        results[rule_id] = result
    if not results:
        raise ValueError(f"No rule results found in {path}")
    return profile, results


def change(before: str, after: str) -> str:
    """Classify a rule transition without treating unassessed states as passes."""
    if before == "fail" and after == "pass":
        return "improved"
    if before == "pass" and after == "fail":
        return "regressed"
    if before == "fail" and after == "fail":
        return "still failing"
    if before == "pass" and after == "pass":
        return "still passing"
    return "other change" if before != after else "unassessed"


def compare(benchmark: Path, before: Path, after: Path) -> dict[str, object]:
    """Build a stable JSON record from matching scans of one benchmark."""
    rules = benchmark_rules(benchmark)
    before_profile, before_results = scan_results(before)
    after_profile, after_results = scan_results(after)
    if not before_profile or before_profile != after_profile:
        raise ValueError("The scans must use the same nonempty XCCDF profile")
    if before_results.keys() != after_results.keys():
        raise ValueError("The scans evaluated different rule sets")
    unknown_rules = before_results.keys() - rules.keys()
    if unknown_rules:
        raise ValueError(
            f"Scan results contain rules absent from the benchmark: {sorted(unknown_rules)}"
        )

    rows = []
    for rule_id in sorted(before_results):
        metadata = rules[rule_id]
        old, new = before_results[rule_id], after_results[rule_id]
        rows.append(
            {
                "id": rule_id,
                "title": metadata["title"],
                "severity": metadata["severity"],
                "category": metadata["category"],
                "group_path": metadata["group_path"],
                "iec_62443_references": metadata["iec_62443_references"],
                "before": old,
                "after": new,
                "change": change(old, new),
            }
        )
    counts = Counter(row["change"] for row in rows)
    return {
        "schema_version": 1,
        "benchmark_sha256": hashlib.sha256(benchmark.read_bytes()).hexdigest(),
        "profile": before_profile,
        "totals": {
            name: counts[name]
            for name in (
                "improved",
                "regressed",
                "still failing",
                "still passing",
                "other change",
                "unassessed",
            )
        },
        "after_failures": sum(row["after"] == "fail" for row in rows),
        "rules": rows,
    }


def render_html(data: dict[str, object]) -> str:
    """Render a standalone category chart and per-rule comparison table."""
    rows = data["rules"]
    by_category: dict[str, Counter[str]] = defaultdict(Counter)
    for row in rows:
        by_category[row["category"]][row["change"]] += 1
    maximum = max((sum(counts.values()) for counts in by_category.values()), default=1)
    totals = data["totals"]
    summary = "".join(
        f"<div class='card'><strong>{totals[label]}</strong><span>{html.escape(label.title())}</span></div>"
        for label in ("improved", "regressed", "still failing", "other change", "unassessed")
    )
    bars = []
    for category, counts in sorted(by_category.items()):
        segments = "".join(
            f"<span class='{css}' style='width:{counts[label] / maximum * 100:.2f}%' "
            f"title='{html.escape(label)}: {counts[label]}'></span>"
            for label, css in (
                ("improved", "good"),
                ("regressed", "bad"),
                ("still failing", "fail"),
                ("still passing", "pass"),
                ("other change", "other"),
                ("unassessed", "other"),
            )
            if counts[label]
        )
        bars.append(
            f"<tr><th>{html.escape(category)}</th><td><div class='bar'>{segments}</div></td>"
            f"<td>{counts['improved']}</td><td>{counts['regressed']}</td>"
            f"<td>{counts['still failing']}</td></tr>"
        )
    details = []
    for row in rows:
        references = ", ".join(ref["id"] for ref in row["iec_62443_references"])
        cells = (
            row["category"],
            row["title"],
            row["severity"],
            row["before"],
            row["after"],
            row["change"],
            references,
        )
        details.append(
            "<tr>" + "".join(f"<td>{html.escape(str(cell))}</td>" for cell in cells) + "</tr>"
        )
    return f"""<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<title>LiteNAS OpenSCAP before and after</title>
<style>
body{{font:16px/1.45 system-ui,sans-serif;max-width:1200px;margin:2rem auto;padding:0 1rem;color:#172033}}
.cards{{display:flex;flex-wrap:wrap;gap:.8rem}}.card{{border:1px solid #ccd3df;border-radius:.5rem;padding:.7rem 1rem;min-width:9rem}}
.card strong{{display:block;font-size:1.6rem}}.card span{{color:#43506a}}table{{border-collapse:collapse;width:100%;margin:1rem 0}}
th,td{{border-bottom:1px solid #d9dfeb;padding:.45rem;text-align:left;vertical-align:top}}.bar{{display:flex;height:1.2rem;min-width:9rem}}
.good{{background:#237a4b}}.bad{{background:#b42332}}.fail{{background:#e78a32}}.pass{{background:#91b99c}}.other{{background:#8792a6}}
.scroll{{overflow-x:auto}}code{{overflow-wrap:anywhere}}small{{color:#43506a}}
</style></head><body>
<h1>LiteNAS OpenSCAP before and after</h1>
<p>Profile: <code>{html.escape(str(data['profile']))}</code><br>
Benchmark SHA-256: <code>{html.escape(str(data['benchmark_sha256']))}</code></p>
<p>This compares checks on one GitHub-hosted Ubuntu VM before and after installing LiteNAS. It is not an IEC 62443 security-level assessment.</p>
<section class="cards">{summary}</section>
<p><strong>{data['after_failures']}</strong> checks fail after installation. Results that are not pass or fail need separate review.</p>
<h2>Changes by benchmark category</h2>
<p><small>Green: improved · Red: regressed · Orange: still failing · Pale green: still passing · Gray: other or unassessed</small></p>
<div class="scroll"><table><thead><tr><th>Category</th><th>Rule outcomes</th><th>Improved</th><th>Regressed</th><th>Still failing</th></tr></thead>
<tbody>{''.join(bars)}</tbody></table></div>
<h2>Every evaluated rule</h2><div class="scroll"><table><thead><tr><th>Category</th><th>Rule</th><th>Severity</th><th>Before</th><th>After</th><th>Change</th><th>IEC 62443 references in benchmark</th></tr></thead>
<tbody>{''.join(details)}</tbody></table></div></body></html>"""


def main() -> None:
    """Validate CLI paths and write both machine-readable and HTML reports."""
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--benchmark", type=Path, required=True)
    parser.add_argument("--before", type=Path, required=True)
    parser.add_argument("--after", type=Path, required=True)
    parser.add_argument("--json", type=Path, required=True)
    parser.add_argument("--html", type=Path, required=True)
    args = parser.parse_args()
    for path in (args.benchmark, args.before, args.after):
        if not path.is_file():
            parser.error(f"Input file does not exist: {path}")
    if args.json == args.html:
        parser.error("JSON and HTML output paths must differ")
    data = compare(args.benchmark, args.before, args.after)
    args.json.parent.mkdir(parents=True, exist_ok=True)
    args.html.parent.mkdir(parents=True, exist_ok=True)
    args.json.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    args.html.write_text(render_html(data), encoding="utf-8")


if __name__ == "__main__":
    main()
