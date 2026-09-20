# mapper-compiler

Schema compiler for the Mapper platform: authors a model once in YAML and
generates a stable-identity Go model plus a runtime schema descriptor.

```bash
go build -o mapper-gen ./cmd/mapper-gen
./mapper-gen validate schema/subscriber.yaml
./mapper-gen generate --input schema/subscriber.yaml \
  --output generated/subscriber.gen.go --package generated
```

Field IDs are generated once (crypto-random, ≤ 2⁵³−1) and pinned in
`*.lock.yaml`. Removed fields stay `removed` and are never recycled.

Split from the [mapper](https://github.com/peacewalker122/mapper) monorepo.
Module path `github.com/peacewalker122/mapper` is kept so the backend SDK
can consume generated descriptors unchanged.

## License

Apache-2.0. See [LICENSE](LICENSE).
