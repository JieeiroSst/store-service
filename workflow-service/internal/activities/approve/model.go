package approve

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeeiroSst/workflow-service/dto"
)

type Status string

const (
	PENDING Status = "PENDING"
	APPROVE Status = "APPROVE"
	REJECT  Status = "REJECT"
)

type ActiveUser struct {
	ID       string
	Key      string
	Value    string
	User     string
	Status   Status
	CreateAt time.Time
	UpdateAt time.Time
	DeleteAt time.Time
}

func (m Status) String() string {
	switch m {
	case PENDING:
		return "PENDING"
	case APPROVE:
		return "APPROVE"
	case REJECT:
		return "REJECT"
	default:
		return fmt.Sprintf("%v", string(m))
	}
}

func ParseStatus(s string) (c Status, err error) {
	status := map[Status]struct{}{
		PENDING: {},
		APPROVE: {},
		REJECT:  {},
	}
	cap := Status(s)
	_, ok := status[cap]
	if !ok {
		return c, fmt.Errorf(`cannot parse:[%s] as status`, s)
	}
	return cap, nil
}

func AccessTable(processTables []ActiveUser) (statusMaps []map[string]interface{}) {
	for _, value := range processTables {
		b, err := json.Marshal(&value)
		if err != nil {
			return nil
		}
		var statusMap map[string]interface{}
		if err := json.Unmarshal(b, &statusMap); err != nil {
			return nil
		}
		statusMaps = append(statusMaps, statusMap)
	}
	return statusMaps
}

func FormatActiveUser(user ActiveUser) dto.ActiveUser {
	return dto.ActiveUser{
		ID:       user.ID,
		Key:      user.Key,
		Value:    user.Value,
		User:     user.User,
		Status:   string(user.Status),
		CreateAt: user.CreateAt,
		UpdateAt: user.UpdateAt,
		DeleteAt: user.DeleteAt,
	}
}
