# PocketBase compatibility

The initial compatibility baseline is PocketBase 0.39.8.

PocketBase's public health response does not include a reliable server version,
so `pb-agent doctor` probes the operations it needs and reports each capability
as supported, authentication-required, or unknown. Unsupported or changed
response shapes fail closed.

Every release is tested against the pinned baseline. A scheduled CI job tests
the latest PocketBase release and opens an issue when behavior drifts. Support
for a new baseline requires its full end-to-end suite to pass.
