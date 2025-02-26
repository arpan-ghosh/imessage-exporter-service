# imessage-exporter-service
golang service of chat.db from iMessages into s3 for imessages-exporter lambda service
=======
# iMessage Exporter Web Service (Go)

This web service processes iMessage `chat.db` files and extracts messages related to a target phone number using `imessage-exporter`. It then uploads the extracted data to AWS S3.

## 🚀 Features
- Upload `chat.db` via API
- Process messages using `imessage-exporter`
- Store results in AWS S3
- Return URLs for download

## 🛠 Setup
```sh
go mod tidy
go run main.go


