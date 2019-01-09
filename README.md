# aws-s3

A tool for doing common AWS S3 tasks


## Prune objects

```bash
./aws-s3 prune -bucket my-bucket -region us-west-1 -prefix daily/ -max-age 168h -dry-run
```
