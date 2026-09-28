package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
)

type NotificationUsecase interface {
	CreateNotification(ctx context.Context, notification *model.Notification) (*model.Notification, error)
	GetNotification(ctx context.Context, id uint) (*model.Notification, error)
	ListNotifications(ctx context.Context) ([]model.Notification, error)
	UpdateNotification(ctx context.Context, notification *model.Notification) (*model.Notification, error)
	DeleteNotification(ctx context.Context, id uint) error

	Dispatch(ctx context.Context, notification *model.Notification, attempt int, final bool) error
}

type RegisterDeviceInput struct {
	UserID     uint
	Token      string
	DeviceID   string
	DeviceType string
	Email      string
	Phone      string
}

type ContactInput struct {
	UserID uint
	Email  string
	Phone  string
}

type UnregisterDeviceInput struct {
	Token    string
	DeviceID string
	UserID   uint
}

type UserDeviceUsecase interface {
	RegisterDevice(ctx context.Context, in RegisterDeviceInput) (*model.UserDevice, error)
	UnregisterDevice(ctx context.Context, in UnregisterDeviceInput) (int64, error)
	UpsertContact(ctx context.Context, in ContactInput) (*model.UserContact, error)
	GetContact(ctx context.Context, userID uint) (*model.UserContact, error)
	DeactivateStaleTokens(ctx context.Context) (int64, error)
	GetDevice(ctx context.Context, id uint) (*model.UserDevice, error)
	ListDevices(ctx context.Context, userID uint) ([]model.UserDevice, error)
	UpdateDevice(ctx context.Context, device *model.UserDevice) (*model.UserDevice, error)
	DeleteDevice(ctx context.Context, id uint) error
}

type CreateCampaignInput struct {
	Channel      model.CampaignChannel
	Audience     model.CampaignAudience
	Title        string
	Message      string
	Data         map[string]string
	TemplateType string
	TemplateData map[string]string
	Topic        string
	UserIDs      []uint
	Emails       []string
	RequestedBy  string
}

type CampaignView struct {
	Campaign model.Campaign         `json:"campaign"`
	Progress model.CampaignProgress `json:"progress"`
}

type BatchOutcome struct {
	Retry   bool
	Attempt int
}

type CampaignUsecase interface {
	Create(ctx context.Context, in CreateCampaignInput) (*model.Campaign, error)
	Get(ctx context.Context, id uint) (*CampaignView, error)
	List(ctx context.Context, limit, offset int) ([]model.Campaign, error)
	Cancel(ctx context.Context, id uint) (*CampaignView, error)
}

type CampaignRunner interface {
	PlanNext(ctx context.Context) (bool, error)
	RepublishStale(ctx context.Context) (int, error)
	SendBatch(ctx context.Context, batchID uint) (BatchOutcome, error)
}

type AuditQuery struct {
	UserID      uint
	Email       string
	Phone       string
	Channel     string
	Status      string
	SourceType  string
	SourceID    uint
	RequestedBy string
	From        time.Time
	To          time.Time
	BeforeID    uint
	Limit       int
}

type AuditPage struct {
	Items        []model.AuditDelivery       `json:"items"`
	Contents     map[uint]model.AuditContent `json:"contents"`
	NextBeforeID uint                        `json:"next_before_id,omitempty"`
}

type AuditUsecase interface {
	ListDeliveries(ctx context.Context, q AuditQuery) (*AuditPage, error)
	GetContent(ctx context.Context, id uint) (*model.AuditContent, error)
	PurgeExpired(ctx context.Context) (int64, error)
}
