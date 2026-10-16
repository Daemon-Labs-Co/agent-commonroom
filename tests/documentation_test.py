"""Offline checks for owned public guides and their local navigation.

No network, database or provider access. Heading IDs match help-reader.js.
"""
from collections import Counter
from pathlib import Path
import re
import unittest
from urllib.parse import unquote, urlsplit

ROOT = Path(__file__).resolve().parents[1]


def documents():
    return sorted([*ROOT.glob('*.md'), *ROOT.joinpath('docs').rglob('*.md'),
                   ROOT / 'docker/engine/README.md'])


def prose(text):
    return re.sub(r'^```[^\n]*\n.*?^```\s*$', '', text,
                  flags=re.MULTILINE | re.DOTALL)


def headings(text):
    counts = Counter()
    result = set()
    for heading in re.findall(r'^#{1,6}\s+(.+)$', prose(text), re.MULTILINE):
        base = ''.join(c for c in heading.lower()
                       if c.isalnum() or c.isspace() or c == '-')
        base = re.sub(r'\s', '-', base)
        count = counts[base]
        counts[base] += 1
        result.add(base + ('-' + str(count) if count else ''))
    return result


class DocumentationTests(unittest.TestCase):
    def test_local_links_and_sections(self):
        failures = []
        for source in documents():
            text = prose(source.read_text())
            for raw in re.findall(r'\[[^\]\n]*\]\(([^)\n]+)\)', text):
                target = raw.strip().strip('<>')
                parsed = urlsplit(target)
                if parsed.scheme or parsed.netloc:
                    continue  # External verification is a separate task.
                destination = (source.parent / unquote(parsed.path)).resolve() if parsed.path else source
                if not destination.exists():
                    failures.append(f'{source.relative_to(ROOT)}: missing {target}')
                    continue
                if parsed.fragment and destination.suffix == '.md':
                    if unquote(parsed.fragment) not in headings(destination.read_text()):
                        failures.append(f'{source.relative_to(ROOT)}: missing section {target}')
        self.assertFalse(failures, '\n' + '\n'.join(failures))

    def test_code_fences_are_closed(self):
        for source in documents():
            with self.subTest(source=source.relative_to(ROOT)):
                fences = re.findall(r'^```', source.read_text(), re.MULTILINE)
                self.assertEqual(len(fences) % 2, 0)

    def test_no_empty_headings(self):
        for source in documents():
            with self.subTest(source=source.relative_to(ROOT)):
                self.assertIsNone(re.search(r'^#{1,6}\s*$', prose(source.read_text()), re.MULTILINE))


if __name__ == '__main__':
    unittest.main()
