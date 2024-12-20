package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/JIeeiroSst/utils/circuit_breaker"
)

type userProxy struct {
	Host                      string
	ClientCircuitBreakerProxy circuit_breaker.ClientCircuitBreakerProxy
}

type UserProxy interface {
	GetUserInfo(ctx context.Context, userID int) (*UserInfo, error)
}

func NewUserProxy(Host string,
	ClientCircuitBreakerProxy circuit_breaker.ClientCircuitBreakerProxy) UserProxy {
	return &userProxy{
		Host:                      Host,
		ClientCircuitBreakerProxy: ClientCircuitBreakerProxy,
	}
}

func (u *userProxy) GetUserInfo(ctx context.Context, userID int) (*UserInfo, error) {
	path := fmt.Sprintf("%s/%s?user_id=%d", u.Host, "user", userID)
	var response UserResponse
	data, err := u.ClientCircuitBreakerProxy.Send(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data.(string)), &response); err != nil {
		return nil, err
	}

	return &response.UserInfo, nil
}
