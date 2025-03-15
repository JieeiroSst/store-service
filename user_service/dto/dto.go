package dto

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LoginRequest struct {
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
}

type LoginResponse struct {
	SessionToken string `protobuf:"bytes,1,opt,name=session_token,json=sessionToken,proto3" json:"session_token,omitempty"`
	RefreshToken string `protobuf:"bytes,2,opt,name=refresh_token,json=refreshToken,proto3" json:"refresh_token,omitempty"`
	ExpiryTime   int64  `protobuf:"varint,3,opt,name=expiry_time,json=expiryTime,proto3" json:"expiry_time,omitempty"`
}

type LogoutRequest struct {
	SessionToken string `protobuf:"bytes,1,opt,name=session_token,json=sessionToken,proto3" json:"session_token,omitempty"`
}

type LogoutResponse struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

type ValidateRequest struct {
	SessionToken string `protobuf:"bytes,1,opt,name=session_token,json=sessionToken,proto3" json:"session_token,omitempty"`
}

type ValidateResponse struct {
	Valid  bool   `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
	UserId string `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `protobuf:"bytes,1,opt,name=refresh_token,json=refreshToken,proto3" json:"refresh_token,omitempty"`
	Username     string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
}

type RefreshResponse struct {
	NewSessionToken string `protobuf:"bytes,1,opt,name=new_session_token,json=newSessionToken,proto3" json:"new_session_token,omitempty"`
	NewRefreshToken string `protobuf:"bytes,2,opt,name=new_refresh_token,json=newRefreshToken,proto3" json:"new_refresh_token,omitempty"`
	ExpiryTime      int64  `protobuf:"varint,3,opt,name=expiry_time,json=expiryTime,proto3" json:"expiry_time,omitempty"`
}

type SignUpRequest struct {
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	Email    string `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	Name     string `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Phone    string `protobuf:"bytes,5,opt,name=phone,proto3" json:"phone,omitempty"`
	Address  string `protobuf:"bytes,6,opt,name=address,proto3" json:"address,omitempty"`
	Sex      string `protobuf:"bytes,7,opt,name=sex,proto3" json:"sex,omitempty"`
}

type SignUpResponse struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	User    *User  `protobuf:"bytes,2,opt,name=user,proto3" json:"user,omitempty"`
}

type User struct {
	Id         int32                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Username   string                 `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Password   string                 `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	Email      string                 `protobuf:"bytes,4,opt,name=email,proto3" json:"email,omitempty"`
	Name       string                 `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	Phone      string                 `protobuf:"bytes,6,opt,name=phone,proto3" json:"phone,omitempty"`
	Address    string                 `protobuf:"bytes,7,opt,name=address,proto3" json:"address,omitempty"`
	Sex        string                 `protobuf:"bytes,8,opt,name=sex,proto3" json:"sex,omitempty"`
	Checked    bool                   `protobuf:"varint,9,opt,name=checked,proto3" json:"checked,omitempty"`
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	UpdateTime *timestamppb.Timestamp `protobuf:"bytes,11,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
	Roles      []*Role                `protobuf:"bytes,12,rep,name=roles,proto3" json:"roles,omitempty"`
}

type Role struct {
	Id    int32   `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Users []*User `protobuf:"bytes,3,rep,name=users,proto3" json:"users,omitempty"`
}

type UpdateProfileRequest struct {
	Id      int32  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name    string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Email   string `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	Phone   string `protobuf:"bytes,4,opt,name=phone,proto3" json:"phone,omitempty"`
	Address string `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
	Sex     string `protobuf:"bytes,6,opt,name=sex,proto3" json:"sex,omitempty"`
}

type UpdateProfileResponse struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	User    *User  `protobuf:"bytes,2,opt,name=user,proto3" json:"user,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `protobuf:"bytes,1,opt,name=refresh_token,json=refreshToken,proto3" json:"refresh_token,omitempty"`
}

type RefreshTokenResponse struct {
	Token        string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	RefreshToken string `protobuf:"bytes,2,opt,name=refresh_token,json=refreshToken,proto3" json:"refresh_token,omitempty"`
	Message      string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

type FindUserRequest struct {
	Username *string `protobuf:"bytes,1,opt,name=username,proto3,oneof" json:"username,omitempty"`
	Email    *string `protobuf:"bytes,2,opt,name=email,proto3,oneof" json:"email,omitempty"`
	Page     *int32  `protobuf:"varint,3,opt,name=page,proto3,oneof" json:"page,omitempty"`
	Limit    *int32  `protobuf:"varint,4,opt,name=limit,proto3,oneof" json:"limit,omitempty"`
	UserId   int32   `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type FindUserResponse struct {
	User *User `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
}

type AddRoleItemRequest struct {
	UserId int32 `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	RoleId int32 `protobuf:"varint,2,opt,name=role_id,json=roleId,proto3" json:"role_id,omitempty"`
}

type AddRoleItemResponse struct {
	Role    *Role  `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

type UpdateRoleItemRequest struct {
	UserId int32 `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	RoleId int32 `protobuf:"varint,2,opt,name=role_id,json=roleId,proto3" json:"role_id,omitempty"`
}

type UpdateRoleItemResponse struct {
	Role    *Role  `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

type RemoveRoleItemRequest struct {
	UserId int32 `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

type RemoveRoleItemResponse struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

type GetRoleRequest struct {
	Id int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

type ListRolesRequest struct {
	Limit int32 `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
	Page  int32 `protobuf:"varint,2,opt,name=page,proto3" json:"page,omitempty"`
}

type ListRolesResponse struct {
	Limit      int32   `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
	Page       int32   `protobuf:"varint,2,opt,name=page,proto3" json:"page,omitempty"`
	Sort       string  `protobuf:"bytes,3,opt,name=sort,proto3" json:"sort,omitempty"`
	TotalRows  int64   `protobuf:"varint,4,opt,name=total_rows,json=totalRows,proto3" json:"total_rows,omitempty"`
	TotalPages int32   `protobuf:"varint,5,opt,name=total_pages,json=totalPages,proto3" json:"total_pages,omitempty"`
	Roles      []*Role `protobuf:"bytes,6,rep,name=roles,proto3" json:"roles,omitempty"`
}

type CreateRoleResquest struct {
	Id   int32  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

type CreateRoleResponse struct {
	Role    *Role  `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

type UpdateRoleRequest struct {
	Id   int32  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

type DeleteRoleItemResponse struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

type UpdateRoleResponse struct {
	Role    *Role  `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

type DeleteRoleRequest struct {
	Id int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

type DeleteRoleResponse struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

type AuthenticationRequest struct {
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Token    string `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
}

type AuthenticationResponse struct {
	Valid   bool   `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

type LockAccountRequest struct {
	Id int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

type LockAccountResponse struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

type GetRoleResponse struct {
	Role    *Role  `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}
