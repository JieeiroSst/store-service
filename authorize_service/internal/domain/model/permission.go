package model

import (
	"strings"
)

// user → operator → admin → super_admin.
const (
	RoleUser       = "user"        // người dùng (end customer)
	RoleOperator   = "operator"    // người vận hành (back-office staff)
	RoleAdmin      = "admin"       // quản trị: users, roles, system config
	RoleSuperAdmin = "super_admin" // toàn quyền; only granted via bootstrap config
)

// CRM roles are functional roles granted on top of a platform role. They
// rank below RoleUser so they never become a user's primary role (other
// services read the primary role); customer-relationship-service reads
// every role from the token's "roles" claim instead.
const (
	RoleCRMViewer  = "crm-viewer"
	RoleCRMStaff   = "crm-staff"
	RoleCRMManager = "crm-manager"
	RoleCRMAdmin   = "crm-admin"
)

const UserSubjectPrefix = "user:"

func UserSubject(userID string) string { return UserSubjectPrefix + userID }

func IsUserSubject(sub string) bool { return strings.HasPrefix(sub, UserSubjectPrefix) }

type Permission struct {
	Object    string `json:"object"`
	Action    string `json:"action"`
	GrantedBy string `json:"granted_by,omitempty"`
}

type RoleDefinition struct {
	Name        string
	Description string
	Rank        int
	Inherits    []string
	Assignable  bool
	Permissions []Permission
}

type UserRoles struct {
	UserID         string
	Roles          []string
	EffectiveRoles []string
	PrimaryRole    string
}

func methods(m ...string) string { return "^(" + strings.Join(m, "|") + ")$" }

var (
	actRead   = methods("GET")
	actCreate = methods("GET", "POST")
	actEdit   = methods("GET", "PUT", "PATCH")
	actWrite  = methods("GET", "POST", "PUT", "PATCH")
	actAll    = methods("GET", "POST", "PUT", "PATCH", "DELETE")
	actAny    = ".*"
)

func grant(act string, objects ...string) []Permission {
	out := make([]Permission, 0, len(objects))
	for _, o := range objects {
		out = append(out, Permission{Object: o, Action: act})
	}
	return out
}

