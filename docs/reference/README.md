# Repository file and API reference

Generated from the local checkout with `go run ./tools/source-reference`. Regenerate after changing files.

Every included file has a source link and a purpose. Go entries include every top-level function, receiver method and type, including struct fields and interface methods. Existing comments supply contracts; direct calls expose implementation dependencies. Where a declaration has no prose contract, the reference says so rather than inventing behavior. JavaScript entries index named function declarations, arrow bindings and classes; anonymous callbacks and dynamically assigned properties must be read in the linked source. This is a source map, not a claim that every internal method has a reviewed narrative contract.

## Modules

- [.agents](agents.md): 103 files.
- [.githooks](githooks.md): 1 files.
- [cmd](cmd.md): 183 files.
- [demo](demo.md): 1 files.
- [docker](docker.md): 13 files.
- [docs](docs.md): 58 files.
- [internal](internal.md): 50 files.
- [migrations](migrations.md): 1 files.
- [root](root.md): 15 files.
- [scripts](scripts.md): 6 files.
- [tests](tests.md): 45 files.
- [tools](tools.md): 2 files.
- [web](web.md): 45 files.

Total: 523 files, plus this generated reference directory. Generated reference pages describe themselves here rather than recursively indexing their own contents.

## Local files intentionally represented by category

- `.git/`: Git objects, refs and repository metadata; no application behavior.
- `.data/`, credential directories, private keys, `.env` and local assistant settings: private runtime state/configuration; never publish their contents.
- `.worktrees/`: separate checkouts; document each checkout from its own root.
- `bin/`, `coverage/`, dependency caches, logs and OS metadata: generated outputs; rebuild or regenerate rather than treating them as source.
- Symlinks: not followed, to avoid escaping the checkout or exposing private data.

These categories deliberately do not enumerate private filenames or duplicate generated artifacts.
