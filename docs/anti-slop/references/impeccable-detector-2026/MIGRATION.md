# Impeccable 3.6.0 adoption note

- Previous catalog: none (initial deterministic adoption)
- Current catalog: `impeccable@3.6.0`, 59 source rules
- Canonical pack: `ansldes-anti-slop@1.1.0`, 63 rules after Hallmark deduplication and supplements
- Mapping change: Hallmark 8 source rules map to 4 existing Impeccable identities and 4 unique canonical rules
- Behavior boundary: regex fallback is `not-run`; only full provider capability satisfies Web capture coverage
- Consumer action: pin the pack ID, version and fingerprint and generate route-specific evidence outside AnslDes

Future snapshot updates must replace this note with the previous/current package, registry, implementation and mapping
diff, compatibility impact, and required consumer migration. The generator pins this file's SHA-256 together with the
upstream package and snapshot, so updating detector provenance without updating the migration note fails the gate.