func concat(groups ...[]Permission) []Permission {
	var out []Permission
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

var userPermissions = concat(
	// account & session
	grant(actCreate, "/auth/*", "/otp*"),
	grant(actEdit, "/users/me*"),
	grant(actAll, "/users/me/addresses*"),

	// browse catalog & content (read-only)
	grant(actRead,
		"/products*", "/catalogues*", "/categories*", "/books*", "/foods*",
		"/restaurants*", "/shops*", "/movies*", "/rooms*", "/rent-houses*",
		"/cars*", "/parking-lots*", "/search*", "/filters*", "/countries*",
		"/geo*", "/livestreams*", "/videos*", "/photos*", "/threads*",
		"/lotteries*", "/offers*",
	),

	// shopping
	grant(actAll, "/baskets*", "/wishlists*"),
	grant(actCreate, "/orders*"),
	grant(actRead, "/shipping*"),

	// payments & wallet
	grant(actCreate, "/payments*", "/wallets*", "/billing*"),
	grant(actRead, "/transactions*"),

	// loyalty
	grant(actCreate, "/vouchers*", "/coupons*", "/points*", "/rewards*", "/referrals*", "/bonuslink*"),

	// bookings
	grant(actAll, "/bookings*", "/reservations*", "/tickets*", "/car-rentals*"),

	// social & content the user owns
	grant(actAll, "/posts*", "/reviews*", "/comments*"),
	grant(actCreate, "/messages*", "/chats*", "/uploads*", "/media*", "/shortlinks*", "/qr*"),
	grant(actEdit, "/notifications*"),

	// identity verification
	grant(actCreate, "/ekyc*", "/face*"),

	// medical / patient self-service
	grant(actCreate, "/appointments*", "/medical-records/me*"),
)

var operatorPermissions = concat(
	grant(actRead, "/admin/dashboard*", "/admin/reports*", "/admin/roles*"),

	// customers: view, edit, lock — no delete
	grant(actEdit, "/admin/users*"),
	grant(methods("POST"), "/admin/users/*/lock", "/admin/users/*/unlock"),

	// catalog & inventory
	grant(actAll, "/admin/products*", "/admin/catalogues*", "/admin/categories*",
		"/admin/books*", "/admin/foods*", "/admin/restaurants*", "/admin/movies*",
		"/admin/rooms*", "/admin/rent-houses*", "/admin/cars*", "/admin/parking-lots*",
	),
	grant(actEdit, "/admin/inventory*"),

	// orders, fulfilment, shipping
	grant(actWrite, "/admin/orders*", "/admin/shipping*"),

	// money: read payments, handle refunds
	grant(actRead, "/admin/payments*", "/admin/transactions*", "/admin/wallets*", "/admin/billing*"),
	grant(actWrite, "/admin/refunds*", "/admin/recompenses*"),

	// marketing & loyalty
	grant(actAll, "/admin/vouchers*", "/admin/coupons*", "/admin/promotions*", "/admin/offers*"),
	grant(actWrite, "/admin/points*", "/admin/rewards*", "/admin/referrals*", "/admin/subsidies*"),

	// bookings
	grant(actWrite, "/admin/bookings*", "/admin/reservations*", "/admin/tickets*", "/admin/car-rentals*"),

	// moderation
	grant(actAll, "/admin/posts*", "/admin/reviews*", "/admin/comments*", "/admin/livestreams*"),

	// customer care
	grant(actWrite, "/admin/notifications*", "/admin/support*", "/admin/call-center*", "/admin/chats*"),
	grant(actEdit, "/admin/ekyc*"),

	// partners & shops
	grant(actEdit, "/admin/partners*", "/admin/shops*"),

	// medical back office
	grant(actWrite, "/admin/appointments*", "/admin/patients*", "/admin/medical-records*"),
)

// adminPermissions: user/role administration and system configuration.
var adminPermissions = concat(
	grant(actAll,
		"/admin/users*", "/admin/roles*", "/admin/permissions*",
		"/admin/partners*", "/admin/shops*", "/admin/payments*",
		"/admin/configs*", "/admin/toggles*", "/admin/kms*",
		"/casbin*", // authorize-service raw rule API
	),
	grant(actRead, "/admin/audit-logs*"),
)

var superAdminPermissions = []Permission{{Object: "*", Action: actAny}}

var (
	crmViewerPermissions  = grant(actRead, "/crm*")
	crmStaffPermissions   = grant(actWrite, "/crm*")
	crmManagerPermissions = grant(actAll, "/crm*")
)

var DefaultRoles = []RoleDefinition{
	{
		Name:        RoleUser,
		Description: "Người dùng: mua sắm, đặt chỗ, thanh toán, quản lý tài khoản của chính mình",
		Rank:        10,
		Assignable:  true,
		Permissions: userPermissions,
	},
	{
		Name:        RoleOperator,
		Description: "Người vận hành: quản lý sản phẩm, đơn hàng, vận chuyển, khuyến mãi, kiểm duyệt, CSKH",
		Rank:        20,
		Inherits:    []string{RoleUser},
		Assignable:  true,
		Permissions: operatorPermissions,
	},
	{
		Name:        RoleAdmin,
		Description: "Quản trị: toàn bộ quyền vận hành + quản lý user, role, phân quyền, cấu hình hệ thống",
		Rank:        30,
		Inherits:    []string{RoleOperator},
		Assignable:  true,
		Permissions: adminPermissions,
	},
	{
		Name:        RoleSuperAdmin,
		Description: "Toàn quyền hệ thống; chỉ cấp qua cấu hình bootstrap",
		Rank:        40,
		Inherits:    []string{RoleAdmin},
		Assignable:  false,
		Permissions: superAdminPermissions,
	},
	{
		Name:        RoleCRMViewer,
		Description: "CRM: xem khách hàng, hợp đồng, hồ sơ",
		Rank:        1,
		Assignable:  true,
		Permissions: crmViewerPermissions,
	},
	{
		Name:        RoleCRMStaff,
		Description: "CRM: xem + tải lên hồ sơ, xoá hồ sơ của chính mình",
		Rank:        2,
		Inherits:    []string{RoleCRMViewer},
		Assignable:  true,
		Permissions: crmStaffPermissions,
	},
	{
		Name:        RoleCRMManager,
		Description: "CRM: quản lý mọi hồ sơ, hồ sơ pháp lý, nhật ký kiểm toán",
		Rank:        3,
		Inherits:    []string{RoleCRMStaff},
		Assignable:  true,
		Permissions: crmManagerPermissions,
	},
	{
		Name:        RoleCRMAdmin,
		Description: "CRM: toàn quyền trong CRM",
		Rank:        4,
		Inherits:    []string{RoleCRMManager},
		Assignable:  true,
	},
}

func FindRoleDefinition(name string) (RoleDefinition, bool) {
	for _, r := range DefaultRoles {
		if r.Name == name {
			return r, true
		}
	}
	return RoleDefinition{}, false
}

func RoleRank(name string) int {
	if r, ok := FindRoleDefinition(name); ok {
		return r.Rank
	}
	return 0
}
