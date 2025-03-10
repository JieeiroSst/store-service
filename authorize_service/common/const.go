package common

const (
	InternalServer = "INTERVAL SERVER"
	Unauthorized   = "Unauthorized"
	NotAllowServer = "THE CUSTOMER IS NOT AUTHORIZED FOR THE CONTENT REQUESTED"
	BadRequest     = "BAD REQUEST"
	NotFoundServer = "Data Entity"
	FailedDBServer = "Cant get data from database"
)

const (
	Authorized    = "THE CUSTOMER IS AUTHORIZED FOR THE CONTENT REQUESTED"
	CreateSuccess = "CREATE SUCCESS"
	UpdateSuccess = "UPDATE SUCCESS"

)

const (
	RBAC_MODEL = "config/conf/rbac_model.conf"
)

const (

)

const (
	ListCasbinKeyCache = "as:list_casbin"
	CasbinByIDKeyCache = "as:casbin_by_id:%v"
)