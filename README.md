# majo-operator
A kubernetes operator to automate database backups to remote storage

note: This project is still in it's early stages. Use in production at your own risk!

## Example

Automate backups from a mongodb databse to wasabi s3 bucket every hour

```
apiVersion: backup.16cloud.online/v1alpha1
kind: MongoBackup
metadata:
  name: backup-mongo
spec:
  host: mongo-mongodb
  database: test
  schedule: "0 * * * *"
  s3Destination:
    accessKeyId: <S3 Access Key>
    secretAccessKey: <Secret Access Key>
    bucket: "majooperator.test"
    endpoint: s3.wasabisys.com
```


