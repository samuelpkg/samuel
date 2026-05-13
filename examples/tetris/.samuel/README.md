# .samuel/

This directory contains project-level Samuel state. The framework owns
the layout; the contents are yours to read and (mostly) edit.

```
.samuel/
├── tasks/      # PRDs and task lists (you author these)
├── builtins/   # local copy of Samuel's embedded built-ins (do not edit)
└── plugins/    # discovered plugin manifests (populated in Milestone 3)
```

The framework treats `.samuel/builtins/` as immutable; if you need to
customize a built-in, fork it as a regular skill under `.samuel/plugins/`
(see the `create-skill` built-in).
