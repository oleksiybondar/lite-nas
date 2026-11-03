"""Exercise XCCDF category and status comparison on small scan fixtures."""

import tempfile
import unittest
from pathlib import Path

from scap_delta_report import compare, render_html


class ScapDeltaReportTests(unittest.TestCase):
    """Check that paired scan evidence becomes an accurate delta report."""

    def setUp(self) -> None:
        """Create one benchmark and two scans with opposite rule transitions."""
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        root = Path(self.directory.name)
        self.benchmark = root / "benchmark.xml"
        self.before = root / "before.xml"
        self.after = root / "after.xml"
        self.benchmark.write_text(
            '<data-stream xmlns:x="http://checklists.nist.gov/xccdf/1.2">'
            '<x:Benchmark id="ubuntu"><x:Group id="accounts"><x:title>Accounts</x:title>'
            '<x:Rule id="lock" severity="high"><x:title>Lock &amp; review</x:title>'
            '<x:reference href="https://example.org/isa-62443-2013">SR 1.1</x:reference>'
            '</x:Rule></x:Group><x:Group id="network"><x:title>Network</x:title>'
            '<x:Rule id="firewall" severity="medium"><x:title>Firewall</x:title></x:Rule>'
            "</x:Group></x:Benchmark></data-stream>",
            encoding="utf-8",
        )
        self.before.write_text(self.scan("fail", "pass"), encoding="utf-8")
        self.after.write_text(self.scan("pass", "fail"), encoding="utf-8")

    @staticmethod
    def scan(lock: str, firewall: str) -> str:
        """Render a minimal XCCDF TestResult for the two fixture rules."""
        return (
            '<Benchmark xmlns="http://checklists.nist.gov/xccdf/1.2">'
            '<TestResult><profile idref="cis_level2_server"/>'
            f'<rule-result idref="lock"><result>{lock}</result></rule-result>'
            f'<rule-result idref="firewall"><result>{firewall}</result></rule-result>'
            "</TestResult></Benchmark>"
        )

    def test_counts_both_transition_directions(self) -> None:
        """A pass-to-fail regression must not be counted as an improvement."""
        report = compare(self.benchmark, self.before, self.after)
        self.assertEqual((report["totals"]["improved"], report["totals"]["regressed"]), (1, 1))

    def test_keeps_benchmark_category(self) -> None:
        """A rule's benchmark group remains available for the category chart."""
        report = compare(self.benchmark, self.before, self.after)
        self.assertEqual(report["rules"][0]["category"], "Network")

    def test_escapes_rule_titles_in_html(self) -> None:
        """Benchmark text must be escaped in the standalone report."""
        report = compare(self.benchmark, self.before, self.after)
        self.assertIn("Lock &amp; review", render_html(report))


if __name__ == "__main__":
    unittest.main()
