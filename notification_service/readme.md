# notification-service

## API

### Send email

```
POST /api/v1/notifications/email
```

Lưu notification (status `pending`), publish lên RabbitMQ; consumer sẽ render template rồi gửi qua Resend API.

Request:

```json
{
  "user_id": 1,
  "recipient": "someone@example.com",
  "template_type": "welcome",
  "template_data": {
    "Name": "Quan",
    "Email": "someone@example.com"
  },
  "priority": 0
}
```

- `recipient`: bắt buộc, phải là email hợp lệ.
- `template_type`: bắt buộc, tên file (không có `.html`) trong [internal/adapter/secondary/template/templates](internal/adapter/secondary/template/templates). Hiện có: `welcome`, `otp`, `reset_password`.
- `template_data`: `map[string]string`, dùng để thay các placeholder `{{.Key}}` trong template.

Response `202 Accepted`: notification vừa tạo (status `pending`).

Muốn thêm template mới: thả file `.html` mới vào `internal/adapter/secondary/template/templates/`, định nghĩa 2 block `{{define "subject"}}...{{end}}` và `{{define "html"}}...{{end}}`, không cần sửa code Go.

### Send Slack message

```
POST /api/v1/notifications/slack
```

Cùng cơ chế với email: lưu notification, publish RabbitMQ, consumer render template rồi gửi qua Slack webhook.

Request:

```json
{
  "user_id": 1,
  "template_type": "order_status",
  "template_data": {
    "OrderID": "1024",
    "Status": "shipped",
    "CustomerName": "Quan"
  },
  "priority": 0
}
```

- `template_type`: bắt buộc, tên file (không có `.txt`) trong [internal/adapter/secondary/slacktemplate/templates](internal/adapter/secondary/slacktemplate/templates). Hiện có: `alert`, `order_status`, `daily_report`.
- `template_data`: `map[string]string`, dùng để thay các placeholder `{{.Key}}` trong template.

Response `202 Accepted`: notification vừa tạo (status `pending`).

Muốn thêm template mới: thả file `.txt` mới vào `internal/adapter/secondary/slacktemplate/templates/`, định nghĩa 2 block `{{define "title"}}...{{end}}` và `{{define "text"}}...{{end}}` (dùng cú pháp Slack mrkdwn, xem [docs.slack.dev/messaging](https://docs.slack.dev/messaging/)), không cần sửa code Go.
