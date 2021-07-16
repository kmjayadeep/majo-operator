# majo-operator
A kubernetes operator to automate database backups to remote storage

note: This project is still in it's early stages. Use in production at your own risk!

## Example

Automate backups from a mongodb database to wasabi s3 bucket every hour

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

## Example 2

Automate backups from a mongodb database to a custom location supported by rclone.
We are using rclone (https://rclone.org/) to do copy the backup files to cloud. You can provide your own rclone config to be able to upload to any destination supported by rclone

Check here for the huge list of supported platforms : <https://rclone.org/overview/>

```
apiVersion: backup.16cloud.online/v1alpha1
kind: MongoBackup
metadata:
  name: majotest
spec:
  host: mongo-mongodb
  database: test
  schedule: "* * * * *"
  rcloneDestination:
    rcloneConfig: "<Rclone config in base64>"
    path: "majo:majooperator.test/"
```
