# aws-s3

A tool for doing common AWS S3 tasks

## Commands

### Prune command

```bash
./aws-s3 prune -bucket my-bucket -region us-west-1 -prefix daily/ -max-age 168h -dry-run
```

## Development

```bash
./run <command> <flags>
```

## Releasing

Pre-requisites:

- [goreleaser](https://goreleaser.com/)
- A github token

```bash
goreleaser release --rm-dist
```
