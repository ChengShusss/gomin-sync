# gomin-sync
a simple file sync tool based on golang and minio

## Config file

use `config.yaml` as config file, which should be placed under same folder with executable file.

template:

```yaml
endPoint: "10.182.169.156:9010"
accessUser: "minio123"
accessPassword: "minio123"
bucket: "images"
```