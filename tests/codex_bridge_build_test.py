"""Synthetic, offline policy-patcher tests; no provider or Docker execution."""
import hashlib
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("bridge_build", ROOT / "docker/engine/lock-codex.py")
bridge = importlib.util.module_from_spec(spec)
spec.loader.exec_module(bridge)

SOURCE = b"before\npub(crate) fn build_tool_router(\n) -> CodexResult<ToolRouter> {\n    native_tools();\n}\nafter\n"
FRAGMENT = "fn commonroom_bridge_router() {}\n"


class BridgeBuildTests(unittest.TestCase):
    def setUp(self):
        self.source_pin = patch.object(bridge, "SOURCE_SHA256", hashlib.sha256(SOURCE).hexdigest())
        self.policy_pin = patch.object(bridge, "FRAGMENT_SHA256", hashlib.sha256(FRAGMENT.encode()).hexdigest())
        self.source_pin.start()
        self.policy_pin.start()
        self.addCleanup(self.source_pin.stop)
        self.addCleanup(self.policy_pin.stop)

    def test_replaces_only_reviewed_factory_and_preserves_surrounding_source(self):
        result = bridge.transform(SOURCE, FRAGMENT)
        self.assertNotIn("native_tools();", result)
        self.assertIn("before\npub(crate)", result)
        self.assertIn("}\nafter\n", result)
        self.assertEqual(result.count(bridge.MARKER), 1)
        self.assertIn("&turn_context.dynamic_tools", result)
        self.assertTrue(result.endswith(FRAGMENT))

    def test_changed_source_or_policy_refused(self):
        for source, policy in [(SOURCE + b" ", FRAGMENT), (SOURCE, FRAGMENT + "native_route();")]:
            with self.subTest(source=source), self.assertRaises(ValueError):
                bridge.transform(source, policy)
        with self.assertRaises(ValueError):
            bridge.transform(bridge.transform(SOURCE, FRAGMENT).encode(), FRAGMENT)

    def test_rejected_policy_preserves_original_file(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            target = root / bridge.SOURCE_FILE
            target.parent.mkdir(parents=True)
            target.write_bytes(SOURCE)
            fragment = root / "policy.rs"
            fragment.write_text(FRAGMENT + "tampered")
            with self.assertRaises(ValueError):
                bridge.apply(root, fragment)
            self.assertEqual(target.read_bytes(), SOURCE)

    def test_symlink_sources_and_policy_refused(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            real = root / "original.rs"
            real.write_bytes(SOURCE)
            target = root / bridge.SOURCE_FILE
            target.parent.mkdir(parents=True)
            target.symlink_to(real)
            policy = root / "policy.rs"
            policy.write_text(FRAGMENT)
            with self.assertRaises(ValueError):
                bridge.apply(root, policy)
            target.unlink()
            target.write_bytes(SOURCE)
            policy.unlink()
            policy.symlink_to(real)
            with self.assertRaises(ValueError):
                bridge.apply(root, policy)
            self.assertEqual(real.read_bytes(), SOURCE)

    def test_checked_in_policy_matches_reviewed_pin(self):
        self.policy_pin.stop()
        policy = (ROOT / "docker/engine/codex-bridge-router.rs").read_bytes()
        self.assertEqual(hashlib.sha256(policy).hexdigest(), bridge.FRAGMENT_SHA256)

    def test_normalization_changes_only_local_versions_and_rejects_drift(self):
        original = b'[[package]]\nname = "local"\nversion = "0.0.0"\n\n[[package]]\nname = "external"\nversion = "0.0.0"\nsource = "registry+example"\nchecksum = "unchanged"\n'
        expected = original.replace(b'name = "local"\nversion = "0.0.0"', b'name = "local"\nversion = "0.162.1"')
        with patch.object(bridge, "LOCK_SHA256", hashlib.sha256(original).hexdigest()), patch.object(bridge, "NORMALIZED_LOCK_SHA256", hashlib.sha256(expected).hexdigest()):
            self.assertEqual(bridge.normalize_lock(original), expected)
            with self.assertRaises(ValueError):
                bridge.normalize_lock(original + b" ")
        with patch.object(bridge, "LOCK_SHA256", hashlib.sha256(original).hexdigest()):
            with self.assertRaises(ValueError):
                bridge.normalize_lock(original)

    def test_bad_lock_never_changes_valid_source(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            target = root / bridge.SOURCE_FILE
            target.parent.mkdir(parents=True)
            target.write_bytes(SOURCE)
            policy = root / "policy.rs"
            policy.write_text(FRAGMENT)
            lock = root / "codex-rs/Cargo.lock"
            lock.write_bytes(b"unreviewed dependencies")
            with self.assertRaises(ValueError):
                bridge.apply(root, policy)
            self.assertEqual(target.read_bytes(), SOURCE)
            self.assertEqual(lock.read_bytes(), b"unreviewed dependencies")

    def test_valid_inputs_prepare_both_files_and_cannot_be_reapplied(self):
        original_lock = b'[[package]]\nname = "local"\nversion = "0.0.0"\n'
        normalized = original_lock.replace(b'"0.0.0"', b'"0.162.1"')
        with tempfile.TemporaryDirectory() as directory, patch.object(bridge, "LOCK_SHA256", hashlib.sha256(original_lock).hexdigest()), patch.object(bridge, "NORMALIZED_LOCK_SHA256", hashlib.sha256(normalized).hexdigest()):
            root = Path(directory)
            target = root / bridge.SOURCE_FILE
            target.parent.mkdir(parents=True)
            target.write_bytes(SOURCE)
            policy = root / "policy.rs"
            policy.write_text(FRAGMENT)
            lock = root / "codex-rs/Cargo.lock"
            lock.write_bytes(original_lock)
            bridge.apply(root, policy)
            expected_source = bridge.transform(SOURCE, FRAGMENT).encode()
            self.assertEqual(target.read_bytes(), expected_source)
            self.assertEqual(lock.read_bytes(), normalized)
            with self.assertRaises(ValueError):
                bridge.apply(root, policy)
            self.assertEqual(target.read_bytes(), expected_source)
            self.assertEqual(lock.read_bytes(), normalized)


if __name__ == "__main__":
    unittest.main()
