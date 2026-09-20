# mapper-compiler

Build-time schema compiler for Mapper. It reads one YAML model, resolves
stable field IDs, and sends the resolved IR to configured generator plugins.

```bash
go build -o mapper-gen ./cmd/mapper-gen
go build -o mapper-gen-go ./generator/go/cmd/mapper-gen-go

./mapper-gen validate mapper.yaml
./mapper-gen generate --input mapper.yaml --lock mapper.lock.yaml
```

Configuration selects generators and output directories:

```yaml
generators:
  - plugin: go
    out: ./internal/mapper
    options:
      package: mapping
```

Plugins receive versioned JSON on stdin and return generated files on stdout.
The host validates paths, detects collisions, and writes artifacts atomically.
Plugin names resolve through the registry and `mapper-gen-<name>` convention.

Field IDs are generated once (crypto-random, ≤ 2⁵³−1) and pinned in
`*.lock.yaml`. Removed fields stay `removed` and are never recycled.

Module path `github.com/peacewalker122/mapper` is kept so generated Go
descriptors remain compatible with the backend SDK.

## License

Apache-2.0. See [LICENSE](LICENSE).
