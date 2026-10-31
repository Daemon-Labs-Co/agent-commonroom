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




class AudienceContractTests(unittest.TestCase):
    def test_user_guides_have_local_images_and_small_bash_helpers(self):
        for source in ROOT.joinpath('docs/user').glob('*.md'):
            with self.subTest(source=source.name):
                text = source.read_text()
                images = re.findall(r'!\[([^\]]+)\]\(([^)]+)\)', text)
                self.assertTrue(images, 'Every user page needs an explanatory image and alt text')
                for alt, target in images:
                    self.assertTrue(alt.strip())
                    self.assertFalse(urlsplit(target).scheme, 'Use reviewed local images')
                    self.assertTrue((source.parent / target).resolve().is_file())
                blocks = re.findall(r'^```bash\n(.*?)^```', text, re.MULTILINE | re.DOTALL)
                self.assertTrue(blocks, 'Every user page needs a small Bash helper')
                self.assertTrue(all(len(b.splitlines()) <= 15 for b in blocks))
                if source.name != 'README.md':
                    self.assertIn('## Expected result', text)
                    self.assertIn('scripts/user-guide.sh ' + source.stem, text)

    def test_active_technical_pages_are_contracts(self):
        expected = ['Input', 'Output', 'Functionality', 'Properties and state changes']
        for source in ROOT.joinpath('docs/technical').glob('*.md'):
            with self.subTest(source=source.name):
                actual = re.findall(r'^## (.+)$', source.read_text(), re.MULTILINE)
                self.assertEqual(actual, expected)

    def test_illustrations_are_static_accessible_svg(self):
        import xml.etree.ElementTree as ET
        allowed = {'svg', 'title', 'desc', 'defs', 'marker', 'path', 'rect', 'text', 'g'}
        for source in ROOT.joinpath('docs/images').glob('user-*.svg'):
            with self.subTest(source=source.name):
                tree = ET.fromstring(source.read_text())
                self.assertIsNotNone(tree.find('{http://www.w3.org/2000/svg}title'))
                self.assertIsNotNone(tree.find('{http://www.w3.org/2000/svg}desc'))
                for element in tree.iter():
                    self.assertIn(element.tag.split('}')[-1], allowed)
                    self.assertFalse(any(k.lower().startswith('on') for k in element.attrib))

    def test_navigation_helper_validates_and_uses_only_loopback_help(self):
        import os
        import subprocess
        import tempfile
        script = ROOT / 'scripts/user-guide.sh'
        subprocess.run(['bash', '-n', str(script)], check=True)
        with tempfile.TemporaryDirectory() as directory:
            folder = Path(directory)
            fake = folder / 'curl'
            marker = folder / 'args'
            fake.write_text('#!/bin/sh\nprintf "%s\\n" "$@" > "$DOC_HELPER_ARGS"\nexit "${DOC_HELPER_EXIT:-0}"\n')
            fake.chmod(0o700)
            env = {'PATH': str(folder) + os.pathsep + '/usr/bin:/bin',
                   'DOC_HELPER_ARGS': str(marker)}
            run = lambda args, e=env: subprocess.run(['bash', str(script), *args],
                    env=e, capture_output=True, text=True, timeout=3)
            result = run(['tasks'])
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn('http://127.0.0.1:8787/help/?doc=user-tasks', result.stdout)
            self.assertEqual(marker.read_text().splitlines()[-1], 'http://127.0.0.1:8787/help/')
            self.assertIn('--max-time\n5', marker.read_text())
            marker.unlink()
            for args in [[], ['secret-canary'], ['tasks', 'extra']]:
                result = run(args)
                self.assertEqual(result.returncode, 2)
                self.assertNotIn('secret-canary', result.stdout + result.stderr)
                self.assertFalse(marker.exists())
            result = run(['storage'], dict(env, DOC_HELPER_EXIT='42'))
            self.assertEqual(result.returncode, 1)
            self.assertNotIn('?doc=', result.stdout)


if __name__ == '__main__':
    unittest.main()
