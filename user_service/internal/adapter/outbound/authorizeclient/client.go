// Package authorizeclient is the gRPC adapter for authorize-service's
// PermissionService. Regenerate pb/ with `make gen-authorize-client` after
// syncing proto/permission.proto from authorize_service/api/permission.
package authorizeclient

import (
	"context"
	"strconv"
	"time"

	permissionpb "github.com/JIeeiroSst/user-service/internal/adapter/outbound/authorizeclient/pb"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	api     permissionpb.PermissionServiceClient
	timeout time.Duration
}

// New creates a lazily-connecting client; it does not block on dial.
func New(address string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:    conn,
		api:     permissionpb.NewPermissionServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) AssignRoles(ctx context.Context, userID int, roles ...string) (domain.UserRoles, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := c.api.AssignRoles(ctx, &permissionpb.AssignRolesRequest{UserId: strconv.Itoa(userID), Roles: roles})
	if err != nil {
		return domain.UserRoles{}, err
	}
	return toUserRoles(res), nil
}

func (c *Client) SetUserRoles(ctx context.Context, userID int, roles ...string) (domain.UserRoles, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := c.api.SetUserRoles(ctx, &permissionpb.SetUserRolesRequest{UserId: strconv.Itoa(userID), Roles: roles})
	if err != nil {
		return domain.UserRoles{}, err
	}
	return toUserRoles(res), nil
}

func (c *Client) RemoveUser(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.api.RemoveUser(ctx, &permissionpb.RemoveUserRequest{UserId: strconv.Itoa(userID)})
	return err
}

func (c *Client) GetUserRoles(ctx context.Context, userID int) (domain.UserRoles, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := c.api.GetUserRoles(ctx, &permissionpb.GetUserRolesRequest{UserId: strconv.Itoa(userID)})
	if err != nil {
		return domain.UserRoles{}, err
	}
	return toUserRoles(res), nil
}

func (c *Client) RoleExists(ctx context.Context, role string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	res, err := c.api.ListRoles(ctx, &permissionpb.ListRolesRequest{})
	if err != nil {
		return false, err
	}
	for _, r := range res.GetRoles() {
		if r.GetName() == role {
			return true, nil
		}
	}
	return false, nil
}

func toUserRoles(res *permissionpb.UserRolesResponse) domain.UserRoles {
	return domain.UserRoles{
		Roles:          res.GetRoles(),
		EffectiveRoles: res.GetEffectiveRoles(),
		PrimaryRole:    res.GetPrimaryRole(),
	}
}
