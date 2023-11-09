# gomin-sync
a simple file sync tool based on golang and minio

## Config file

use `config.yaml` as config file, which should be placed under same folder with executable file.

template:

```yaml
endPoint: "example.com:port"
accessUser: "minio123"
accessPassword: "minio123"
bucket: "images"
```