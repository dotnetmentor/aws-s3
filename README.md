# aws-s3

A tool for doing common AWS S3 tasks

[![Docker Automated build](https://img.shields.io/docker/cloud/automated/dotnetmentor/aws-s3.svg?style=for-the-badge)](https://hub.docker.com/r/dotnetmentor/aws-s3/)
[![Docker Build Status](https://img.shields.io/docker/cloud/build/dotnetmentor/aws-s3.svg?style=for-the-badge)](https://hub.docker.com/r/dotnetmentor/aws-s3/)
[![MicroBadger Size](https://img.shields.io/microbadger/image-size/dotnetmentor/aws-s3.svg?style=for-the-badge)](https://hub.docker.com/r/dotnetmentor/aws-s3/)
[![Docker Pulls](https://img.shields.io/docker/pulls/dotnetmentor/aws-s3.svg?style=for-the-badge)](https://hub.docker.com/r/dotnetmentor/aws-s3/)

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
goreleaser --rm-dist
```
