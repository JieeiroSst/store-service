package model

import (
	"fmt"
	"time"
)

type JobKind string

const (
	JobNotify   JobKind = "NOTIFY"
	JobCallback JobKind = "CALLBACK"
)

type JobState string

const (
	JobPending JobState = "PENDING"
	JobDone    JobState = "DONE"
	JobFailed  JobState = "FAILED"
)

type OutboxJob struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Kind          JobKind   `json:"kind"`
	ShipmentID    int64     `json:"shipment_id"`
	Payload       string    `json:"payload" gorm:"type:text"`
	State         JobState  `json:"state"`
	Attempts      int       `json:"attempts"`
	NextAttemptAt time.Time `json:"next_attempt_at"`
	LastError     string    `json:"last_error"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (OutboxJob) TableName() string { return "outbox_jobs" }

type Notification struct {
	UserID    uint   `json:"user_id"`
	Recipient string `json:"recipient"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Channel   string `json:"type"`
}

type CallbackEvent struct {
	EventID         string    `json:"event_id"`
	ShipmentID      int64     `json:"shipment_id"`
	ShipmentCode    string    `json:"shipment_code"`
	ClientOrderCode string    `json:"client_order_code"`
	Status          Status    `json:"status"`
	PreviousStatus  Status    `json:"previous_status"`
	CarrierStatus   string    `json:"carrier_status"`
	CarrierOrder    string    `json:"carrier_order_code"`
	Reason          string    `json:"reason"`
	OccurredAt      time.Time `json:"occurred_at"`
}

type CallbackJob struct {
	URL   string        `json:"url"`
	Event CallbackEvent `json:"event"`
}

var customerMessages = map[Status]string{
	StatusCreated:        "Đơn hàng %s đã được tạo, đang chờ lấy hàng.",
	StatusPicked:         "Đơn hàng %s đã được lấy hàng và đang trên đường vận chuyển.",
	StatusDelivering:     "Đơn hàng %s đang được giao đến bạn.",
	StatusDelivered:      "Đơn hàng %s đã giao thành công. Cảm ơn bạn!",
	StatusDeliveryFailed: "Giao đơn hàng %s chưa thành công, đơn vị vận chuyển sẽ liên hệ lại.",
	StatusReturning:      "Đơn hàng %s đang được hoàn về người gửi.",
	StatusReturned:       "Đơn hàng %s đã được hoàn về người gửi.",
	StatusCancelled:      "Đơn hàng %s đã bị hủy.",
	StatusException:      "Đơn hàng %s đang gặp sự cố, bộ phận hỗ trợ sẽ liên hệ bạn.",
}

func CustomerNotifications(s Shipment) []Notification {
	tpl, ok := customerMessages[s.Status]
	if !ok {
		return nil
	}
	ref := s.ClientOrderCode
	if ref == "" {
		ref = s.Code
	}
	title := "Cập nhật giao hàng " + ref
	message := fmt.Sprintf(tpl, ref)

	var out []Notification
	if s.Customer.Email != "" {
		out = append(out, Notification{UserID: s.Customer.UserID, Recipient: s.Customer.Email, Title: title, Message: message, Channel: "email"})
	}
	if s.Customer.UserID != 0 {
		out = append(out, Notification{UserID: s.Customer.UserID, Recipient: fmt.Sprint(s.Customer.UserID), Title: title, Message: message, Channel: "push"})
	}
	return out
}
