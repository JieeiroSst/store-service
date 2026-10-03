package manage_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "manage_service"

const DefaultBaseURL = "http://manage-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) TokenUser(ctx context.Context, realm *string) (*model.KcTokenInfo, error) {
	path := "/token-user"
	q := url.Values{}
	if realm != nil {
		q.Set("realm", *realm)
	}
	h := http.Header{}
	var out *model.KcTokenInfo
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Issuer(ctx context.Context, realm string) (*model.KcIssuerResponse, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/issuer"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcIssuerResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Certs(ctx context.Context, realm string) (*model.KcCertResponse, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/protocol/certs"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcCertResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RawUserInfo(ctx context.Context, realm string) (map[string]interface{}, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/protocol/userinfo"
	q := url.Values{}
	h := http.Header{}
	var out map[string]interface{}
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) DecodeAccessToken(ctx context.Context, realm string) (map[string]interface{}, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/protocol/decode"
	q := url.Values{}
	h := http.Header{}
	var out map[string]interface{}
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserPermissions(ctx context.Context, realm string) ([]*model.KcPermissionGrantResponseRepresentation, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/authz/uma-permissions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcPermissionGrantResponseRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ResourcesClient(ctx context.Context, realm string, deep *string, first *string, max *string, name *string, owner *string, typeArg *string, uri *string, scope *string, matchingURI *string, exactName *string) ([]*model.KcResourceRepresentation, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/authz/resources"
	q := url.Values{}
	if deep != nil {
		q.Set("deep", *deep)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if name != nil {
		q.Set("name", *name)
	}
	if owner != nil {
		q.Set("owner", *owner)
	}
	if typeArg != nil {
		q.Set("type", *typeArg)
	}
	if uri != nil {
		q.Set("uri", *uri)
	}
	if scope != nil {
		q.Set("scope", *scope)
	}
	if matchingURI != nil {
		q.Set("matchingUri", *matchingURI)
	}
	if exactName != nil {
		q.Set("exactName", *exactName)
	}
	h := http.Header{}
	var out []*model.KcResourceRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ResourceClient(ctx context.Context, realm string, resourceID string) (*model.KcResourceRepresentation, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/authz/resources/" + url.PathEscape(resourceID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcResourceRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ResourcePolicies(ctx context.Context, realm string, resource *string, name *string, scope *string, first *string, max *string) ([]*model.KcResourcePolicyRepresentation, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/authz/resource-policies"
	q := url.Values{}
	if resource != nil {
		q.Set("resource", *resource)
	}
	if name != nil {
		q.Set("name", *name)
	}
	if scope != nil {
		q.Set("scope", *scope)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	h := http.Header{}
	var out []*model.KcResourcePolicyRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ResourcePolicy(ctx context.Context, realm string, permissionID string) (*model.KcResourcePolicyRepresentation, error) {
	path := "/api/v1/realms/" + url.PathEscape(realm) + "/authz/resource-policies/" + url.PathEscape(permissionID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcResourcePolicyRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ServerInfo(ctx context.Context) (*model.KcServerInfoRepresentation, error) {
	path := "/api/v1/admin/server-info"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcServerInfoRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Realms(ctx context.Context) ([]*model.KcRealmRepresentation, error) {
	path := "/api/v1/admin/realms"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcRealmRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Realm(ctx context.Context, realm string) (*model.KcRealmRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcRealmRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) KeyStoreConfig(ctx context.Context, realm string) (*model.KcKeyStoreConfig, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/keys"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcKeyStoreConfig
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Events(ctx context.Context, realm string, client *string, dateFrom *string, dateTo *string, first *string, ipAddress *string, max *string, user *string) ([]*model.KcEventRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/events"
	q := url.Values{}
	if client != nil {
		q.Set("client", *client)
	}
	if dateFrom != nil {
		q.Set("dateFrom", *dateFrom)
	}
	if dateTo != nil {
		q.Set("dateTo", *dateTo)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if ipAddress != nil {
		q.Set("ipAddress", *ipAddress)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if user != nil {
		q.Set("user", *user)
	}
	h := http.Header{}
	var out []*model.KcEventRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CredentialRegistrators(ctx context.Context, realm string) ([]string, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/credential-registrators"
	q := url.Values{}
	h := http.Header{}
	var out []string
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Users(ctx context.Context, realm string, briefRepresentation *string, email *string, emailVerified *string, enabled *string, exact *string, first *string, firstName *string, idpAlias *string, idpUserID *string, lastName *string, max *string, qArg *string, search *string, username *string) ([]*model.KcUser, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users"
	q := url.Values{}
	if briefRepresentation != nil {
		q.Set("briefRepresentation", *briefRepresentation)
	}
	if email != nil {
		q.Set("email", *email)
	}
	if emailVerified != nil {
		q.Set("emailVerified", *emailVerified)
	}
	if enabled != nil {
		q.Set("enabled", *enabled)
	}
	if exact != nil {
		q.Set("exact", *exact)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if firstName != nil {
		q.Set("firstName", *firstName)
	}
	if idpAlias != nil {
		q.Set("idpAlias", *idpAlias)
	}
	if idpUserID != nil {
		q.Set("idpUserId", *idpUserID)
	}
	if lastName != nil {
		q.Set("lastName", *lastName)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if search != nil {
		q.Set("search", *search)
	}
	if username != nil {
		q.Set("username", *username)
	}
	h := http.Header{}
	var out []*model.KcUser
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserCount(ctx context.Context, realm string, briefRepresentation *string, email *string, emailVerified *string, enabled *string, exact *string, first *string, firstName *string, idpAlias *string, idpUserID *string, lastName *string, max *string, qArg *string, search *string, username *string) (*int, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/count"
	q := url.Values{}
	if briefRepresentation != nil {
		q.Set("briefRepresentation", *briefRepresentation)
	}
	if email != nil {
		q.Set("email", *email)
	}
	if emailVerified != nil {
		q.Set("emailVerified", *emailVerified)
	}
	if enabled != nil {
		q.Set("enabled", *enabled)
	}
	if exact != nil {
		q.Set("exact", *exact)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if firstName != nil {
		q.Set("firstName", *firstName)
	}
	if idpAlias != nil {
		q.Set("idpAlias", *idpAlias)
	}
	if idpUserID != nil {
		q.Set("idpUserId", *idpUserID)
	}
	if lastName != nil {
		q.Set("lastName", *lastName)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if search != nil {
		q.Set("search", *search)
	}
	if username != nil {
		q.Set("username", *username)
	}
	h := http.Header{}
	var out *int
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserByID(ctx context.Context, realm string, userID string) (*model.KcUser, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcUser
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserBruteForceDetectionStatus(ctx context.Context, realm string, userID string) (*model.KcBruteForceStatus, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID) + "/brute-force"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcBruteForceStatus
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserSessions(ctx context.Context, realm string, userID string) ([]*model.KcUserSessionRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID) + "/sessions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcUserSessionRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserOfflineSessionsForClient(ctx context.Context, realm string, userID string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID) + "/offline-sessions/" + url.PathEscape(clientID)
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcUserSessionRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserGroups(ctx context.Context, realm string, userID string, briefRepresentation *string, exact *string, first *string, full *string, max *string, qArg *string, search *string) ([]*model.KcGroup, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID) + "/groups"
	q := url.Values{}
	if briefRepresentation != nil {
		q.Set("briefRepresentation", *briefRepresentation)
	}
	if exact != nil {
		q.Set("exact", *exact)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if full != nil {
		q.Set("full", *full)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if search != nil {
		q.Set("search", *search)
	}
	h := http.Header{}
	var out []*model.KcGroup
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Credentials(ctx context.Context, realm string, userID string) ([]*model.KcCredentialRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID) + "/credentials"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcCredentialRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ConfiguredUserStorageCredentialTypes(ctx context.Context, realm string, userID string) ([]string, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID) + "/configured-user-storage-credential-types"
	q := url.Values{}
	h := http.Header{}
	var out []string
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserFederatedIdentities(ctx context.Context, realm string, userID string) ([]*model.KcFederatedIdentityRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/users/" + url.PathEscape(userID) + "/federated-identity"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcFederatedIdentityRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Groups(ctx context.Context, realm string, briefRepresentation *string, exact *string, first *string, full *string, max *string, qArg *string, search *string) ([]*model.KcGroup, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/groups"
	q := url.Values{}
	if briefRepresentation != nil {
		q.Set("briefRepresentation", *briefRepresentation)
	}
	if exact != nil {
		q.Set("exact", *exact)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if full != nil {
		q.Set("full", *full)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if search != nil {
		q.Set("search", *search)
	}
	h := http.Header{}
	var out []*model.KcGroup
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GroupsCount(ctx context.Context, realm string, briefRepresentation *string, exact *string, first *string, full *string, max *string, qArg *string, search *string) (*int, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/groups/count"
	q := url.Values{}
	if briefRepresentation != nil {
		q.Set("briefRepresentation", *briefRepresentation)
	}
	if exact != nil {
		q.Set("exact", *exact)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if full != nil {
		q.Set("full", *full)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if search != nil {
		q.Set("search", *search)
	}
	h := http.Header{}
	var out *int
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Group(ctx context.Context, realm string, groupID string) (*model.KcGroup, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/groups/" + url.PathEscape(groupID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcGroup
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GroupMembers(ctx context.Context, realm string, groupID string, briefRepresentation *string, exact *string, first *string, full *string, max *string, qArg *string, search *string) ([]*model.KcUser, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/groups/" + url.PathEscape(groupID) + "/members"
	q := url.Values{}
	if briefRepresentation != nil {
		q.Set("briefRepresentation", *briefRepresentation)
	}
	if exact != nil {
		q.Set("exact", *exact)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if full != nil {
		q.Set("full", *full)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if search != nil {
		q.Set("search", *search)
	}
	h := http.Header{}
	var out []*model.KcUser
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GroupByPath(ctx context.Context, realm string, pathArg string) (*model.KcGroup, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/group-by-path/" + rest.EscapePath(pathArg)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcGroup
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) DefaultGroups(ctx context.Context, realm string) ([]*model.KcGroup, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/default-groups"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcGroup
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RealmRoles(ctx context.Context, realm string, first *string, max *string, search *string, briefRepresentation *string) ([]*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles"
	q := url.Values{}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if search != nil {
		q.Set("search", *search)
	}
	if briefRepresentation != nil {
		q.Set("briefRepresentation", *briefRepresentation)
	}
	h := http.Header{}
	var out []*model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RealmRole(ctx context.Context, realm string, roleName string) (*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles/" + url.PathEscape(roleName)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CompositeRealmRoles(ctx context.Context, realm string, roleName string) ([]*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles/" + url.PathEscape(roleName) + "/composites"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UsersByRoleName(ctx context.Context, realm string, roleName string, first *string, max *string) ([]*model.KcUser, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles/" + url.PathEscape(roleName) + "/users"
	q := url.Values{}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	h := http.Header{}
	var out []*model.KcUser
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GroupsByRole(ctx context.Context, realm string, roleName string) ([]*model.KcGroup, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles/" + url.PathEscape(roleName) + "/groups"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcGroup
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RealmRoleByID(ctx context.Context, realm string, roleID string) (*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles-by-id/" + url.PathEscape(roleID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CompositeRolesByRoleID(ctx context.Context, realm string, roleID string) ([]*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles-by-id/" + url.PathEscape(roleID) + "/composites"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CompositeRealmRolesByRoleID(ctx context.Context, realm string, roleID string) ([]*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles-by-id/" + url.PathEscape(roleID) + "/composites/realm"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CompositeClientRolesByRoleID(ctx context.Context, realm string, roleID string, clientID string) ([]*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/roles-by-id/" + url.PathEscape(roleID) + "/composites/clients/" + url.PathEscape(clientID)
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientsWithParams(ctx context.Context, realm string, clientID *string, viewableOnly *string, first *string, max *string, search *string, qArg *string) ([]*model.KcClient, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients"
	q := url.Values{}
	if clientID != nil {
		q.Set("clientId", *clientID)
	}
	if viewableOnly != nil {
		q.Set("viewableOnly", *viewableOnly)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if search != nil {
		q.Set("search", *search)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	h := http.Header{}
	var out []*model.KcClient
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Client(ctx context.Context, realm string, clientID string) (*model.KcClient, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcClient
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientSecret(ctx context.Context, realm string, clientID string) (*model.KcCredentialRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/client-secret"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcCredentialRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientServiceAccount(ctx context.Context, realm string, clientID string) (*model.KcUser, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/service-account-user"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcUser
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientUserSessions(ctx context.Context, realm string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/user-sessions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcUserSessionRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientOfflineSessions(ctx context.Context, realm string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/offline-sessions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcUserSessionRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientsDefaultScopes(ctx context.Context, realm string, clientID string) ([]*model.KcClientScope, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/default-client-scopes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcClientScope
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientsOptionalScopes(ctx context.Context, realm string, clientID string) ([]*model.KcClientScope, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/optional-client-scopes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcClientScope
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientScopeMappings(ctx context.Context, realm string, clientID string) (*model.KcMappingsRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/scope-mappings"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcMappingsRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientRoles(ctx context.Context, realm string, clientID string, first *string, max *string, search *string, briefRepresentation *string) ([]*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/roles"
	q := url.Values{}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if search != nil {
		q.Set("search", *search)
	}
	if briefRepresentation != nil {
		q.Set("briefRepresentation", *briefRepresentation)
	}
	h := http.Header{}
	var out []*model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientRole(ctx context.Context, realm string, clientID string, roleName string) (*model.KcRole, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/roles/" + url.PathEscape(roleName)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcRole
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UsersByClientRoleName(ctx context.Context, realm string, clientID string, roleName string, first *string, max *string) ([]*model.KcUser, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/roles/" + url.PathEscape(roleName) + "/users"
	q := url.Values{}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	h := http.Header{}
	var out []*model.KcUser
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GroupsByClientRole(ctx context.Context, realm string, clientID string, roleName string) ([]*model.KcGroup, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/roles/" + url.PathEscape(roleName) + "/groups"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcGroup
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ResourceServer(ctx context.Context, realm string, clientID string) (*model.KcResourceServerRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcResourceServerRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Resources(ctx context.Context, realm string, clientID string, deep *string, first *string, max *string, name *string, owner *string, typeArg *string, uri *string, scope *string, matchingURI *string, exactName *string) ([]*model.KcResourceRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/resource"
	q := url.Values{}
	if deep != nil {
		q.Set("deep", *deep)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if name != nil {
		q.Set("name", *name)
	}
	if owner != nil {
		q.Set("owner", *owner)
	}
	if typeArg != nil {
		q.Set("type", *typeArg)
	}
	if uri != nil {
		q.Set("uri", *uri)
	}
	if scope != nil {
		q.Set("scope", *scope)
	}
	if matchingURI != nil {
		q.Set("matchingUri", *matchingURI)
	}
	if exactName != nil {
		q.Set("exactName", *exactName)
	}
	h := http.Header{}
	var out []*model.KcResourceRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Resource(ctx context.Context, realm string, clientID string, resourceID string) (*model.KcResourceRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/resource/" + url.PathEscape(resourceID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcResourceRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Scopes(ctx context.Context, realm string, clientID string, deep *string, first *string, max *string, name *string) ([]*model.KcScopeRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/scope"
	q := url.Values{}
	if deep != nil {
		q.Set("deep", *deep)
	}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if name != nil {
		q.Set("name", *name)
	}
	h := http.Header{}
	var out []*model.KcScopeRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Scope(ctx context.Context, realm string, clientID string, scopeID string) (*model.KcScopeRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/scope/" + url.PathEscape(scopeID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcScopeRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Policies(ctx context.Context, realm string, clientID string, first *string, max *string, name *string, permission *string, typeArg *string) ([]*model.KcPolicyRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/policy"
	q := url.Values{}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if name != nil {
		q.Set("name", *name)
	}
	if permission != nil {
		q.Set("permission", *permission)
	}
	if typeArg != nil {
		q.Set("type", *typeArg)
	}
	h := http.Header{}
	var out []*model.KcPolicyRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Policy(ctx context.Context, realm string, clientID string, policyID string) (*model.KcPolicyRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/policy/" + url.PathEscape(policyID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcPolicyRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuthorizationPolicyAssociatedPolicies(ctx context.Context, realm string, clientID string, policyID string) ([]*model.KcPolicyRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/policy/" + url.PathEscape(policyID) + "/associatedPolicies"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcPolicyRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuthorizationPolicyResources(ctx context.Context, realm string, clientID string, policyID string) ([]*model.KcPolicyResourceRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/policy/" + url.PathEscape(policyID) + "/resources"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcPolicyResourceRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuthorizationPolicyScopes(ctx context.Context, realm string, clientID string, policyID string) ([]*model.KcPolicyScopeRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/policy/" + url.PathEscape(policyID) + "/scopes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcPolicyScopeRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) DependentPermissions(ctx context.Context, realm string, clientID string, policyID string) ([]*model.KcPermissionRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/policy/" + url.PathEscape(policyID) + "/dependentPolicies"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcPermissionRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Permissions(ctx context.Context, realm string, clientID string, first *string, max *string, name *string, resource *string, scope *string, typeArg *string) ([]*model.KcPermissionRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/permission"
	q := url.Values{}
	if first != nil {
		q.Set("first", *first)
	}
	if max != nil {
		q.Set("max", *max)
	}
	if name != nil {
		q.Set("name", *name)
	}
	if resource != nil {
		q.Set("resource", *resource)
	}
	if scope != nil {
		q.Set("scope", *scope)
	}
	if typeArg != nil {
		q.Set("type", *typeArg)
	}
	h := http.Header{}
	var out []*model.KcPermissionRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Permission(ctx context.Context, realm string, clientID string, permissionID string) (*model.KcPermissionRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/permission/" + url.PathEscape(permissionID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcPermissionRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PermissionResources(ctx context.Context, realm string, clientID string, permissionID string) ([]*model.KcPermissionResource, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/permission/" + url.PathEscape(permissionID) + "/resources"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcPermissionResource
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) PermissionScopes(ctx context.Context, realm string, clientID string, permissionID string) ([]*model.KcPermissionScope, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients/" + url.PathEscape(clientID) + "/authz/resource-server/permission/" + url.PathEscape(permissionID) + "/scopes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcPermissionScope
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientRepresentation(ctx context.Context, realm string, clientID string) (*model.KcClient, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients-registrations/" + url.PathEscape(clientID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcClient
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AdapterConfiguration(ctx context.Context, realm string, clientID string) (*model.KcAdapterConfiguration, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/clients-registrations/" + url.PathEscape(clientID) + "/installation"
	q := url.Values{}
	h := http.Header{}
	var out *model.KcAdapterConfiguration
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientScopes(ctx context.Context, realm string) ([]*model.KcClientScope, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/client-scopes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcClientScope
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientScope(ctx context.Context, realm string, scopeID string) (*model.KcClientScope, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/client-scopes/" + url.PathEscape(scopeID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcClientScope
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientScopeProtocolMappers(ctx context.Context, realm string, scopeID string) ([]*model.KcProtocolMappers, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/client-scopes/" + url.PathEscape(scopeID) + "/protocol-mappers"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcProtocolMappers
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ClientScopeProtocolMapper(ctx context.Context, realm string, scopeID string, mapperID string) (*model.KcProtocolMappers, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/client-scopes/" + url.PathEscape(scopeID) + "/protocol-mappers/" + url.PathEscape(mapperID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcProtocolMappers
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) DefaultDefaultClientScopes(ctx context.Context, realm string) ([]*model.KcClientScope, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/default-default-client-scopes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcClientScope
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) DefaultOptionalClientScopes(ctx context.Context, realm string) ([]*model.KcClientScope, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/default-optional-client-scopes"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcClientScope
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ComponentsWithParams(ctx context.Context, realm string, name *string, provider *string, parent *string) ([]*model.KcComponent, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/components"
	q := url.Values{}
	if name != nil {
		q.Set("name", *name)
	}
	if provider != nil {
		q.Set("provider", *provider)
	}
	if parent != nil {
		q.Set("parent", *parent)
	}
	h := http.Header{}
	var out []*model.KcComponent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Component(ctx context.Context, realm string, componentID string) (*model.KcComponent, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/components/" + url.PathEscape(componentID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcComponent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuthenticationFlows(ctx context.Context, realm string) ([]*model.KcAuthenticationFlowRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/authentication/flows"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcAuthenticationFlowRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuthenticationFlow(ctx context.Context, realm string, flow string) (*model.KcAuthenticationFlowRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/authentication/flows/" + url.PathEscape(flow)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcAuthenticationFlowRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) AuthenticationExecutions(ctx context.Context, realm string, flow string) ([]*model.KcModifyAuthenticationExecutionRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/authentication/flows/" + url.PathEscape(flow) + "/executions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcModifyAuthenticationExecutionRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RequiredActions(ctx context.Context, realm string) ([]*model.KcRequiredActionProviderRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/authentication/required-actions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcRequiredActionProviderRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RequiredAction(ctx context.Context, realm string, alias string) (*model.KcRequiredActionProviderRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/authentication/required-actions/" + url.PathEscape(alias)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcRequiredActionProviderRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) IdentityProviders(ctx context.Context, realm string) ([]*model.KcIdentityProviderRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/identity-providers"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcIdentityProviderRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) IdentityProvider(ctx context.Context, realm string, alias string) (*model.KcIdentityProviderRepresentation, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/identity-providers/" + url.PathEscape(alias)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcIdentityProviderRepresentation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ExportIDPPublicBrokerConfig(ctx context.Context, realm string, alias string) (*string, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/identity-providers/" + url.PathEscape(alias) + "/export"
	q := url.Values{}
	h := http.Header{}
	var out *string
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) IdentityProviderMappers(ctx context.Context, realm string, alias string) ([]*model.KcIdentityProviderMapper, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/identity-providers/" + url.PathEscape(alias) + "/mappers"
	q := url.Values{}
	h := http.Header{}
	var out []*model.KcIdentityProviderMapper
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) IdentityProviderMapperByID(ctx context.Context, realm string, alias string, mapperID string) (*model.KcIdentityProviderMapper, error) {
	path := "/api/v1/admin/realms/" + url.PathEscape(realm) + "/identity-providers/" + url.PathEscape(alias) + "/mappers/" + url.PathEscape(mapperID)
	q := url.Values{}
	h := http.Header{}
	var out *model.KcIdentityProviderMapper
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
