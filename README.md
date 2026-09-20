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

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/peacewalker122/mapper-compiler/main/install.sh | bash
```

This installs `mapper-gen` and `mapper-gen-go` from GitHub Releases
(checksum-verified). Pin a version with `MAPPER_VERSION=v0.1.0`, change the
target with `MAPPER_INSTALL_DIR`, or run `sh install.sh --help` for options.
Alternatively, download the platform archive from the release manually (see
Releases below) or build from source:

```bash
go build -o mapper-gen ./cmd/mapper-gen
go build -o mapper-gen-go ./generator/go/cmd/mapper-gen-go
```

## Releases

Push a semver tag to publish versioned binaries through GitHub Releases:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Each release includes `mapper-gen` and `mapper-gen-go` for Linux, macOS, and
Windows on amd64 and arm64. Download the platform archive from the release,
extract the binary, and verify it with the published SHA-256 checksums.

Compiler releases are binary-only. The Go module path remains
`github.com/peacewalker122/mapper`; releases do not publish a second Go module.

## License

Apache-2.0. See [LICENSE](LICENSE).
