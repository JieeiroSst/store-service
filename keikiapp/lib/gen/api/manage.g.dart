// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ManageApi {
  const ManageApi(this._executor);

  final GatewayExecutor _executor;

  Future<KcTokenInfo?> kcTokenUser({String? realm, String? selectionSet}) async {
    final data = await _executor.query('query kcTokenUser(\$realm: String) { kcTokenUser(realm: \$realm) ${selectionSet ?? '{ AccessToken RefreshToken TokenType ExpiresIn RefreshExpiresIn Scope }'} }', variables: {
      'realm': ?realm,
    }
    );
    return data['kcTokenUser'] == null ? null : KcTokenInfo.fromJson((data['kcTokenUser'] as Map).cast<String, dynamic>());
  }

  Future<KcIssuerResponse?> kcIssuer({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcIssuer(\$realm: String!) { kcIssuer(realm: \$realm) ${selectionSet ?? '{ realm public_key token_service account_service tokens_not_before }'} }', variables: {
      'realm': realm,
    }
    );
    return data['kcIssuer'] == null ? null : KcIssuerResponse.fromJson((data['kcIssuer'] as Map).cast<String, dynamic>());
  }

  Future<KcCertResponse?> kcCerts({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcCerts(\$realm: String!) { kcCerts(realm: \$realm) ${selectionSet ?? '{ keys { kid kty alg use n e x y crv key_ops x5u x5c x5t x5t_S256 } }'} }', variables: {
      'realm': realm,
    }
    );
    return data['kcCerts'] == null ? null : KcCertResponse.fromJson((data['kcCerts'] as Map).cast<String, dynamic>());
  }

  Future<Map<String, dynamic>?> kcRawUserInfo({required String realm}) async {
    final data = await _executor.query('query kcRawUserInfo(\$realm: String!) { kcRawUserInfo(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return (data['kcRawUserInfo'] as Map?)?.cast<String, dynamic>();
  }

  Future<Map<String, dynamic>?> kcDecodeAccessToken({required String realm}) async {
    final data = await _executor.query('query kcDecodeAccessToken(\$realm: String!) { kcDecodeAccessToken(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return (data['kcDecodeAccessToken'] as Map?)?.cast<String, dynamic>();
  }

  Future<List<KcPermissionGrantResponseRepresentation>?> kcUserPermissions({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcUserPermissions(\$realm: String!) { kcUserPermissions(realm: \$realm) ${selectionSet ?? '{ id owner resource scope granted requester }'} }', variables: {
      'realm': realm,
    }
    );
    return (data['kcUserPermissions'] as List?)?.map((e) => (e == null ? null : KcPermissionGrantResponseRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcResourceRepresentation>?> kcResourcesClient({required String realm, String? deep, String? first, String? max, String? name_, String? owner, String? type_, String? uri, String? scope, String? matchingUri, String? exactName, String? selectionSet}) async {
    final data = await _executor.query('query kcResourcesClient(\$realm: String!, \$deep: String, \$first: String, \$max: String, \$name: String, \$owner: String, \$type: String, \$uri: String, \$scope: String, \$matchingUri: String, \$exactName: String) { kcResourcesClient(realm: \$realm, deep: \$deep, first: \$first, max: \$max, name: \$name, owner: \$owner, type: \$type, uri: \$uri, scope: \$scope, matchingUri: \$matchingUri, exactName: \$exactName) ${selectionSet ?? '{ _id attributes displayName icon_uri name owner { id name } ownerManagedAccess resource_scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } } scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } } type uris }'} }', variables: {
      'realm': realm,
      'deep': ?deep,
      'first': ?first,
      'max': ?max,
      'name': ?name_,
      'owner': ?owner,
      'type': ?type_,
      'uri': ?uri,
      'scope': ?scope,
      'matchingUri': ?matchingUri,
      'exactName': ?exactName,
    }
    );
    return (data['kcResourcesClient'] as List?)?.map((e) => (e == null ? null : KcResourceRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcResourceRepresentation?> kcResourceClient({required String realm, required String resourceID, String? selectionSet}) async {
    final data = await _executor.query('query kcResourceClient(\$realm: String!, \$resourceID: String!) { kcResourceClient(realm: \$realm, resourceID: \$resourceID) ${selectionSet ?? '{ _id attributes displayName icon_uri name owner { id name } ownerManagedAccess resource_scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } } scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } } type uris }'} }', variables: {
      'realm': realm,
      'resourceID': resourceID,
    }
    );
    return data['kcResourceClient'] == null ? null : KcResourceRepresentation.fromJson((data['kcResourceClient'] as Map).cast<String, dynamic>());
  }

  Future<List<KcResourcePolicyRepresentation>?> kcResourcePolicies({required String realm, String? resource, String? name_, String? scope, String? first, String? max, String? selectionSet}) async {
    final data = await _executor.query('query kcResourcePolicies(\$realm: String!, \$resource: String, \$name: String, \$scope: String, \$first: String, \$max: String) { kcResourcePolicies(realm: \$realm, resource: \$resource, name: \$name, scope: \$scope, first: \$first, max: \$max) ${selectionSet ?? '{ name description scopes roles groups clients id logic decisionStrategy owner type users }'} }', variables: {
      'realm': realm,
      'resource': ?resource,
      'name': ?name_,
      'scope': ?scope,
      'first': ?first,
      'max': ?max,
    }
    );
    return (data['kcResourcePolicies'] as List?)?.map((e) => (e == null ? null : KcResourcePolicyRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcResourcePolicyRepresentation?> kcResourcePolicy({required String realm, required String permissionID, String? selectionSet}) async {
    final data = await _executor.query('query kcResourcePolicy(\$realm: String!, \$permissionID: String!) { kcResourcePolicy(realm: \$realm, permissionID: \$permissionID) ${selectionSet ?? '{ name description scopes roles groups clients id logic decisionStrategy owner type users }'} }', variables: {
      'realm': realm,
      'permissionID': permissionID,
    }
    );
    return data['kcResourcePolicy'] == null ? null : KcResourcePolicyRepresentation.fromJson((data['kcResourcePolicy'] as Map).cast<String, dynamic>());
  }

  Future<KcServerInfoRepresentation?> kcServerInfo({String? selectionSet}) async {
    final data = await _executor.query('query kcServerInfo { kcServerInfo ${selectionSet ?? '{ systemInfo { fileEncoding javaHome javaRuntime javaVendor javaVersion javaVm javaVmVersion osArchitecture osName osVersion serverTime uptime uptimeMillis userDir userLocale userName userTimezone version } memoryInfo { free freeFormated freePercentage total totalFormated used usedFormated } passwordPolicies { configType defaultValue displayName id multipleSupported } protocolMapperTypes { docker_v2 { id name category helpText priority } saml { id name category helpText priority } openid_connect { id name category helpText priority } } builtinProtocolMappers { saml { config id name protocol protocolMapper consentRequired } openid_connect { config id name protocol protocolMapper consentRequired } } themes { account { name locales } admin { name locales } common { name locales } email { name locales } login { name locales } welcome { name locales } } }'} }'
    );
    return data['kcServerInfo'] == null ? null : KcServerInfoRepresentation.fromJson((data['kcServerInfo'] as Map).cast<String, dynamic>());
  }

  Future<List<KcRealmRepresentation>?> kcRealms({String? selectionSet}) async {
    final data = await _executor.query('query kcRealms { kcRealms ${selectionSet ?? '{ accessCodeLifespan accessCodeLifespanLogin accessCodeLifespanUserAction accessTokenLifespan accessTokenLifespanForImplicitFlow accountTheme actionTokenGeneratedByAdminLifespan actionTokenGeneratedByUserLifespan adminEventsDetailsEnabled adminEventsEnabled adminTheme attributes authenticationFlows authenticatorConfig browserFlow browserSecurityHeaders bruteForceProtected clientAuthenticationFlow clientScopeMappings clientScopes { id name description protocol attributes { consent_screen_text display_on_consent_screen include_in_token_scope } protocolMappers { id name protocol protocolMapper consentRequired } } clients { access adminUrl attributes authenticationFlowBindingOverrides authorizationServicesEnabled authorizationSettings { allowRemoteResourceManagement clientId id name policyEnforcementMode decisionStrategy } baseUrl bearerOnly clientAuthenticatorType clientId consentRequired defaultClientScopes defaultRoles description directAccessGrantsEnabled enabled frontchannelLogout fullScopeAllowed id implicitFlowEnabled name nodeReRegistrationTimeout notBefore optionalClientScopes origin protocol protocolMappers { config id name protocol protocolMapper consentRequired } publicClient redirectUris registeredNodes registrationAccessToken rootUrl secret serviceAccountsEnabled standardFlowEnabled surrogateAuthRequired webOrigins } components defaultDefaultClientScopes defaultGroups defaultLocale defaultOptionalClientScopes defaultRole { id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes } defaultRoles defaultSignatureAlgorithm directGrantFlow displayName displayNameHtml dockerAuthenticationFlow duplicateEmailsAllowed editUsernameAllowed emailTheme enabled enabledEventTypes eventsEnabled eventsExpiration eventsListeners failureFactor federatedUsers groups id identityProviderMappers identityProviders internationalizationEnabled keycloakVersion loginTheme loginWithEmailAllowed maxDeltaTimeSeconds maxFailureWaitSeconds minimumQuickLoginWaitSeconds notBefore offlineSessionIdleTimeout offlineSessionMaxLifespan offlineSessionMaxLifespanEnabled otpPolicyAlgorithm otpPolicyDigits otpPolicyInitialCounter otpPolicyLookAheadWindow otpPolicyPeriod otpPolicyType otpSupportedApplications passwordPolicy permanentLockout protocolMappers quickLoginCheckMilliSeconds realm refreshTokenMaxReuse registrationAllowed registrationEmailAsUsername registrationFlow rememberMe requiredActions resetCredentialsFlow resetPasswordAllowed revokeRefreshToken roles { client realm { id name scopeParamRequired composite clientRole containerId description attributes } } scopeMappings smtpServer sslRequired ssoSessionIdleTimeout ssoSessionIdleTimeoutRememberMe ssoSessionMaxLifespan ssoSessionMaxLifespanRememberMe supportedLocales userFederationMappers userFederationProviders userManagedAccessAllowed users { id createdTimestamp username enabled totp emailVerified firstName lastName email federationLink attributes disableableCredentialTypes requiredActions access clientRoles realmRoles groups serviceAccountClientId credentials { createdDate temporary type value algorithm counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel } } verifyEmail waitIncrementSeconds webAuthnPolicyAcceptableAaguids webAuthnPolicyAttestationConveyancePreference webAuthnPolicyAuthenticatorAttachment webAuthnPolicyAvoidSameAuthenticatorRegister webAuthnPolicyCreateTimeout webAuthnPolicyPasswordlessAcceptableAaguids webAuthnPolicyPasswordlessAttestationConveyancePreference webAuthnPolicyPasswordlessAuthenticatorAttachment webAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister webAuthnPolicyPasswordlessCreateTimeout webAuthnPolicyPasswordlessRequireResidentKey webAuthnPolicyPasswordlessRpEntityName webAuthnPolicyPasswordlessRpId webAuthnPolicyPasswordlessSignatureAlgorithms webAuthnPolicyPasswordlessUserVerificationRequirement webAuthnPolicyRequireResidentKey webAuthnPolicyRpEntityName webAuthnPolicyRpId webAuthnPolicySignatureAlgorithms webAuthnPolicyUserVerificationRequirement }'} }'
    );
    return (data['kcRealms'] as List?)?.map((e) => (e == null ? null : KcRealmRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcRealmRepresentation?> kcRealm({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcRealm(\$realm: String!) { kcRealm(realm: \$realm) ${selectionSet ?? '{ accessCodeLifespan accessCodeLifespanLogin accessCodeLifespanUserAction accessTokenLifespan accessTokenLifespanForImplicitFlow accountTheme actionTokenGeneratedByAdminLifespan actionTokenGeneratedByUserLifespan adminEventsDetailsEnabled adminEventsEnabled adminTheme attributes authenticationFlows authenticatorConfig browserFlow browserSecurityHeaders bruteForceProtected clientAuthenticationFlow clientScopeMappings clientScopes { id name description protocol attributes { consent_screen_text display_on_consent_screen include_in_token_scope } protocolMappers { id name protocol protocolMapper consentRequired } } clients { access adminUrl attributes authenticationFlowBindingOverrides authorizationServicesEnabled authorizationSettings { allowRemoteResourceManagement clientId id name policyEnforcementMode decisionStrategy } baseUrl bearerOnly clientAuthenticatorType clientId consentRequired defaultClientScopes defaultRoles description directAccessGrantsEnabled enabled frontchannelLogout fullScopeAllowed id implicitFlowEnabled name nodeReRegistrationTimeout notBefore optionalClientScopes origin protocol protocolMappers { config id name protocol protocolMapper consentRequired } publicClient redirectUris registeredNodes registrationAccessToken rootUrl secret serviceAccountsEnabled standardFlowEnabled surrogateAuthRequired webOrigins } components defaultDefaultClientScopes defaultGroups defaultLocale defaultOptionalClientScopes defaultRole { id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes } defaultRoles defaultSignatureAlgorithm directGrantFlow displayName displayNameHtml dockerAuthenticationFlow duplicateEmailsAllowed editUsernameAllowed emailTheme enabled enabledEventTypes eventsEnabled eventsExpiration eventsListeners failureFactor federatedUsers groups id identityProviderMappers identityProviders internationalizationEnabled keycloakVersion loginTheme loginWithEmailAllowed maxDeltaTimeSeconds maxFailureWaitSeconds minimumQuickLoginWaitSeconds notBefore offlineSessionIdleTimeout offlineSessionMaxLifespan offlineSessionMaxLifespanEnabled otpPolicyAlgorithm otpPolicyDigits otpPolicyInitialCounter otpPolicyLookAheadWindow otpPolicyPeriod otpPolicyType otpSupportedApplications passwordPolicy permanentLockout protocolMappers quickLoginCheckMilliSeconds realm refreshTokenMaxReuse registrationAllowed registrationEmailAsUsername registrationFlow rememberMe requiredActions resetCredentialsFlow resetPasswordAllowed revokeRefreshToken roles { client realm { id name scopeParamRequired composite clientRole containerId description attributes } } scopeMappings smtpServer sslRequired ssoSessionIdleTimeout ssoSessionIdleTimeoutRememberMe ssoSessionMaxLifespan ssoSessionMaxLifespanRememberMe supportedLocales userFederationMappers userFederationProviders userManagedAccessAllowed users { id createdTimestamp username enabled totp emailVerified firstName lastName email federationLink attributes disableableCredentialTypes requiredActions access clientRoles realmRoles groups serviceAccountClientId credentials { createdDate temporary type value algorithm counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel } } verifyEmail waitIncrementSeconds webAuthnPolicyAcceptableAaguids webAuthnPolicyAttestationConveyancePreference webAuthnPolicyAuthenticatorAttachment webAuthnPolicyAvoidSameAuthenticatorRegister webAuthnPolicyCreateTimeout webAuthnPolicyPasswordlessAcceptableAaguids webAuthnPolicyPasswordlessAttestationConveyancePreference webAuthnPolicyPasswordlessAuthenticatorAttachment webAuthnPolicyPasswordlessAvoidSameAuthenticatorRegister webAuthnPolicyPasswordlessCreateTimeout webAuthnPolicyPasswordlessRequireResidentKey webAuthnPolicyPasswordlessRpEntityName webAuthnPolicyPasswordlessRpId webAuthnPolicyPasswordlessSignatureAlgorithms webAuthnPolicyPasswordlessUserVerificationRequirement webAuthnPolicyRequireResidentKey webAuthnPolicyRpEntityName webAuthnPolicyRpId webAuthnPolicySignatureAlgorithms webAuthnPolicyUserVerificationRequirement }'} }', variables: {
      'realm': realm,
    }
    );
    return data['kcRealm'] == null ? null : KcRealmRepresentation.fromJson((data['kcRealm'] as Map).cast<String, dynamic>());
  }

  Future<KcKeyStoreConfig?> kcKeyStoreConfig({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcKeyStoreConfig(\$realm: String!) { kcKeyStoreConfig(realm: \$realm) ${selectionSet ?? '{ active { HS256 RS256 AES } keys { providerId providerPriority kid status type algorithm publicKey certificate } }'} }', variables: {
      'realm': realm,
    }
    );
    return data['kcKeyStoreConfig'] == null ? null : KcKeyStoreConfig.fromJson((data['kcKeyStoreConfig'] as Map).cast<String, dynamic>());
  }

  Future<List<KcEventRepresentation>?> kcEvents({required String realm, String? client, String? dateFrom, String? dateTo, String? first, String? ipAddress, String? max, String? user, String? selectionSet}) async {
    final data = await _executor.query('query kcEvents(\$realm: String!, \$client: String, \$dateFrom: String, \$dateTo: String, \$first: String, \$ipAddress: String, \$max: String, \$user: String) { kcEvents(realm: \$realm, client: \$client, dateFrom: \$dateFrom, dateTo: \$dateTo, first: \$first, ipAddress: \$ipAddress, max: \$max, user: \$user) ${selectionSet ?? '{ time type realmId clientId userId sessionId ipAddress details }'} }', variables: {
      'realm': realm,
      'client': ?client,
      'dateFrom': ?dateFrom,
      'dateTo': ?dateTo,
      'first': ?first,
      'ipAddress': ?ipAddress,
      'max': ?max,
      'user': ?user,
    }
    );
    return (data['kcEvents'] as List?)?.map((e) => (e == null ? null : KcEventRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<String>?> kcCredentialRegistrators({required String realm}) async {
    final data = await _executor.query('query kcCredentialRegistrators(\$realm: String!) { kcCredentialRegistrators(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return (data['kcCredentialRegistrators'] as List?)?.map((e) => (e?.toString())!).toList();
  }

  Future<List<KcUser>?> kcUsers({required String realm, String? briefRepresentation, String? email, String? emailVerified, String? enabled, String? exact, String? first, String? firstName, String? idpAlias, String? idpUserId, String? lastName, String? max, String? q, String? search, String? username, String? selectionSet}) async {
    final data = await _executor.query('query kcUsers(\$realm: String!, \$briefRepresentation: String, \$email: String, \$emailVerified: String, \$enabled: String, \$exact: String, \$first: String, \$firstName: String, \$idpAlias: String, \$idpUserId: String, \$lastName: String, \$max: String, \$q: String, \$search: String, \$username: String) { kcUsers(realm: \$realm, briefRepresentation: \$briefRepresentation, email: \$email, emailVerified: \$emailVerified, enabled: \$enabled, exact: \$exact, first: \$first, firstName: \$firstName, idpAlias: \$idpAlias, idpUserId: \$idpUserId, lastName: \$lastName, max: \$max, q: \$q, search: \$search, username: \$username) ${selectionSet ?? '{ id createdTimestamp username enabled totp emailVerified firstName lastName email federationLink attributes disableableCredentialTypes requiredActions access clientRoles realmRoles groups serviceAccountClientId credentials { createdDate temporary type value algorithm config { empty loadFactor threshold } counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel } }'} }', variables: {
      'realm': realm,
      'briefRepresentation': ?briefRepresentation,
      'email': ?email,
      'emailVerified': ?emailVerified,
      'enabled': ?enabled,
      'exact': ?exact,
      'first': ?first,
      'firstName': ?firstName,
      'idpAlias': ?idpAlias,
      'idpUserId': ?idpUserId,
      'lastName': ?lastName,
      'max': ?max,
      'q': ?q,
      'search': ?search,
      'username': ?username,
    }
    );
    return (data['kcUsers'] as List?)?.map((e) => (e == null ? null : KcUser.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<int?> kcUserCount({required String realm, String? briefRepresentation, String? email, String? emailVerified, String? enabled, String? exact, String? first, String? firstName, String? idpAlias, String? idpUserId, String? lastName, String? max, String? q, String? search, String? username}) async {
    final data = await _executor.query('query kcUserCount(\$realm: String!, \$briefRepresentation: String, \$email: String, \$emailVerified: String, \$enabled: String, \$exact: String, \$first: String, \$firstName: String, \$idpAlias: String, \$idpUserId: String, \$lastName: String, \$max: String, \$q: String, \$search: String, \$username: String) { kcUserCount(realm: \$realm, briefRepresentation: \$briefRepresentation, email: \$email, emailVerified: \$emailVerified, enabled: \$enabled, exact: \$exact, first: \$first, firstName: \$firstName, idpAlias: \$idpAlias, idpUserId: \$idpUserId, lastName: \$lastName, max: \$max, q: \$q, search: \$search, username: \$username) }', variables: {
      'realm': realm,
      'briefRepresentation': ?briefRepresentation,
      'email': ?email,
      'emailVerified': ?emailVerified,
      'enabled': ?enabled,
      'exact': ?exact,
      'first': ?first,
      'firstName': ?firstName,
      'idpAlias': ?idpAlias,
      'idpUserId': ?idpUserId,
      'lastName': ?lastName,
      'max': ?max,
      'q': ?q,
      'search': ?search,
      'username': ?username,
    }
    );
    return gqlInt(data['kcUserCount']);
  }

  Future<KcUser?> kcUserByID({required String realm, required String userID, String? selectionSet}) async {
    final data = await _executor.query('query kcUserByID(\$realm: String!, \$userID: String!) { kcUserByID(realm: \$realm, userID: \$userID) ${selectionSet ?? '{ id createdTimestamp username enabled totp emailVerified firstName lastName email federationLink attributes disableableCredentialTypes requiredActions access clientRoles realmRoles groups serviceAccountClientId credentials { createdDate temporary type value algorithm config { empty loadFactor threshold } counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel } }'} }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return data['kcUserByID'] == null ? null : KcUser.fromJson((data['kcUserByID'] as Map).cast<String, dynamic>());
  }

  Future<KcBruteForceStatus?> kcUserBruteForceDetectionStatus({required String realm, required String userID, String? selectionSet}) async {
    final data = await _executor.query('query kcUserBruteForceDetectionStatus(\$realm: String!, \$userID: String!) { kcUserBruteForceDetectionStatus(realm: \$realm, userID: \$userID) ${selectionSet ?? '{ numFailures disabled lastIPFailure lastFailure }'} }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return data['kcUserBruteForceDetectionStatus'] == null ? null : KcBruteForceStatus.fromJson((data['kcUserBruteForceDetectionStatus'] as Map).cast<String, dynamic>());
  }

  Future<List<KcUserSessionRepresentation>?> kcUserSessions({required String realm, required String userID, String? selectionSet}) async {
    final data = await _executor.query('query kcUserSessions(\$realm: String!, \$userID: String!) { kcUserSessions(realm: \$realm, userID: \$userID) ${selectionSet ?? '{ clients id ipAddress lastAccess start userId username }'} }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return (data['kcUserSessions'] as List?)?.map((e) => (e == null ? null : KcUserSessionRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcUserSessionRepresentation>?> kcUserOfflineSessionsForClient({required String realm, required String userID, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcUserOfflineSessionsForClient(\$realm: String!, \$userID: String!, \$clientID: String!) { kcUserOfflineSessionsForClient(realm: \$realm, userID: \$userID, clientID: \$clientID) ${selectionSet ?? '{ clients id ipAddress lastAccess start userId username }'} }', variables: {
      'realm': realm,
      'userID': userID,
      'clientID': clientID,
    }
    );
    return (data['kcUserOfflineSessionsForClient'] as List?)?.map((e) => (e == null ? null : KcUserSessionRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcGroup>?> kcUserGroups({required String realm, required String userID, String? briefRepresentation, String? exact, String? first, String? full, String? max, String? q, String? search, String? selectionSet}) async {
    final data = await _executor.query('query kcUserGroups(\$realm: String!, \$userID: String!, \$briefRepresentation: String, \$exact: String, \$first: String, \$full: String, \$max: String, \$q: String, \$search: String) { kcUserGroups(realm: \$realm, userID: \$userID, briefRepresentation: \$briefRepresentation, exact: \$exact, first: \$first, full: \$full, max: \$max, q: \$q, search: \$search) ${selectionSet ?? '{ id name path subGroups { id name path attributes access clientRoles realmRoles } attributes access clientRoles realmRoles }'} }', variables: {
      'realm': realm,
      'userID': userID,
      'briefRepresentation': ?briefRepresentation,
      'exact': ?exact,
      'first': ?first,
      'full': ?full,
      'max': ?max,
      'q': ?q,
      'search': ?search,
    }
    );
    return (data['kcUserGroups'] as List?)?.map((e) => (e == null ? null : KcGroup.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcCredentialRepresentation>?> kcCredentials({required String realm, required String userID, String? selectionSet}) async {
    final data = await _executor.query('query kcCredentials(\$realm: String!, \$userID: String!) { kcCredentials(realm: \$realm, userID: \$userID) ${selectionSet ?? '{ createdDate temporary type value algorithm config { empty loadFactor threshold } counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel }'} }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return (data['kcCredentials'] as List?)?.map((e) => (e == null ? null : KcCredentialRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<String>?> kcConfiguredUserStorageCredentialTypes({required String realm, required String userID}) async {
    final data = await _executor.query('query kcConfiguredUserStorageCredentialTypes(\$realm: String!, \$userID: String!) { kcConfiguredUserStorageCredentialTypes(realm: \$realm, userID: \$userID) }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return (data['kcConfiguredUserStorageCredentialTypes'] as List?)?.map((e) => (e?.toString())!).toList();
  }

  Future<List<KcFederatedIdentityRepresentation>?> kcUserFederatedIdentities({required String realm, required String userID, String? selectionSet}) async {
    final data = await _executor.query('query kcUserFederatedIdentities(\$realm: String!, \$userID: String!) { kcUserFederatedIdentities(realm: \$realm, userID: \$userID) ${selectionSet ?? '{ identityProvider userId userName }'} }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return (data['kcUserFederatedIdentities'] as List?)?.map((e) => (e == null ? null : KcFederatedIdentityRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcGroup>?> kcGroups({required String realm, String? briefRepresentation, String? exact, String? first, String? full, String? max, String? q, String? search, String? selectionSet}) async {
    final data = await _executor.query('query kcGroups(\$realm: String!, \$briefRepresentation: String, \$exact: String, \$first: String, \$full: String, \$max: String, \$q: String, \$search: String) { kcGroups(realm: \$realm, briefRepresentation: \$briefRepresentation, exact: \$exact, first: \$first, full: \$full, max: \$max, q: \$q, search: \$search) ${selectionSet ?? '{ id name path subGroups { id name path attributes access clientRoles realmRoles } attributes access clientRoles realmRoles }'} }', variables: {
      'realm': realm,
      'briefRepresentation': ?briefRepresentation,
      'exact': ?exact,
      'first': ?first,
      'full': ?full,
      'max': ?max,
      'q': ?q,
      'search': ?search,
    }
    );
    return (data['kcGroups'] as List?)?.map((e) => (e == null ? null : KcGroup.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<int?> kcGroupsCount({required String realm, String? briefRepresentation, String? exact, String? first, String? full, String? max, String? q, String? search}) async {
    final data = await _executor.query('query kcGroupsCount(\$realm: String!, \$briefRepresentation: String, \$exact: String, \$first: String, \$full: String, \$max: String, \$q: String, \$search: String) { kcGroupsCount(realm: \$realm, briefRepresentation: \$briefRepresentation, exact: \$exact, first: \$first, full: \$full, max: \$max, q: \$q, search: \$search) }', variables: {
      'realm': realm,
      'briefRepresentation': ?briefRepresentation,
      'exact': ?exact,
      'first': ?first,
      'full': ?full,
      'max': ?max,
      'q': ?q,
      'search': ?search,
    }
    );
    return gqlInt(data['kcGroupsCount']);
  }

  Future<KcGroup?> kcGroup({required String realm, required String groupID, String? selectionSet}) async {
    final data = await _executor.query('query kcGroup(\$realm: String!, \$groupID: String!) { kcGroup(realm: \$realm, groupID: \$groupID) ${selectionSet ?? '{ id name path subGroups { id name path attributes access clientRoles realmRoles } attributes access clientRoles realmRoles }'} }', variables: {
      'realm': realm,
      'groupID': groupID,
    }
    );
    return data['kcGroup'] == null ? null : KcGroup.fromJson((data['kcGroup'] as Map).cast<String, dynamic>());
  }

  Future<List<KcUser>?> kcGroupMembers({required String realm, required String groupID, String? briefRepresentation, String? exact, String? first, String? full, String? max, String? q, String? search, String? selectionSet}) async {
    final data = await _executor.query('query kcGroupMembers(\$realm: String!, \$groupID: String!, \$briefRepresentation: String, \$exact: String, \$first: String, \$full: String, \$max: String, \$q: String, \$search: String) { kcGroupMembers(realm: \$realm, groupID: \$groupID, briefRepresentation: \$briefRepresentation, exact: \$exact, first: \$first, full: \$full, max: \$max, q: \$q, search: \$search) ${selectionSet ?? '{ id createdTimestamp username enabled totp emailVerified firstName lastName email federationLink attributes disableableCredentialTypes requiredActions access clientRoles realmRoles groups serviceAccountClientId credentials { createdDate temporary type value algorithm config { empty loadFactor threshold } counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel } }'} }', variables: {
      'realm': realm,
      'groupID': groupID,
      'briefRepresentation': ?briefRepresentation,
      'exact': ?exact,
      'first': ?first,
      'full': ?full,
      'max': ?max,
      'q': ?q,
      'search': ?search,
    }
    );
    return (data['kcGroupMembers'] as List?)?.map((e) => (e == null ? null : KcUser.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcGroup?> kcGroupByPath({required String realm, required String path, String? selectionSet}) async {
    final data = await _executor.query('query kcGroupByPath(\$realm: String!, \$path: String!) { kcGroupByPath(realm: \$realm, path: \$path) ${selectionSet ?? '{ id name path subGroups { id name path attributes access clientRoles realmRoles } attributes access clientRoles realmRoles }'} }', variables: {
      'realm': realm,
      'path': path,
    }
    );
    return data['kcGroupByPath'] == null ? null : KcGroup.fromJson((data['kcGroupByPath'] as Map).cast<String, dynamic>());
  }

  Future<List<KcGroup>?> kcDefaultGroups({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcDefaultGroups(\$realm: String!) { kcDefaultGroups(realm: \$realm) ${selectionSet ?? '{ id name path subGroups { id name path attributes access clientRoles realmRoles } attributes access clientRoles realmRoles }'} }', variables: {
      'realm': realm,
    }
    );
    return (data['kcDefaultGroups'] as List?)?.map((e) => (e == null ? null : KcGroup.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcRole>?> kcRealmRoles({required String realm, String? first, String? max, String? search, String? briefRepresentation, String? selectionSet}) async {
    final data = await _executor.query('query kcRealmRoles(\$realm: String!, \$first: String, \$max: String, \$search: String, \$briefRepresentation: String) { kcRealmRoles(realm: \$realm, first: \$first, max: \$max, search: \$search, briefRepresentation: \$briefRepresentation) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'first': ?first,
      'max': ?max,
      'search': ?search,
      'briefRepresentation': ?briefRepresentation,
    }
    );
    return (data['kcRealmRoles'] as List?)?.map((e) => (e == null ? null : KcRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcRole?> kcRealmRole({required String realm, required String roleName, String? selectionSet}) async {
    final data = await _executor.query('query kcRealmRole(\$realm: String!, \$roleName: String!) { kcRealmRole(realm: \$realm, roleName: \$roleName) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'roleName': roleName,
    }
    );
    return data['kcRealmRole'] == null ? null : KcRole.fromJson((data['kcRealmRole'] as Map).cast<String, dynamic>());
  }

  Future<List<KcRole>?> kcCompositeRealmRoles({required String realm, required String roleName, String? selectionSet}) async {
    final data = await _executor.query('query kcCompositeRealmRoles(\$realm: String!, \$roleName: String!) { kcCompositeRealmRoles(realm: \$realm, roleName: \$roleName) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'roleName': roleName,
    }
    );
    return (data['kcCompositeRealmRoles'] as List?)?.map((e) => (e == null ? null : KcRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcUser>?> kcUsersByRoleName({required String realm, required String roleName, String? first, String? max, String? selectionSet}) async {
    final data = await _executor.query('query kcUsersByRoleName(\$realm: String!, \$roleName: String!, \$first: String, \$max: String) { kcUsersByRoleName(realm: \$realm, roleName: \$roleName, first: \$first, max: \$max) ${selectionSet ?? '{ id createdTimestamp username enabled totp emailVerified firstName lastName email federationLink attributes disableableCredentialTypes requiredActions access clientRoles realmRoles groups serviceAccountClientId credentials { createdDate temporary type value algorithm config { empty loadFactor threshold } counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel } }'} }', variables: {
      'realm': realm,
      'roleName': roleName,
      'first': ?first,
      'max': ?max,
    }
    );
    return (data['kcUsersByRoleName'] as List?)?.map((e) => (e == null ? null : KcUser.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcGroup>?> kcGroupsByRole({required String realm, required String roleName, String? selectionSet}) async {
    final data = await _executor.query('query kcGroupsByRole(\$realm: String!, \$roleName: String!) { kcGroupsByRole(realm: \$realm, roleName: \$roleName) ${selectionSet ?? '{ id name path subGroups { id name path attributes access clientRoles realmRoles } attributes access clientRoles realmRoles }'} }', variables: {
      'realm': realm,
      'roleName': roleName,
    }
    );
    return (data['kcGroupsByRole'] as List?)?.map((e) => (e == null ? null : KcGroup.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcRole?> kcRealmRoleByID({required String realm, required String roleID, String? selectionSet}) async {
    final data = await _executor.query('query kcRealmRoleByID(\$realm: String!, \$roleID: String!) { kcRealmRoleByID(realm: \$realm, roleID: \$roleID) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'roleID': roleID,
    }
    );
    return data['kcRealmRoleByID'] == null ? null : KcRole.fromJson((data['kcRealmRoleByID'] as Map).cast<String, dynamic>());
  }

  Future<List<KcRole>?> kcCompositeRolesByRoleID({required String realm, required String roleID, String? selectionSet}) async {
    final data = await _executor.query('query kcCompositeRolesByRoleID(\$realm: String!, \$roleID: String!) { kcCompositeRolesByRoleID(realm: \$realm, roleID: \$roleID) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'roleID': roleID,
    }
    );
    return (data['kcCompositeRolesByRoleID'] as List?)?.map((e) => (e == null ? null : KcRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcRole>?> kcCompositeRealmRolesByRoleID({required String realm, required String roleID, String? selectionSet}) async {
    final data = await _executor.query('query kcCompositeRealmRolesByRoleID(\$realm: String!, \$roleID: String!) { kcCompositeRealmRolesByRoleID(realm: \$realm, roleID: \$roleID) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'roleID': roleID,
    }
    );
    return (data['kcCompositeRealmRolesByRoleID'] as List?)?.map((e) => (e == null ? null : KcRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcRole>?> kcCompositeClientRolesByRoleID({required String realm, required String roleID, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcCompositeClientRolesByRoleID(\$realm: String!, \$roleID: String!, \$clientID: String!) { kcCompositeClientRolesByRoleID(realm: \$realm, roleID: \$roleID, clientID: \$clientID) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'roleID': roleID,
      'clientID': clientID,
    }
    );
    return (data['kcCompositeClientRolesByRoleID'] as List?)?.map((e) => (e == null ? null : KcRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcClient>?> kcClientsWithParams({required String realm, String? clientId, String? viewableOnly, String? first, String? max, String? search, String? q, String? selectionSet}) async {
    final data = await _executor.query('query kcClientsWithParams(\$realm: String!, \$clientId: String, \$viewableOnly: String, \$first: String, \$max: String, \$search: String, \$q: String) { kcClientsWithParams(realm: \$realm, clientId: \$clientId, viewableOnly: \$viewableOnly, first: \$first, max: \$max, search: \$search, q: \$q) ${selectionSet ?? '{ access adminUrl attributes authenticationFlowBindingOverrides authorizationServicesEnabled authorizationSettings { allowRemoteResourceManagement clientId id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } policyEnforcementMode resources { _id attributes displayName icon_uri name ownerManagedAccess type uris } scopes { displayName iconUri id name } decisionStrategy } baseUrl bearerOnly clientAuthenticatorType clientId consentRequired defaultClientScopes defaultRoles description directAccessGrantsEnabled enabled frontchannelLogout fullScopeAllowed id implicitFlowEnabled name nodeReRegistrationTimeout notBefore optionalClientScopes origin protocol protocolMappers { config id name protocol protocolMapper consentRequired } publicClient redirectUris registeredNodes registrationAccessToken rootUrl secret serviceAccountsEnabled standardFlowEnabled surrogateAuthRequired webOrigins }'} }', variables: {
      'realm': realm,
      'clientId': ?clientId,
      'viewableOnly': ?viewableOnly,
      'first': ?first,
      'max': ?max,
      'search': ?search,
      'q': ?q,
    }
    );
    return (data['kcClientsWithParams'] as List?)?.map((e) => (e == null ? null : KcClient.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcClient?> kcClient({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClient(\$realm: String!, \$clientID: String!) { kcClient(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ access adminUrl attributes authenticationFlowBindingOverrides authorizationServicesEnabled authorizationSettings { allowRemoteResourceManagement clientId id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } policyEnforcementMode resources { _id attributes displayName icon_uri name ownerManagedAccess type uris } scopes { displayName iconUri id name } decisionStrategy } baseUrl bearerOnly clientAuthenticatorType clientId consentRequired defaultClientScopes defaultRoles description directAccessGrantsEnabled enabled frontchannelLogout fullScopeAllowed id implicitFlowEnabled name nodeReRegistrationTimeout notBefore optionalClientScopes origin protocol protocolMappers { config id name protocol protocolMapper consentRequired } publicClient redirectUris registeredNodes registrationAccessToken rootUrl secret serviceAccountsEnabled standardFlowEnabled surrogateAuthRequired webOrigins }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcClient'] == null ? null : KcClient.fromJson((data['kcClient'] as Map).cast<String, dynamic>());
  }

  Future<KcCredentialRepresentation?> kcClientSecret({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientSecret(\$realm: String!, \$clientID: String!) { kcClientSecret(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ createdDate temporary type value algorithm config { empty loadFactor threshold } counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcClientSecret'] == null ? null : KcCredentialRepresentation.fromJson((data['kcClientSecret'] as Map).cast<String, dynamic>());
  }

  Future<KcUser?> kcClientServiceAccount({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientServiceAccount(\$realm: String!, \$clientID: String!) { kcClientServiceAccount(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ id createdTimestamp username enabled totp emailVerified firstName lastName email federationLink attributes disableableCredentialTypes requiredActions access clientRoles realmRoles groups serviceAccountClientId credentials { createdDate temporary type value algorithm config { empty loadFactor threshold } counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel } }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcClientServiceAccount'] == null ? null : KcUser.fromJson((data['kcClientServiceAccount'] as Map).cast<String, dynamic>());
  }

  Future<List<KcUserSessionRepresentation>?> kcClientUserSessions({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientUserSessions(\$realm: String!, \$clientID: String!) { kcClientUserSessions(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ clients id ipAddress lastAccess start userId username }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return (data['kcClientUserSessions'] as List?)?.map((e) => (e == null ? null : KcUserSessionRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcUserSessionRepresentation>?> kcClientOfflineSessions({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientOfflineSessions(\$realm: String!, \$clientID: String!) { kcClientOfflineSessions(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ clients id ipAddress lastAccess start userId username }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return (data['kcClientOfflineSessions'] as List?)?.map((e) => (e == null ? null : KcUserSessionRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcClientScope>?> kcClientsDefaultScopes({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientsDefaultScopes(\$realm: String!, \$clientID: String!) { kcClientsDefaultScopes(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ id name description protocol attributes { consent_screen_text display_on_consent_screen include_in_token_scope } protocolMappers { id name protocol protocolMapper consentRequired config { userinfo_token_claim user_attribute id_token_claim access_token_claim claim_name claim_value jsonType_label multivalued usermodel_clientRoleMapping_clientId included_client_audience full_path attribute_name attribute_nameformat single } } }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return (data['kcClientsDefaultScopes'] as List?)?.map((e) => (e == null ? null : KcClientScope.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcClientScope>?> kcClientsOptionalScopes({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientsOptionalScopes(\$realm: String!, \$clientID: String!) { kcClientsOptionalScopes(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ id name description protocol attributes { consent_screen_text display_on_consent_screen include_in_token_scope } protocolMappers { id name protocol protocolMapper consentRequired config { userinfo_token_claim user_attribute id_token_claim access_token_claim claim_name claim_value jsonType_label multivalued usermodel_clientRoleMapping_clientId included_client_audience full_path attribute_name attribute_nameformat single } } }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return (data['kcClientsOptionalScopes'] as List?)?.map((e) => (e == null ? null : KcClientScope.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcMappingsRepresentation?> kcClientScopeMappings({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientScopeMappings(\$realm: String!, \$clientID: String!) { kcClientScopeMappings(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ clientMappings realmMappings { id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes } }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcClientScopeMappings'] == null ? null : KcMappingsRepresentation.fromJson((data['kcClientScopeMappings'] as Map).cast<String, dynamic>());
  }

  Future<List<KcRole>?> kcClientRoles({required String realm, required String clientID, String? first, String? max, String? search, String? briefRepresentation, String? selectionSet}) async {
    final data = await _executor.query('query kcClientRoles(\$realm: String!, \$clientID: String!, \$first: String, \$max: String, \$search: String, \$briefRepresentation: String) { kcClientRoles(realm: \$realm, clientID: \$clientID, first: \$first, max: \$max, search: \$search, briefRepresentation: \$briefRepresentation) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'first': ?first,
      'max': ?max,
      'search': ?search,
      'briefRepresentation': ?briefRepresentation,
    }
    );
    return (data['kcClientRoles'] as List?)?.map((e) => (e == null ? null : KcRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcRole?> kcClientRole({required String realm, required String clientID, required String roleName, String? selectionSet}) async {
    final data = await _executor.query('query kcClientRole(\$realm: String!, \$clientID: String!, \$roleName: String!) { kcClientRole(realm: \$realm, clientID: \$clientID, roleName: \$roleName) ${selectionSet ?? '{ id name scopeParamRequired composite composites { client realm } clientRole containerId description attributes }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'roleName': roleName,
    }
    );
    return data['kcClientRole'] == null ? null : KcRole.fromJson((data['kcClientRole'] as Map).cast<String, dynamic>());
  }

  Future<List<KcUser>?> kcUsersByClientRoleName({required String realm, required String clientID, required String roleName, String? first, String? max, String? selectionSet}) async {
    final data = await _executor.query('query kcUsersByClientRoleName(\$realm: String!, \$clientID: String!, \$roleName: String!, \$first: String, \$max: String) { kcUsersByClientRoleName(realm: \$realm, clientID: \$clientID, roleName: \$roleName, first: \$first, max: \$max) ${selectionSet ?? '{ id createdTimestamp username enabled totp emailVerified firstName lastName email federationLink attributes disableableCredentialTypes requiredActions access clientRoles realmRoles groups serviceAccountClientId credentials { createdDate temporary type value algorithm config { empty loadFactor threshold } counter device digits hashIterations hashedSaltedValue period salt credentialData id priority secretData userLabel } }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'roleName': roleName,
      'first': ?first,
      'max': ?max,
    }
    );
    return (data['kcUsersByClientRoleName'] as List?)?.map((e) => (e == null ? null : KcUser.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcGroup>?> kcGroupsByClientRole({required String realm, required String clientID, required String roleName, String? selectionSet}) async {
    final data = await _executor.query('query kcGroupsByClientRole(\$realm: String!, \$clientID: String!, \$roleName: String!) { kcGroupsByClientRole(realm: \$realm, clientID: \$clientID, roleName: \$roleName) ${selectionSet ?? '{ id name path subGroups { id name path attributes access clientRoles realmRoles } attributes access clientRoles realmRoles }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'roleName': roleName,
    }
    );
    return (data['kcGroupsByClientRole'] as List?)?.map((e) => (e == null ? null : KcGroup.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcResourceServerRepresentation?> kcResourceServer({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcResourceServer(\$realm: String!, \$clientID: String!) { kcResourceServer(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ allowRemoteResourceManagement clientId id name policies { config decisionStrategy description id logic name owner policies resources scopes type roles { id private required } code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groups { id path extendChildren } groupsClaim } policyEnforcementMode resources { _id attributes displayName icon_uri name owner { id name } ownerManagedAccess resource_scopes { displayName iconUri id name } scopes { displayName iconUri id name } type uris } scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } resources { _id attributes displayName icon_uri name ownerManagedAccess type uris } } decisionStrategy }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcResourceServer'] == null ? null : KcResourceServerRepresentation.fromJson((data['kcResourceServer'] as Map).cast<String, dynamic>());
  }

  Future<List<KcResourceRepresentation>?> kcResources({required String realm, required String clientID, String? deep, String? first, String? max, String? name_, String? owner, String? type_, String? uri, String? scope, String? matchingUri, String? exactName, String? selectionSet}) async {
    final data = await _executor.query('query kcResources(\$realm: String!, \$clientID: String!, \$deep: String, \$first: String, \$max: String, \$name: String, \$owner: String, \$type: String, \$uri: String, \$scope: String, \$matchingUri: String, \$exactName: String) { kcResources(realm: \$realm, clientID: \$clientID, deep: \$deep, first: \$first, max: \$max, name: \$name, owner: \$owner, type: \$type, uri: \$uri, scope: \$scope, matchingUri: \$matchingUri, exactName: \$exactName) ${selectionSet ?? '{ _id attributes displayName icon_uri name owner { id name } ownerManagedAccess resource_scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } } scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } } type uris }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'deep': ?deep,
      'first': ?first,
      'max': ?max,
      'name': ?name_,
      'owner': ?owner,
      'type': ?type_,
      'uri': ?uri,
      'scope': ?scope,
      'matchingUri': ?matchingUri,
      'exactName': ?exactName,
    }
    );
    return (data['kcResources'] as List?)?.map((e) => (e == null ? null : KcResourceRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcResourceRepresentation?> kcResource({required String realm, required String clientID, required String resourceID, String? selectionSet}) async {
    final data = await _executor.query('query kcResource(\$realm: String!, \$clientID: String!, \$resourceID: String!) { kcResource(realm: \$realm, clientID: \$clientID, resourceID: \$resourceID) ${selectionSet ?? '{ _id attributes displayName icon_uri name owner { id name } ownerManagedAccess resource_scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } } scopes { displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } } type uris }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'resourceID': resourceID,
    }
    );
    return data['kcResource'] == null ? null : KcResourceRepresentation.fromJson((data['kcResource'] as Map).cast<String, dynamic>());
  }

  Future<List<KcScopeRepresentation>?> kcScopes({required String realm, required String clientID, String? deep, String? first, String? max, String? name_, String? selectionSet}) async {
    final data = await _executor.query('query kcScopes(\$realm: String!, \$clientID: String!, \$deep: String, \$first: String, \$max: String, \$name: String) { kcScopes(realm: \$realm, clientID: \$clientID, deep: \$deep, first: \$first, max: \$max, name: \$name) ${selectionSet ?? '{ displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type roles { id private required } code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groups { id path extendChildren } groupsClaim } resources { _id attributes displayName icon_uri name owner { id name } ownerManagedAccess type uris } }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'deep': ?deep,
      'first': ?first,
      'max': ?max,
      'name': ?name_,
    }
    );
    return (data['kcScopes'] as List?)?.map((e) => (e == null ? null : KcScopeRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcScopeRepresentation?> kcScope({required String realm, required String clientID, required String scopeID, String? selectionSet}) async {
    final data = await _executor.query('query kcScope(\$realm: String!, \$clientID: String!, \$scopeID: String!) { kcScope(realm: \$realm, clientID: \$clientID, scopeID: \$scopeID) ${selectionSet ?? '{ displayName iconUri id name policies { config decisionStrategy description id logic name owner policies resources scopes type roles { id private required } code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groups { id path extendChildren } groupsClaim } resources { _id attributes displayName icon_uri name owner { id name } ownerManagedAccess type uris } }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'scopeID': scopeID,
    }
    );
    return data['kcScope'] == null ? null : KcScopeRepresentation.fromJson((data['kcScope'] as Map).cast<String, dynamic>());
  }

  Future<List<KcPolicyRepresentation>?> kcPolicies({required String realm, required String clientID, String? first, String? max, String? name_, String? permission, String? type_, String? selectionSet}) async {
    final data = await _executor.query('query kcPolicies(\$realm: String!, \$clientID: String!, \$first: String, \$max: String, \$name: String, \$permission: String, \$type: String) { kcPolicies(realm: \$realm, clientID: \$clientID, first: \$first, max: \$max, name: \$name, permission: \$permission, type: \$type) ${selectionSet ?? '{ config decisionStrategy description id logic name owner policies resources scopes type roles { id private required } code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groups { id path extendChildren } groupsClaim }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'first': ?first,
      'max': ?max,
      'name': ?name_,
      'permission': ?permission,
      'type': ?type_,
    }
    );
    return (data['kcPolicies'] as List?)?.map((e) => (e == null ? null : KcPolicyRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcPolicyRepresentation?> kcPolicy({required String realm, required String clientID, required String policyID, String? selectionSet}) async {
    final data = await _executor.query('query kcPolicy(\$realm: String!, \$clientID: String!, \$policyID: String!) { kcPolicy(realm: \$realm, clientID: \$clientID, policyID: \$policyID) ${selectionSet ?? '{ config decisionStrategy description id logic name owner policies resources scopes type roles { id private required } code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groups { id path extendChildren } groupsClaim }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'policyID': policyID,
    }
    );
    return data['kcPolicy'] == null ? null : KcPolicyRepresentation.fromJson((data['kcPolicy'] as Map).cast<String, dynamic>());
  }

  Future<List<KcPolicyRepresentation>?> kcAuthorizationPolicyAssociatedPolicies({required String realm, required String clientID, required String policyID, String? selectionSet}) async {
    final data = await _executor.query('query kcAuthorizationPolicyAssociatedPolicies(\$realm: String!, \$clientID: String!, \$policyID: String!) { kcAuthorizationPolicyAssociatedPolicies(realm: \$realm, clientID: \$clientID, policyID: \$policyID) ${selectionSet ?? '{ config decisionStrategy description id logic name owner policies resources scopes type roles { id private required } code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groups { id path extendChildren } groupsClaim }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'policyID': policyID,
    }
    );
    return (data['kcAuthorizationPolicyAssociatedPolicies'] as List?)?.map((e) => (e == null ? null : KcPolicyRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcPolicyResourceRepresentation>?> kcAuthorizationPolicyResources({required String realm, required String clientID, required String policyID, String? selectionSet}) async {
    final data = await _executor.query('query kcAuthorizationPolicyResources(\$realm: String!, \$clientID: String!, \$policyID: String!) { kcAuthorizationPolicyResources(realm: \$realm, clientID: \$clientID, policyID: \$policyID) ${selectionSet ?? '{ _id name }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'policyID': policyID,
    }
    );
    return (data['kcAuthorizationPolicyResources'] as List?)?.map((e) => (e == null ? null : KcPolicyResourceRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcPolicyScopeRepresentation>?> kcAuthorizationPolicyScopes({required String realm, required String clientID, required String policyID, String? selectionSet}) async {
    final data = await _executor.query('query kcAuthorizationPolicyScopes(\$realm: String!, \$clientID: String!, \$policyID: String!) { kcAuthorizationPolicyScopes(realm: \$realm, clientID: \$clientID, policyID: \$policyID) ${selectionSet ?? '{ id name }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'policyID': policyID,
    }
    );
    return (data['kcAuthorizationPolicyScopes'] as List?)?.map((e) => (e == null ? null : KcPolicyScopeRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcPermissionRepresentation>?> kcDependentPermissions({required String realm, required String clientID, required String policyID, String? selectionSet}) async {
    final data = await _executor.query('query kcDependentPermissions(\$realm: String!, \$clientID: String!, \$policyID: String!) { kcDependentPermissions(realm: \$realm, clientID: \$clientID, policyID: \$policyID) ${selectionSet ?? '{ decisionStrategy description id logic name policies resources resourceType scopes type }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'policyID': policyID,
    }
    );
    return (data['kcDependentPermissions'] as List?)?.map((e) => (e == null ? null : KcPermissionRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcPermissionRepresentation>?> kcPermissions({required String realm, required String clientID, String? first, String? max, String? name_, String? resource, String? scope, String? type_, String? selectionSet}) async {
    final data = await _executor.query('query kcPermissions(\$realm: String!, \$clientID: String!, \$first: String, \$max: String, \$name: String, \$resource: String, \$scope: String, \$type: String) { kcPermissions(realm: \$realm, clientID: \$clientID, first: \$first, max: \$max, name: \$name, resource: \$resource, scope: \$scope, type: \$type) ${selectionSet ?? '{ decisionStrategy description id logic name policies resources resourceType scopes type }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'first': ?first,
      'max': ?max,
      'name': ?name_,
      'resource': ?resource,
      'scope': ?scope,
      'type': ?type_,
    }
    );
    return (data['kcPermissions'] as List?)?.map((e) => (e == null ? null : KcPermissionRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcPermissionRepresentation?> kcPermission({required String realm, required String clientID, required String permissionID, String? selectionSet}) async {
    final data = await _executor.query('query kcPermission(\$realm: String!, \$clientID: String!, \$permissionID: String!) { kcPermission(realm: \$realm, clientID: \$clientID, permissionID: \$permissionID) ${selectionSet ?? '{ decisionStrategy description id logic name policies resources resourceType scopes type }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'permissionID': permissionID,
    }
    );
    return data['kcPermission'] == null ? null : KcPermissionRepresentation.fromJson((data['kcPermission'] as Map).cast<String, dynamic>());
  }

  Future<List<KcPermissionResource>?> kcPermissionResources({required String realm, required String clientID, required String permissionID, String? selectionSet}) async {
    final data = await _executor.query('query kcPermissionResources(\$realm: String!, \$clientID: String!, \$permissionID: String!) { kcPermissionResources(realm: \$realm, clientID: \$clientID, permissionID: \$permissionID) ${selectionSet ?? '{ _id name }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'permissionID': permissionID,
    }
    );
    return (data['kcPermissionResources'] as List?)?.map((e) => (e == null ? null : KcPermissionResource.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcPermissionScope>?> kcPermissionScopes({required String realm, required String clientID, required String permissionID, String? selectionSet}) async {
    final data = await _executor.query('query kcPermissionScopes(\$realm: String!, \$clientID: String!, \$permissionID: String!) { kcPermissionScopes(realm: \$realm, clientID: \$clientID, permissionID: \$permissionID) ${selectionSet ?? '{ id name }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
      'permissionID': permissionID,
    }
    );
    return (data['kcPermissionScopes'] as List?)?.map((e) => (e == null ? null : KcPermissionScope.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcClient?> kcClientRepresentation({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientRepresentation(\$realm: String!, \$clientID: String!) { kcClientRepresentation(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ access adminUrl attributes authenticationFlowBindingOverrides authorizationServicesEnabled authorizationSettings { allowRemoteResourceManagement clientId id name policies { config decisionStrategy description id logic name owner policies resources scopes type code clients notBefore notOnOrAfter dayMonth dayMonthEnd month monthEnd year yearEnd hour hourEnd minute minuteEnd users groupsClaim } policyEnforcementMode resources { _id attributes displayName icon_uri name ownerManagedAccess type uris } scopes { displayName iconUri id name } decisionStrategy } baseUrl bearerOnly clientAuthenticatorType clientId consentRequired defaultClientScopes defaultRoles description directAccessGrantsEnabled enabled frontchannelLogout fullScopeAllowed id implicitFlowEnabled name nodeReRegistrationTimeout notBefore optionalClientScopes origin protocol protocolMappers { config id name protocol protocolMapper consentRequired } publicClient redirectUris registeredNodes registrationAccessToken rootUrl secret serviceAccountsEnabled standardFlowEnabled surrogateAuthRequired webOrigins }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcClientRepresentation'] == null ? null : KcClient.fromJson((data['kcClientRepresentation'] as Map).cast<String, dynamic>());
  }

  Future<KcAdapterConfiguration?> kcAdapterConfiguration({required String realm, required String clientID, String? selectionSet}) async {
    final data = await _executor.query('query kcAdapterConfiguration(\$realm: String!, \$clientID: String!) { kcAdapterConfiguration(realm: \$realm, clientID: \$clientID) ${selectionSet ?? '{ realm auth_server_url ssl_required resource credentials confidential_port }'} }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcAdapterConfiguration'] == null ? null : KcAdapterConfiguration.fromJson((data['kcAdapterConfiguration'] as Map).cast<String, dynamic>());
  }

  Future<List<KcClientScope>?> kcClientScopes({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcClientScopes(\$realm: String!) { kcClientScopes(realm: \$realm) ${selectionSet ?? '{ id name description protocol attributes { consent_screen_text display_on_consent_screen include_in_token_scope } protocolMappers { id name protocol protocolMapper consentRequired config { userinfo_token_claim user_attribute id_token_claim access_token_claim claim_name claim_value jsonType_label multivalued usermodel_clientRoleMapping_clientId included_client_audience full_path attribute_name attribute_nameformat single } } }'} }', variables: {
      'realm': realm,
    }
    );
    return (data['kcClientScopes'] as List?)?.map((e) => (e == null ? null : KcClientScope.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcClientScope?> kcClientScope({required String realm, required String scopeID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientScope(\$realm: String!, \$scopeID: String!) { kcClientScope(realm: \$realm, scopeID: \$scopeID) ${selectionSet ?? '{ id name description protocol attributes { consent_screen_text display_on_consent_screen include_in_token_scope } protocolMappers { id name protocol protocolMapper consentRequired config { userinfo_token_claim user_attribute id_token_claim access_token_claim claim_name claim_value jsonType_label multivalued usermodel_clientRoleMapping_clientId included_client_audience full_path attribute_name attribute_nameformat single } } }'} }', variables: {
      'realm': realm,
      'scopeID': scopeID,
    }
    );
    return data['kcClientScope'] == null ? null : KcClientScope.fromJson((data['kcClientScope'] as Map).cast<String, dynamic>());
  }

  Future<List<KcProtocolMappers>?> kcClientScopeProtocolMappers({required String realm, required String scopeID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientScopeProtocolMappers(\$realm: String!, \$scopeID: String!) { kcClientScopeProtocolMappers(realm: \$realm, scopeID: \$scopeID) ${selectionSet ?? '{ id name protocol protocolMapper consentRequired config { userinfo_token_claim user_attribute id_token_claim access_token_claim claim_name claim_value jsonType_label multivalued usermodel_clientRoleMapping_clientId included_client_audience full_path attribute_name attribute_nameformat single } }'} }', variables: {
      'realm': realm,
      'scopeID': scopeID,
    }
    );
    return (data['kcClientScopeProtocolMappers'] as List?)?.map((e) => (e == null ? null : KcProtocolMappers.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcProtocolMappers?> kcClientScopeProtocolMapper({required String realm, required String scopeID, required String mapperID, String? selectionSet}) async {
    final data = await _executor.query('query kcClientScopeProtocolMapper(\$realm: String!, \$scopeID: String!, \$mapperID: String!) { kcClientScopeProtocolMapper(realm: \$realm, scopeID: \$scopeID, mapperID: \$mapperID) ${selectionSet ?? '{ id name protocol protocolMapper consentRequired config { userinfo_token_claim user_attribute id_token_claim access_token_claim claim_name claim_value jsonType_label multivalued usermodel_clientRoleMapping_clientId included_client_audience full_path attribute_name attribute_nameformat single } }'} }', variables: {
      'realm': realm,
      'scopeID': scopeID,
      'mapperID': mapperID,
    }
    );
    return data['kcClientScopeProtocolMapper'] == null ? null : KcProtocolMappers.fromJson((data['kcClientScopeProtocolMapper'] as Map).cast<String, dynamic>());
  }

  Future<List<KcClientScope>?> kcDefaultDefaultClientScopes({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcDefaultDefaultClientScopes(\$realm: String!) { kcDefaultDefaultClientScopes(realm: \$realm) ${selectionSet ?? '{ id name description protocol attributes { consent_screen_text display_on_consent_screen include_in_token_scope } protocolMappers { id name protocol protocolMapper consentRequired config { userinfo_token_claim user_attribute id_token_claim access_token_claim claim_name claim_value jsonType_label multivalued usermodel_clientRoleMapping_clientId included_client_audience full_path attribute_name attribute_nameformat single } } }'} }', variables: {
      'realm': realm,
    }
    );
    return (data['kcDefaultDefaultClientScopes'] as List?)?.map((e) => (e == null ? null : KcClientScope.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcClientScope>?> kcDefaultOptionalClientScopes({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcDefaultOptionalClientScopes(\$realm: String!) { kcDefaultOptionalClientScopes(realm: \$realm) ${selectionSet ?? '{ id name description protocol attributes { consent_screen_text display_on_consent_screen include_in_token_scope } protocolMappers { id name protocol protocolMapper consentRequired config { userinfo_token_claim user_attribute id_token_claim access_token_claim claim_name claim_value jsonType_label multivalued usermodel_clientRoleMapping_clientId included_client_audience full_path attribute_name attribute_nameformat single } } }'} }', variables: {
      'realm': realm,
    }
    );
    return (data['kcDefaultOptionalClientScopes'] as List?)?.map((e) => (e == null ? null : KcClientScope.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcComponent>?> kcComponentsWithParams({required String realm, String? name_, String? provider, String? parent, String? selectionSet}) async {
    final data = await _executor.query('query kcComponentsWithParams(\$realm: String!, \$name: String, \$provider: String, \$parent: String) { kcComponentsWithParams(realm: \$realm, name: \$name, provider: \$provider, parent: \$parent) ${selectionSet ?? '{ id name providerId providerType parentId config subType }'} }', variables: {
      'realm': realm,
      'name': ?name_,
      'provider': ?provider,
      'parent': ?parent,
    }
    );
    return (data['kcComponentsWithParams'] as List?)?.map((e) => (e == null ? null : KcComponent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcComponent?> kcComponent({required String realm, required String componentID, String? selectionSet}) async {
    final data = await _executor.query('query kcComponent(\$realm: String!, \$componentID: String!) { kcComponent(realm: \$realm, componentID: \$componentID) ${selectionSet ?? '{ id name providerId providerType parentId config subType }'} }', variables: {
      'realm': realm,
      'componentID': componentID,
    }
    );
    return data['kcComponent'] == null ? null : KcComponent.fromJson((data['kcComponent'] as Map).cast<String, dynamic>());
  }

  Future<List<KcAuthenticationFlowRepresentation>?> kcAuthenticationFlows({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcAuthenticationFlows(\$realm: String!) { kcAuthenticationFlows(realm: \$realm) ${selectionSet ?? '{ alias authenticationExecutions { authenticator authenticatorConfig authenticatorFlow autheticatorFlow flowAlias priority requirement userSetupAllowed } builtIn description id providerId topLevel }'} }', variables: {
      'realm': realm,
    }
    );
    return (data['kcAuthenticationFlows'] as List?)?.map((e) => (e == null ? null : KcAuthenticationFlowRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcAuthenticationFlowRepresentation?> kcAuthenticationFlow({required String realm, required String flow, String? selectionSet}) async {
    final data = await _executor.query('query kcAuthenticationFlow(\$realm: String!, \$flow: String!) { kcAuthenticationFlow(realm: \$realm, flow: \$flow) ${selectionSet ?? '{ alias authenticationExecutions { authenticator authenticatorConfig authenticatorFlow autheticatorFlow flowAlias priority requirement userSetupAllowed } builtIn description id providerId topLevel }'} }', variables: {
      'realm': realm,
      'flow': flow,
    }
    );
    return data['kcAuthenticationFlow'] == null ? null : KcAuthenticationFlowRepresentation.fromJson((data['kcAuthenticationFlow'] as Map).cast<String, dynamic>());
  }

  Future<List<KcModifyAuthenticationExecutionRepresentation>?> kcAuthenticationExecutions({required String realm, required String flow, String? selectionSet}) async {
    final data = await _executor.query('query kcAuthenticationExecutions(\$realm: String!, \$flow: String!) { kcAuthenticationExecutions(realm: \$realm, flow: \$flow) ${selectionSet ?? '{ id providerId authenticationConfig authenticationFlow requirement flowId displayName alias requirementChoices configurable level index description }'} }', variables: {
      'realm': realm,
      'flow': flow,
    }
    );
    return (data['kcAuthenticationExecutions'] as List?)?.map((e) => (e == null ? null : KcModifyAuthenticationExecutionRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<KcRequiredActionProviderRepresentation>?> kcRequiredActions({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcRequiredActions(\$realm: String!) { kcRequiredActions(realm: \$realm) ${selectionSet ?? '{ alias config defaultAction enabled name priority providerId }'} }', variables: {
      'realm': realm,
    }
    );
    return (data['kcRequiredActions'] as List?)?.map((e) => (e == null ? null : KcRequiredActionProviderRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcRequiredActionProviderRepresentation?> kcRequiredAction({required String realm, required String alias, String? selectionSet}) async {
    final data = await _executor.query('query kcRequiredAction(\$realm: String!, \$alias: String!) { kcRequiredAction(realm: \$realm, alias: \$alias) ${selectionSet ?? '{ alias config defaultAction enabled name priority providerId }'} }', variables: {
      'realm': realm,
      'alias': alias,
    }
    );
    return data['kcRequiredAction'] == null ? null : KcRequiredActionProviderRepresentation.fromJson((data['kcRequiredAction'] as Map).cast<String, dynamic>());
  }

  Future<List<KcIdentityProviderRepresentation>?> kcIdentityProviders({required String realm, String? selectionSet}) async {
    final data = await _executor.query('query kcIdentityProviders(\$realm: String!) { kcIdentityProviders(realm: \$realm) ${selectionSet ?? '{ addReadTokenRoleOnCreate alias config displayName enabled firstBrokerLoginFlowAlias internalId linkOnly postBrokerLoginFlowAlias providerId storeToken trustEmail }'} }', variables: {
      'realm': realm,
    }
    );
    return (data['kcIdentityProviders'] as List?)?.map((e) => (e == null ? null : KcIdentityProviderRepresentation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcIdentityProviderRepresentation?> kcIdentityProvider({required String realm, required String alias, String? selectionSet}) async {
    final data = await _executor.query('query kcIdentityProvider(\$realm: String!, \$alias: String!) { kcIdentityProvider(realm: \$realm, alias: \$alias) ${selectionSet ?? '{ addReadTokenRoleOnCreate alias config displayName enabled firstBrokerLoginFlowAlias internalId linkOnly postBrokerLoginFlowAlias providerId storeToken trustEmail }'} }', variables: {
      'realm': realm,
      'alias': alias,
    }
    );
    return data['kcIdentityProvider'] == null ? null : KcIdentityProviderRepresentation.fromJson((data['kcIdentityProvider'] as Map).cast<String, dynamic>());
  }

  Future<String?> kcExportIDPPublicBrokerConfig({required String realm, required String alias}) async {
    final data = await _executor.query('query kcExportIDPPublicBrokerConfig(\$realm: String!, \$alias: String!) { kcExportIDPPublicBrokerConfig(realm: \$realm, alias: \$alias) }', variables: {
      'realm': realm,
      'alias': alias,
    }
    );
    return data['kcExportIDPPublicBrokerConfig']?.toString();
  }

  Future<List<KcIdentityProviderMapper>?> kcIdentityProviderMappers({required String realm, required String alias, String? selectionSet}) async {
    final data = await _executor.query('query kcIdentityProviderMappers(\$realm: String!, \$alias: String!) { kcIdentityProviderMappers(realm: \$realm, alias: \$alias) ${selectionSet ?? '{ id name identityProviderMapper identityProviderAlias config }'} }', variables: {
      'realm': realm,
      'alias': alias,
    }
    );
    return (data['kcIdentityProviderMappers'] as List?)?.map((e) => (e == null ? null : KcIdentityProviderMapper.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KcIdentityProviderMapper?> kcIdentityProviderMapperByID({required String realm, required String alias, required String mapperID, String? selectionSet}) async {
    final data = await _executor.query('query kcIdentityProviderMapperByID(\$realm: String!, \$alias: String!, \$mapperID: String!) { kcIdentityProviderMapperByID(realm: \$realm, alias: \$alias, mapperID: \$mapperID) ${selectionSet ?? '{ id name identityProviderMapper identityProviderAlias config }'} }', variables: {
      'realm': realm,
      'alias': alias,
      'mapperID': mapperID,
    }
    );
    return data['kcIdentityProviderMapperByID'] == null ? null : KcIdentityProviderMapper.fromJson((data['kcIdentityProviderMapperByID'] as Map).cast<String, dynamic>());
  }

  Future<Object?> kcPostRealmsProtocolToken({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsProtocolToken(\$realm: String!, \$input: Map!) { kcPostRealmsProtocolToken(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostRealmsProtocolToken'];
  }

  Future<Object?> kcPostRealmsProtocolIntrospect({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsProtocolIntrospect(\$realm: String!, \$input: Map!) { kcPostRealmsProtocolIntrospect(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostRealmsProtocolIntrospect'];
  }

  Future<Object?> kcPostRealmsProtocolRevoke({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsProtocolRevoke(\$realm: String!, \$input: Map!) { kcPostRealmsProtocolRevoke(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostRealmsProtocolRevoke'];
  }

  Future<Object?> kcPostRealmsProtocolTokenExchange({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsProtocolTokenExchange(\$realm: String!, \$input: Map!) { kcPostRealmsProtocolTokenExchange(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostRealmsProtocolTokenExchange'];
  }

  Future<Object?> kcPostRealmsProtocolLogoutPublic({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsProtocolLogoutPublic(\$realm: String!, \$input: Map!) { kcPostRealmsProtocolLogoutPublic(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostRealmsProtocolLogoutPublic'];
  }

  Future<Object?> kcPostRealmsAuthzRpt({required String realm}) async {
    final data = await _executor.mutation('mutation kcPostRealmsAuthzRpt(\$realm: String!) { kcPostRealmsAuthzRpt(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return data['kcPostRealmsAuthzRpt'];
  }

  Future<Object?> kcPostRealmsAuthzRptPermissions({required String realm}) async {
    final data = await _executor.mutation('mutation kcPostRealmsAuthzRptPermissions(\$realm: String!) { kcPostRealmsAuthzRptPermissions(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return data['kcPostRealmsAuthzRptPermissions'];
  }

  Future<Object?> kcPostRealmsAuthzRptDecision({required String realm}) async {
    final data = await _executor.mutation('mutation kcPostRealmsAuthzRptDecision(\$realm: String!) { kcPostRealmsAuthzRptDecision(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return data['kcPostRealmsAuthzRptDecision'];
  }

  Future<Object?> kcPostRealmsAuthzPermissionTickets({required String realm, required List<Map<String, dynamic>> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsAuthzPermissionTickets(\$realm: String!, \$input: [Map!]!) { kcPostRealmsAuthzPermissionTickets(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostRealmsAuthzPermissionTickets'];
  }

  Future<Object?> kcPostRealmsAuthzUmaPermissions({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsAuthzUmaPermissions(\$realm: String!, \$input: Map!) { kcPostRealmsAuthzUmaPermissions(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostRealmsAuthzUmaPermissions'];
  }

  Future<Object?> kcPutRealmsAuthzUmaPermissions({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutRealmsAuthzUmaPermissions(\$realm: String!, \$input: Map!) { kcPutRealmsAuthzUmaPermissions(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPutRealmsAuthzUmaPermissions'];
  }

  Future<Object?> kcDeleteRealmsAuthzUmaPermissions({required String realm, required String ticketID}) async {
    final data = await _executor.mutation('mutation kcDeleteRealmsAuthzUmaPermissions(\$realm: String!, \$ticketID: String!) { kcDeleteRealmsAuthzUmaPermissions(realm: \$realm, ticketID: \$ticketID) }', variables: {
      'realm': realm,
      'ticketID': ticketID,
    }
    );
    return data['kcDeleteRealmsAuthzUmaPermissions'];
  }

  Future<Object?> kcPostRealmsAuthzResources({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsAuthzResources(\$realm: String!, \$input: Map!) { kcPostRealmsAuthzResources(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostRealmsAuthzResources'];
  }

  Future<Object?> kcPutRealmsAuthzResources({required String realm, required String resourceID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutRealmsAuthzResources(\$realm: String!, \$resourceID: String!, \$input: Map!) { kcPutRealmsAuthzResources(realm: \$realm, resourceID: \$resourceID, input: \$input) }', variables: {
      'realm': realm,
      'resourceID': resourceID,
      'input': input,
    }
    );
    return data['kcPutRealmsAuthzResources'];
  }

  Future<Object?> kcDeleteRealmsAuthzResources({required String realm, required String resourceID}) async {
    final data = await _executor.mutation('mutation kcDeleteRealmsAuthzResources(\$realm: String!, \$resourceID: String!) { kcDeleteRealmsAuthzResources(realm: \$realm, resourceID: \$resourceID) }', variables: {
      'realm': realm,
      'resourceID': resourceID,
    }
    );
    return data['kcDeleteRealmsAuthzResources'];
  }

  Future<Object?> kcPostRealmsAuthzResourcesPolicies({required String realm, required String resourceID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostRealmsAuthzResourcesPolicies(\$realm: String!, \$resourceID: String!, \$input: Map!) { kcPostRealmsAuthzResourcesPolicies(realm: \$realm, resourceID: \$resourceID, input: \$input) }', variables: {
      'realm': realm,
      'resourceID': resourceID,
      'input': input,
    }
    );
    return data['kcPostRealmsAuthzResourcesPolicies'];
  }

  Future<Object?> kcPutRealmsAuthzResourcePolicies({required String realm, required String permissionID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutRealmsAuthzResourcePolicies(\$realm: String!, \$permissionID: String!, \$input: Map!) { kcPutRealmsAuthzResourcePolicies(realm: \$realm, permissionID: \$permissionID, input: \$input) }', variables: {
      'realm': realm,
      'permissionID': permissionID,
      'input': input,
    }
    );
    return data['kcPutRealmsAuthzResourcePolicies'];
  }

  Future<Object?> kcDeleteRealmsAuthzResourcePolicies({required String realm, required String permissionID}) async {
    final data = await _executor.mutation('mutation kcDeleteRealmsAuthzResourcePolicies(\$realm: String!, \$permissionID: String!) { kcDeleteRealmsAuthzResourcePolicies(realm: \$realm, permissionID: \$permissionID) }', variables: {
      'realm': realm,
      'permissionID': permissionID,
    }
    );
    return data['kcDeleteRealmsAuthzResourcePolicies'];
  }

  Future<Object?> kcPostAdminRealms({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealms(\$input: Map!) { kcPostAdminRealms(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['kcPostAdminRealms'];
  }

  Future<Object?> kcPutAdminRealms({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealms(\$realm: String!, \$input: Map!) { kcPutAdminRealms(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPutAdminRealms'];
  }

  Future<Object?> kcDeleteAdminRealms({required String realm}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealms(\$realm: String!) { kcDeleteAdminRealms(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return data['kcDeleteAdminRealms'];
  }

  Future<Object?> kcPostAdminRealmsClearRealmCache({required String realm}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClearRealmCache(\$realm: String!) { kcPostAdminRealmsClearRealmCache(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return data['kcPostAdminRealmsClearRealmCache'];
  }

  Future<Object?> kcPostAdminRealmsClearUserCache({required String realm}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClearUserCache(\$realm: String!) { kcPostAdminRealmsClearUserCache(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return data['kcPostAdminRealmsClearUserCache'];
  }

  Future<Object?> kcPostAdminRealmsClearKeysCache({required String realm}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClearKeysCache(\$realm: String!) { kcPostAdminRealmsClearKeysCache(realm: \$realm) }', variables: {
      'realm': realm,
    }
    );
    return data['kcPostAdminRealmsClearKeysCache'];
  }

  Future<Object?> kcDeleteAdminRealmsSessions({required String realm, required String sessionID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsSessions(\$realm: String!, \$sessionID: String!) { kcDeleteAdminRealmsSessions(realm: \$realm, sessionID: \$sessionID) }', variables: {
      'realm': realm,
      'sessionID': sessionID,
    }
    );
    return data['kcDeleteAdminRealmsSessions'];
  }

  Future<Object?> kcPostAdminRealmsUsers({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsUsers(\$realm: String!, \$input: Map!) { kcPostAdminRealmsUsers(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsUsers'];
  }

  Future<Object?> kcPutAdminRealmsUsers({required String realm, required String userID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsUsers(\$realm: String!, \$userID: String!, \$input: Map!) { kcPutAdminRealmsUsers(realm: \$realm, userID: \$userID, input: \$input) }', variables: {
      'realm': realm,
      'userID': userID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsUsers'];
  }

  Future<Object?> kcDeleteAdminRealmsUsers({required String realm, required String userID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsUsers(\$realm: String!, \$userID: String!) { kcDeleteAdminRealmsUsers(realm: \$realm, userID: \$userID) }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return data['kcDeleteAdminRealmsUsers'];
  }

  Future<Object?> kcPutAdminRealmsUsersResetPassword({required String realm, required String userID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsUsersResetPassword(\$realm: String!, \$userID: String!, \$input: Map!) { kcPutAdminRealmsUsersResetPassword(realm: \$realm, userID: \$userID, input: \$input) }', variables: {
      'realm': realm,
      'userID': userID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsUsersResetPassword'];
  }

  Future<Object?> kcPutAdminRealmsUsersExecuteActionsEmail({required String realm, required String userID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsUsersExecuteActionsEmail(\$realm: String!, \$userID: String!, \$input: Map!) { kcPutAdminRealmsUsersExecuteActionsEmail(realm: \$realm, userID: \$userID, input: \$input) }', variables: {
      'realm': realm,
      'userID': userID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsUsersExecuteActionsEmail'];
  }

  Future<Object?> kcPutAdminRealmsUsersSendVerifyEmail({required String realm, required String userID}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsUsersSendVerifyEmail(\$realm: String!, \$userID: String!) { kcPutAdminRealmsUsersSendVerifyEmail(realm: \$realm, userID: \$userID) }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return data['kcPutAdminRealmsUsersSendVerifyEmail'];
  }

  Future<Object?> kcDeleteAdminRealmsUsersConsents({required String realm, required String userID, required String clientID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsUsersConsents(\$realm: String!, \$userID: String!, \$clientID: String!) { kcDeleteAdminRealmsUsersConsents(realm: \$realm, userID: \$userID, clientID: \$clientID) }', variables: {
      'realm': realm,
      'userID': userID,
      'clientID': clientID,
    }
    );
    return data['kcDeleteAdminRealmsUsersConsents'];
  }

  Future<Object?> kcPostAdminRealmsUsersLogout({required String realm, required String userID}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsUsersLogout(\$realm: String!, \$userID: String!) { kcPostAdminRealmsUsersLogout(realm: \$realm, userID: \$userID) }', variables: {
      'realm': realm,
      'userID': userID,
    }
    );
    return data['kcPostAdminRealmsUsersLogout'];
  }

  Future<Object?> kcPutAdminRealmsUsersGroups({required String realm, required String userID, required String groupID}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsUsersGroups(\$realm: String!, \$userID: String!, \$groupID: String!) { kcPutAdminRealmsUsersGroups(realm: \$realm, userID: \$userID, groupID: \$groupID) }', variables: {
      'realm': realm,
      'userID': userID,
      'groupID': groupID,
    }
    );
    return data['kcPutAdminRealmsUsersGroups'];
  }

  Future<Object?> kcDeleteAdminRealmsUsersGroups({required String realm, required String userID, required String groupID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsUsersGroups(\$realm: String!, \$userID: String!, \$groupID: String!) { kcDeleteAdminRealmsUsersGroups(realm: \$realm, userID: \$userID, groupID: \$groupID) }', variables: {
      'realm': realm,
      'userID': userID,
      'groupID': groupID,
    }
    );
    return data['kcDeleteAdminRealmsUsersGroups'];
  }

  Future<Object?> kcDeleteAdminRealmsUsersCredentials({required String realm, required String userID, required String credentialID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsUsersCredentials(\$realm: String!, \$userID: String!, \$credentialID: String!) { kcDeleteAdminRealmsUsersCredentials(realm: \$realm, userID: \$userID, credentialID: \$credentialID) }', variables: {
      'realm': realm,
      'userID': userID,
      'credentialID': credentialID,
    }
    );
    return data['kcDeleteAdminRealmsUsersCredentials'];
  }

  Future<Object?> kcPutAdminRealmsUsersCredentialsUserLabel({required String realm, required String userID, required String credentialID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsUsersCredentialsUserLabel(\$realm: String!, \$userID: String!, \$credentialID: String!, \$input: Map!) { kcPutAdminRealmsUsersCredentialsUserLabel(realm: \$realm, userID: \$userID, credentialID: \$credentialID, input: \$input) }', variables: {
      'realm': realm,
      'userID': userID,
      'credentialID': credentialID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsUsersCredentialsUserLabel'];
  }

  Future<Object?> kcPostAdminRealmsUsersCredentialsMoveToFirst({required String realm, required String userID, required String credentialID}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsUsersCredentialsMoveToFirst(\$realm: String!, \$userID: String!, \$credentialID: String!) { kcPostAdminRealmsUsersCredentialsMoveToFirst(realm: \$realm, userID: \$userID, credentialID: \$credentialID) }', variables: {
      'realm': realm,
      'userID': userID,
      'credentialID': credentialID,
    }
    );
    return data['kcPostAdminRealmsUsersCredentialsMoveToFirst'];
  }

  Future<Object?> kcPostAdminRealmsUsersCredentialsMoveAfter({required String realm, required String userID, required String credentialID, required String previousCredentialID}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsUsersCredentialsMoveAfter(\$realm: String!, \$userID: String!, \$credentialID: String!, \$previousCredentialID: String!) { kcPostAdminRealmsUsersCredentialsMoveAfter(realm: \$realm, userID: \$userID, credentialID: \$credentialID, previousCredentialID: \$previousCredentialID) }', variables: {
      'realm': realm,
      'userID': userID,
      'credentialID': credentialID,
      'previousCredentialID': previousCredentialID,
    }
    );
    return data['kcPostAdminRealmsUsersCredentialsMoveAfter'];
  }

  Future<Object?> kcPutAdminRealmsUsersDisableCredentialTypes({required String realm, required String userID, required List<String> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsUsersDisableCredentialTypes(\$realm: String!, \$userID: String!, \$input: [String!]!) { kcPutAdminRealmsUsersDisableCredentialTypes(realm: \$realm, userID: \$userID, input: \$input) }', variables: {
      'realm': realm,
      'userID': userID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsUsersDisableCredentialTypes'];
  }

  Future<Object?> kcPostAdminRealmsUsersFederatedIdentity({required String realm, required String userID, required String providerID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsUsersFederatedIdentity(\$realm: String!, \$userID: String!, \$providerID: String!, \$input: Map!) { kcPostAdminRealmsUsersFederatedIdentity(realm: \$realm, userID: \$userID, providerID: \$providerID, input: \$input) }', variables: {
      'realm': realm,
      'userID': userID,
      'providerID': providerID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsUsersFederatedIdentity'];
  }

  Future<Object?> kcDeleteAdminRealmsUsersFederatedIdentity({required String realm, required String userID, required String providerID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsUsersFederatedIdentity(\$realm: String!, \$userID: String!, \$providerID: String!) { kcDeleteAdminRealmsUsersFederatedIdentity(realm: \$realm, userID: \$userID, providerID: \$providerID) }', variables: {
      'realm': realm,
      'userID': userID,
      'providerID': providerID,
    }
    );
    return data['kcDeleteAdminRealmsUsersFederatedIdentity'];
  }

  Future<Object?> kcPostAdminRealmsGroups({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsGroups(\$realm: String!, \$input: Map!) { kcPostAdminRealmsGroups(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsGroups'];
  }

  Future<Object?> kcPutAdminRealmsGroups({required String realm, required String groupID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsGroups(\$realm: String!, \$groupID: String!, \$input: Map!) { kcPutAdminRealmsGroups(realm: \$realm, groupID: \$groupID, input: \$input) }', variables: {
      'realm': realm,
      'groupID': groupID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsGroups'];
  }

  Future<Object?> kcDeleteAdminRealmsGroups({required String realm, required String groupID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsGroups(\$realm: String!, \$groupID: String!) { kcDeleteAdminRealmsGroups(realm: \$realm, groupID: \$groupID) }', variables: {
      'realm': realm,
      'groupID': groupID,
    }
    );
    return data['kcDeleteAdminRealmsGroups'];
  }

  Future<Object?> kcPostAdminRealmsGroupsChildren({required String realm, required String groupID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsGroupsChildren(\$realm: String!, \$groupID: String!, \$input: Map!) { kcPostAdminRealmsGroupsChildren(realm: \$realm, groupID: \$groupID, input: \$input) }', variables: {
      'realm': realm,
      'groupID': groupID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsGroupsChildren'];
  }

  Future<Object?> kcPutAdminRealmsDefaultGroups({required String realm, required String groupID}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsDefaultGroups(\$realm: String!, \$groupID: String!) { kcPutAdminRealmsDefaultGroups(realm: \$realm, groupID: \$groupID) }', variables: {
      'realm': realm,
      'groupID': groupID,
    }
    );
    return data['kcPutAdminRealmsDefaultGroups'];
  }

  Future<Object?> kcDeleteAdminRealmsDefaultGroups({required String realm, required String groupID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsDefaultGroups(\$realm: String!, \$groupID: String!) { kcDeleteAdminRealmsDefaultGroups(realm: \$realm, groupID: \$groupID) }', variables: {
      'realm': realm,
      'groupID': groupID,
    }
    );
    return data['kcDeleteAdminRealmsDefaultGroups'];
  }

  Future<Object?> kcPostAdminRealmsRoles({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsRoles(\$realm: String!, \$input: Map!) { kcPostAdminRealmsRoles(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsRoles'];
  }

  Future<Object?> kcPutAdminRealmsRoles({required String realm, required String roleName, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsRoles(\$realm: String!, \$roleName: String!, \$input: Map!) { kcPutAdminRealmsRoles(realm: \$realm, roleName: \$roleName, input: \$input) }', variables: {
      'realm': realm,
      'roleName': roleName,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsRoles'];
  }

  Future<Object?> kcDeleteAdminRealmsRoles({required String realm, required String roleName}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsRoles(\$realm: String!, \$roleName: String!) { kcDeleteAdminRealmsRoles(realm: \$realm, roleName: \$roleName) }', variables: {
      'realm': realm,
      'roleName': roleName,
    }
    );
    return data['kcDeleteAdminRealmsRoles'];
  }

  Future<Object?> kcPostAdminRealmsRolesComposites({required String realm, required String roleName, required List<Map<String, dynamic>> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsRolesComposites(\$realm: String!, \$roleName: String!, \$input: [Map!]!) { kcPostAdminRealmsRolesComposites(realm: \$realm, roleName: \$roleName, input: \$input) }', variables: {
      'realm': realm,
      'roleName': roleName,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsRolesComposites'];
  }

  Future<Object?> kcDeleteAdminRealmsRolesComposites({required String realm, required String roleName, required List<Map<String, dynamic>> input}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsRolesComposites(\$realm: String!, \$roleName: String!, \$input: [Map!]!) { kcDeleteAdminRealmsRolesComposites(realm: \$realm, roleName: \$roleName, input: \$input) }', variables: {
      'realm': realm,
      'roleName': roleName,
      'input': input,
    }
    );
    return data['kcDeleteAdminRealmsRolesComposites'];
  }

  Future<Object?> kcPutAdminRealmsRolesById({required String realm, required String roleID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsRolesById(\$realm: String!, \$roleID: String!, \$input: Map!) { kcPutAdminRealmsRolesById(realm: \$realm, roleID: \$roleID, input: \$input) }', variables: {
      'realm': realm,
      'roleID': roleID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsRolesById'];
  }

  Future<Object?> kcPostAdminRealmsRolesByIdComposites({required String realm, required String roleID, required List<Map<String, dynamic>> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsRolesByIdComposites(\$realm: String!, \$roleID: String!, \$input: [Map!]!) { kcPostAdminRealmsRolesByIdComposites(realm: \$realm, roleID: \$roleID, input: \$input) }', variables: {
      'realm': realm,
      'roleID': roleID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsRolesByIdComposites'];
  }

  Future<Object?> kcDeleteAdminRealmsRolesByIdComposites({required String realm, required String roleID, required List<Map<String, dynamic>> input}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsRolesByIdComposites(\$realm: String!, \$roleID: String!, \$input: [Map!]!) { kcDeleteAdminRealmsRolesByIdComposites(realm: \$realm, roleID: \$roleID, input: \$input) }', variables: {
      'realm': realm,
      'roleID': roleID,
      'input': input,
    }
    );
    return data['kcDeleteAdminRealmsRolesByIdComposites'];
  }

  Future<Object?> kcPostAdminRealmsClients({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClients(\$realm: String!, \$input: Map!) { kcPostAdminRealmsClients(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClients'];
  }

  Future<Object?> kcPutAdminRealmsClients({required String realm, required String clientID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClients(\$realm: String!, \$clientID: String!, \$input: Map!) { kcPutAdminRealmsClients(realm: \$realm, clientID: \$clientID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClients'];
  }

  Future<Object?> kcDeleteAdminRealmsClients({required String realm, required String clientID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClients(\$realm: String!, \$clientID: String!) { kcDeleteAdminRealmsClients(realm: \$realm, clientID: \$clientID) }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcDeleteAdminRealmsClients'];
  }

  Future<Object?> kcPostAdminRealmsClientsClientSecret({required String realm, required String clientID}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientsClientSecret(\$realm: String!, \$clientID: String!) { kcPostAdminRealmsClientsClientSecret(realm: \$realm, clientID: \$clientID) }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcPostAdminRealmsClientsClientSecret'];
  }

  Future<Object?> kcPutAdminRealmsClientsDefaultClientScopes({required String realm, required String clientID, required String scopeID}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsDefaultClientScopes(\$realm: String!, \$clientID: String!, \$scopeID: String!) { kcPutAdminRealmsClientsDefaultClientScopes(realm: \$realm, clientID: \$clientID, scopeID: \$scopeID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'scopeID': scopeID,
    }
    );
    return data['kcPutAdminRealmsClientsDefaultClientScopes'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsDefaultClientScopes({required String realm, required String clientID, required String scopeID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsDefaultClientScopes(\$realm: String!, \$clientID: String!, \$scopeID: String!) { kcDeleteAdminRealmsClientsDefaultClientScopes(realm: \$realm, clientID: \$clientID, scopeID: \$scopeID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'scopeID': scopeID,
    }
    );
    return data['kcDeleteAdminRealmsClientsDefaultClientScopes'];
  }

  Future<Object?> kcPutAdminRealmsClientsOptionalClientScopes({required String realm, required String clientID, required String scopeID}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsOptionalClientScopes(\$realm: String!, \$clientID: String!, \$scopeID: String!) { kcPutAdminRealmsClientsOptionalClientScopes(realm: \$realm, clientID: \$clientID, scopeID: \$scopeID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'scopeID': scopeID,
    }
    );
    return data['kcPutAdminRealmsClientsOptionalClientScopes'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsOptionalClientScopes({required String realm, required String clientID, required String scopeID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsOptionalClientScopes(\$realm: String!, \$clientID: String!, \$scopeID: String!) { kcDeleteAdminRealmsClientsOptionalClientScopes(realm: \$realm, clientID: \$clientID, scopeID: \$scopeID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'scopeID': scopeID,
    }
    );
    return data['kcDeleteAdminRealmsClientsOptionalClientScopes'];
  }

  Future<Object?> kcPostAdminRealmsClientsRoles({required String realm, required String clientID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientsRoles(\$realm: String!, \$clientID: String!, \$input: Map!) { kcPostAdminRealmsClientsRoles(realm: \$realm, clientID: \$clientID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientsRoles'];
  }

  Future<Object?> kcPutAdminRealmsClientsRoles({required String realm, required String clientID, required String roleName, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsRoles(\$realm: String!, \$clientID: String!, \$roleName: String!, \$input: Map!) { kcPutAdminRealmsClientsRoles(realm: \$realm, clientID: \$clientID, roleName: \$roleName, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'roleName': roleName,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientsRoles'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsRoles({required String realm, required String clientID, required String roleName}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsRoles(\$realm: String!, \$clientID: String!, \$roleName: String!) { kcDeleteAdminRealmsClientsRoles(realm: \$realm, clientID: \$clientID, roleName: \$roleName) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'roleName': roleName,
    }
    );
    return data['kcDeleteAdminRealmsClientsRoles'];
  }

  Future<Object?> kcPostAdminRealmsClientsProtocolMappers({required String realm, required String clientID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientsProtocolMappers(\$realm: String!, \$clientID: String!, \$input: Map!) { kcPostAdminRealmsClientsProtocolMappers(realm: \$realm, clientID: \$clientID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientsProtocolMappers'];
  }

  Future<Object?> kcPutAdminRealmsClientsProtocolMappers({required String realm, required String clientID, required String mapperID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsProtocolMappers(\$realm: String!, \$clientID: String!, \$mapperID: String!, \$input: Map!) { kcPutAdminRealmsClientsProtocolMappers(realm: \$realm, clientID: \$clientID, mapperID: \$mapperID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'mapperID': mapperID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientsProtocolMappers'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsProtocolMappers({required String realm, required String clientID, required String mapperID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsProtocolMappers(\$realm: String!, \$clientID: String!, \$mapperID: String!) { kcDeleteAdminRealmsClientsProtocolMappers(realm: \$realm, clientID: \$clientID, mapperID: \$mapperID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'mapperID': mapperID,
    }
    );
    return data['kcDeleteAdminRealmsClientsProtocolMappers'];
  }

  Future<Object?> kcPostAdminRealmsClientsAuthzResourceServerResource({required String realm, required String clientID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientsAuthzResourceServerResource(\$realm: String!, \$clientID: String!, \$input: Map!) { kcPostAdminRealmsClientsAuthzResourceServerResource(realm: \$realm, clientID: \$clientID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientsAuthzResourceServerResource'];
  }

  Future<Object?> kcPutAdminRealmsClientsAuthzResourceServerResource({required String realm, required String clientID, required String resourceID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsAuthzResourceServerResource(\$realm: String!, \$clientID: String!, \$resourceID: String!, \$input: Map!) { kcPutAdminRealmsClientsAuthzResourceServerResource(realm: \$realm, clientID: \$clientID, resourceID: \$resourceID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'resourceID': resourceID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientsAuthzResourceServerResource'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsAuthzResourceServerResource({required String realm, required String clientID, required String resourceID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsAuthzResourceServerResource(\$realm: String!, \$clientID: String!, \$resourceID: String!) { kcDeleteAdminRealmsClientsAuthzResourceServerResource(realm: \$realm, clientID: \$clientID, resourceID: \$resourceID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'resourceID': resourceID,
    }
    );
    return data['kcDeleteAdminRealmsClientsAuthzResourceServerResource'];
  }

  Future<Object?> kcPostAdminRealmsClientsAuthzResourceServerScope({required String realm, required String clientID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientsAuthzResourceServerScope(\$realm: String!, \$clientID: String!, \$input: Map!) { kcPostAdminRealmsClientsAuthzResourceServerScope(realm: \$realm, clientID: \$clientID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientsAuthzResourceServerScope'];
  }

  Future<Object?> kcPutAdminRealmsClientsAuthzResourceServerScope({required String realm, required String clientID, required String scopeID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsAuthzResourceServerScope(\$realm: String!, \$clientID: String!, \$scopeID: String!, \$input: Map!) { kcPutAdminRealmsClientsAuthzResourceServerScope(realm: \$realm, clientID: \$clientID, scopeID: \$scopeID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'scopeID': scopeID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientsAuthzResourceServerScope'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsAuthzResourceServerScope({required String realm, required String clientID, required String scopeID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsAuthzResourceServerScope(\$realm: String!, \$clientID: String!, \$scopeID: String!) { kcDeleteAdminRealmsClientsAuthzResourceServerScope(realm: \$realm, clientID: \$clientID, scopeID: \$scopeID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'scopeID': scopeID,
    }
    );
    return data['kcDeleteAdminRealmsClientsAuthzResourceServerScope'];
  }

  Future<Object?> kcPostAdminRealmsClientsAuthzResourceServerPolicy({required String realm, required String clientID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientsAuthzResourceServerPolicy(\$realm: String!, \$clientID: String!, \$input: Map!) { kcPostAdminRealmsClientsAuthzResourceServerPolicy(realm: \$realm, clientID: \$clientID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientsAuthzResourceServerPolicy'];
  }

  Future<Object?> kcPutAdminRealmsClientsAuthzResourceServerPolicy({required String realm, required String clientID, required String policyID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsAuthzResourceServerPolicy(\$realm: String!, \$clientID: String!, \$policyID: String!, \$input: Map!) { kcPutAdminRealmsClientsAuthzResourceServerPolicy(realm: \$realm, clientID: \$clientID, policyID: \$policyID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'policyID': policyID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientsAuthzResourceServerPolicy'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsAuthzResourceServerPolicy({required String realm, required String clientID, required String policyID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsAuthzResourceServerPolicy(\$realm: String!, \$clientID: String!, \$policyID: String!) { kcDeleteAdminRealmsClientsAuthzResourceServerPolicy(realm: \$realm, clientID: \$clientID, policyID: \$policyID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'policyID': policyID,
    }
    );
    return data['kcDeleteAdminRealmsClientsAuthzResourceServerPolicy'];
  }

  Future<Object?> kcPostAdminRealmsClientsAuthzResourceServerPermission({required String realm, required String clientID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientsAuthzResourceServerPermission(\$realm: String!, \$clientID: String!, \$input: Map!) { kcPostAdminRealmsClientsAuthzResourceServerPermission(realm: \$realm, clientID: \$clientID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientsAuthzResourceServerPermission'];
  }

  Future<Object?> kcPutAdminRealmsClientsAuthzResourceServerPermission({required String realm, required String clientID, required String permissionID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsAuthzResourceServerPermission(\$realm: String!, \$clientID: String!, \$permissionID: String!, \$input: Map!) { kcPutAdminRealmsClientsAuthzResourceServerPermission(realm: \$realm, clientID: \$clientID, permissionID: \$permissionID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'permissionID': permissionID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientsAuthzResourceServerPermission'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsAuthzResourceServerPermission({required String realm, required String clientID, required String permissionID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsAuthzResourceServerPermission(\$realm: String!, \$clientID: String!, \$permissionID: String!) { kcDeleteAdminRealmsClientsAuthzResourceServerPermission(realm: \$realm, clientID: \$clientID, permissionID: \$permissionID) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'permissionID': permissionID,
    }
    );
    return data['kcDeleteAdminRealmsClientsAuthzResourceServerPermission'];
  }

  Future<Object?> kcPostAdminRealmsClientsRegistrations({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientsRegistrations(\$realm: String!, \$input: Map!) { kcPostAdminRealmsClientsRegistrations(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientsRegistrations'];
  }

  Future<Object?> kcPutAdminRealmsClientsRegistrations({required String realm, required String clientID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientsRegistrations(\$realm: String!, \$clientID: String!, \$input: Map!) { kcPutAdminRealmsClientsRegistrations(realm: \$realm, clientID: \$clientID, input: \$input) }', variables: {
      'realm': realm,
      'clientID': clientID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientsRegistrations'];
  }

  Future<Object?> kcDeleteAdminRealmsClientsRegistrations({required String realm, required String clientID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientsRegistrations(\$realm: String!, \$clientID: String!) { kcDeleteAdminRealmsClientsRegistrations(realm: \$realm, clientID: \$clientID) }', variables: {
      'realm': realm,
      'clientID': clientID,
    }
    );
    return data['kcDeleteAdminRealmsClientsRegistrations'];
  }

  Future<Object?> kcPostAdminRealmsClientScopes({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientScopes(\$realm: String!, \$input: Map!) { kcPostAdminRealmsClientScopes(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientScopes'];
  }

  Future<Object?> kcPutAdminRealmsClientScopes({required String realm, required String scopeID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientScopes(\$realm: String!, \$scopeID: String!, \$input: Map!) { kcPutAdminRealmsClientScopes(realm: \$realm, scopeID: \$scopeID, input: \$input) }', variables: {
      'realm': realm,
      'scopeID': scopeID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientScopes'];
  }

  Future<Object?> kcDeleteAdminRealmsClientScopes({required String realm, required String scopeID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientScopes(\$realm: String!, \$scopeID: String!) { kcDeleteAdminRealmsClientScopes(realm: \$realm, scopeID: \$scopeID) }', variables: {
      'realm': realm,
      'scopeID': scopeID,
    }
    );
    return data['kcDeleteAdminRealmsClientScopes'];
  }

  Future<Object?> kcPostAdminRealmsClientScopesProtocolMappers({required String realm, required String scopeID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsClientScopesProtocolMappers(\$realm: String!, \$scopeID: String!, \$input: Map!) { kcPostAdminRealmsClientScopesProtocolMappers(realm: \$realm, scopeID: \$scopeID, input: \$input) }', variables: {
      'realm': realm,
      'scopeID': scopeID,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsClientScopesProtocolMappers'];
  }

  Future<Object?> kcPutAdminRealmsClientScopesProtocolMappers({required String realm, required String scopeID, required String mapperID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsClientScopesProtocolMappers(\$realm: String!, \$scopeID: String!, \$mapperID: String!, \$input: Map!) { kcPutAdminRealmsClientScopesProtocolMappers(realm: \$realm, scopeID: \$scopeID, mapperID: \$mapperID, input: \$input) }', variables: {
      'realm': realm,
      'scopeID': scopeID,
      'mapperID': mapperID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsClientScopesProtocolMappers'];
  }

  Future<Object?> kcDeleteAdminRealmsClientScopesProtocolMappers({required String realm, required String scopeID, required String mapperID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsClientScopesProtocolMappers(\$realm: String!, \$scopeID: String!, \$mapperID: String!) { kcDeleteAdminRealmsClientScopesProtocolMappers(realm: \$realm, scopeID: \$scopeID, mapperID: \$mapperID) }', variables: {
      'realm': realm,
      'scopeID': scopeID,
      'mapperID': mapperID,
    }
    );
    return data['kcDeleteAdminRealmsClientScopesProtocolMappers'];
  }

  Future<Object?> kcPostAdminRealmsComponents({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsComponents(\$realm: String!, \$input: Map!) { kcPostAdminRealmsComponents(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsComponents'];
  }

  Future<Object?> kcPutAdminRealmsComponents({required String realm, required String componentID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsComponents(\$realm: String!, \$componentID: String!, \$input: Map!) { kcPutAdminRealmsComponents(realm: \$realm, componentID: \$componentID, input: \$input) }', variables: {
      'realm': realm,
      'componentID': componentID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsComponents'];
  }

  Future<Object?> kcDeleteAdminRealmsComponents({required String realm, required String componentID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsComponents(\$realm: String!, \$componentID: String!) { kcDeleteAdminRealmsComponents(realm: \$realm, componentID: \$componentID) }', variables: {
      'realm': realm,
      'componentID': componentID,
    }
    );
    return data['kcDeleteAdminRealmsComponents'];
  }

  Future<Object?> kcPostAdminRealmsAuthenticationFlows({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsAuthenticationFlows(\$realm: String!, \$input: Map!) { kcPostAdminRealmsAuthenticationFlows(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsAuthenticationFlows'];
  }

  Future<Object?> kcPutAdminRealmsAuthenticationFlows({required String realm, required String flow, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsAuthenticationFlows(\$realm: String!, \$flow: String!, \$input: Map!) { kcPutAdminRealmsAuthenticationFlows(realm: \$realm, flow: \$flow, input: \$input) }', variables: {
      'realm': realm,
      'flow': flow,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsAuthenticationFlows'];
  }

  Future<Object?> kcDeleteAdminRealmsAuthenticationFlows({required String realm, required String flow}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsAuthenticationFlows(\$realm: String!, \$flow: String!) { kcDeleteAdminRealmsAuthenticationFlows(realm: \$realm, flow: \$flow) }', variables: {
      'realm': realm,
      'flow': flow,
    }
    );
    return data['kcDeleteAdminRealmsAuthenticationFlows'];
  }

  Future<Object?> kcPutAdminRealmsAuthenticationFlowsExecutions({required String realm, required String flow, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsAuthenticationFlowsExecutions(\$realm: String!, \$flow: String!, \$input: Map!) { kcPutAdminRealmsAuthenticationFlowsExecutions(realm: \$realm, flow: \$flow, input: \$input) }', variables: {
      'realm': realm,
      'flow': flow,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsAuthenticationFlowsExecutions'];
  }

  Future<Object?> kcPostAdminRealmsAuthenticationFlowsExecutionsExecution({required String realm, required String flow, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsAuthenticationFlowsExecutionsExecution(\$realm: String!, \$flow: String!, \$input: Map!) { kcPostAdminRealmsAuthenticationFlowsExecutionsExecution(realm: \$realm, flow: \$flow, input: \$input) }', variables: {
      'realm': realm,
      'flow': flow,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsAuthenticationFlowsExecutionsExecution'];
  }

  Future<Object?> kcPostAdminRealmsAuthenticationFlowsExecutionsFlow({required String realm, required String flow, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsAuthenticationFlowsExecutionsFlow(\$realm: String!, \$flow: String!, \$input: Map!) { kcPostAdminRealmsAuthenticationFlowsExecutionsFlow(realm: \$realm, flow: \$flow, input: \$input) }', variables: {
      'realm': realm,
      'flow': flow,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsAuthenticationFlowsExecutionsFlow'];
  }

  Future<Object?> kcDeleteAdminRealmsAuthenticationExecutions({required String realm, required String executionID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsAuthenticationExecutions(\$realm: String!, \$executionID: String!) { kcDeleteAdminRealmsAuthenticationExecutions(realm: \$realm, executionID: \$executionID) }', variables: {
      'realm': realm,
      'executionID': executionID,
    }
    );
    return data['kcDeleteAdminRealmsAuthenticationExecutions'];
  }

  Future<Object?> kcPostAdminRealmsAuthenticationRegisterRequiredAction({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsAuthenticationRegisterRequiredAction(\$realm: String!, \$input: Map!) { kcPostAdminRealmsAuthenticationRegisterRequiredAction(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsAuthenticationRegisterRequiredAction'];
  }

  Future<Object?> kcPutAdminRealmsAuthenticationRequiredActions({required String realm, required String alias, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsAuthenticationRequiredActions(\$realm: String!, \$alias: String!, \$input: Map!) { kcPutAdminRealmsAuthenticationRequiredActions(realm: \$realm, alias: \$alias, input: \$input) }', variables: {
      'realm': realm,
      'alias': alias,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsAuthenticationRequiredActions'];
  }

  Future<Object?> kcDeleteAdminRealmsAuthenticationRequiredActions({required String realm, required String alias}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsAuthenticationRequiredActions(\$realm: String!, \$alias: String!) { kcDeleteAdminRealmsAuthenticationRequiredActions(realm: \$realm, alias: \$alias) }', variables: {
      'realm': realm,
      'alias': alias,
    }
    );
    return data['kcDeleteAdminRealmsAuthenticationRequiredActions'];
  }

  Future<Object?> kcPostAdminRealmsIdentityProviders({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsIdentityProviders(\$realm: String!, \$input: Map!) { kcPostAdminRealmsIdentityProviders(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsIdentityProviders'];
  }

  Future<Object?> kcPostAdminRealmsIdentityProvidersImportConfig({required String realm, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsIdentityProvidersImportConfig(\$realm: String!, \$input: Map!) { kcPostAdminRealmsIdentityProvidersImportConfig(realm: \$realm, input: \$input) }', variables: {
      'realm': realm,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsIdentityProvidersImportConfig'];
  }

  Future<Object?> kcPutAdminRealmsIdentityProviders({required String realm, required String alias, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsIdentityProviders(\$realm: String!, \$alias: String!, \$input: Map!) { kcPutAdminRealmsIdentityProviders(realm: \$realm, alias: \$alias, input: \$input) }', variables: {
      'realm': realm,
      'alias': alias,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsIdentityProviders'];
  }

  Future<Object?> kcDeleteAdminRealmsIdentityProviders({required String realm, required String alias}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsIdentityProviders(\$realm: String!, \$alias: String!) { kcDeleteAdminRealmsIdentityProviders(realm: \$realm, alias: \$alias) }', variables: {
      'realm': realm,
      'alias': alias,
    }
    );
    return data['kcDeleteAdminRealmsIdentityProviders'];
  }

  Future<Object?> kcPostAdminRealmsIdentityProvidersMappers({required String realm, required String alias, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPostAdminRealmsIdentityProvidersMappers(\$realm: String!, \$alias: String!, \$input: Map!) { kcPostAdminRealmsIdentityProvidersMappers(realm: \$realm, alias: \$alias, input: \$input) }', variables: {
      'realm': realm,
      'alias': alias,
      'input': input,
    }
    );
    return data['kcPostAdminRealmsIdentityProvidersMappers'];
  }

  Future<Object?> kcPutAdminRealmsIdentityProvidersMappers({required String realm, required String alias, required String mapperID, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation kcPutAdminRealmsIdentityProvidersMappers(\$realm: String!, \$alias: String!, \$mapperID: String!, \$input: Map!) { kcPutAdminRealmsIdentityProvidersMappers(realm: \$realm, alias: \$alias, mapperID: \$mapperID, input: \$input) }', variables: {
      'realm': realm,
      'alias': alias,
      'mapperID': mapperID,
      'input': input,
    }
    );
    return data['kcPutAdminRealmsIdentityProvidersMappers'];
  }

  Future<Object?> kcDeleteAdminRealmsIdentityProvidersMappers({required String realm, required String alias, required String mapperID}) async {
    final data = await _executor.mutation('mutation kcDeleteAdminRealmsIdentityProvidersMappers(\$realm: String!, \$alias: String!, \$mapperID: String!) { kcDeleteAdminRealmsIdentityProvidersMappers(realm: \$realm, alias: \$alias, mapperID: \$mapperID) }', variables: {
      'realm': realm,
      'alias': alias,
      'mapperID': mapperID,
    }
    );
    return data['kcDeleteAdminRealmsIdentityProvidersMappers'];
  }

  Future<Object?> kcPostLoginAdmin({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation kcPostLoginAdmin(\$input: Map) { kcPostLoginAdmin(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostLoginAdmin'];
  }

  Future<Object?> kcPostUser({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation kcPostUser(\$input: Map) { kcPostUser(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostUser'];
  }

  Future<Object?> kcPostLogin({Map<String, dynamic>? input}) async {
    final data = await _executor.public('mutation kcPostLogin(\$input: Map) { kcPostLogin(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostLogin'];
  }

  Future<Object?> kcPostLoginOtp({Map<String, dynamic>? input}) async {
    final data = await _executor.public('mutation kcPostLoginOtp(\$input: Map) { kcPostLoginOtp(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostLoginOtp'];
  }

  Future<Object?> kcPostLogout({Map<String, dynamic>? input}) async {
    final data = await _executor.public('mutation kcPostLogout(\$input: Map) { kcPostLogout(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostLogout'];
  }

  Future<Object?> kcPostLoginClient({Map<String, dynamic>? input}) async {
    final data = await _executor.public('mutation kcPostLoginClient(\$input: Map) { kcPostLoginClient(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostLoginClient'];
  }

  Future<Object?> kcPostRefreshToken({Map<String, dynamic>? input}) async {
    final data = await _executor.public('mutation kcPostRefreshToken(\$input: Map) { kcPostRefreshToken(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostRefreshToken'];
  }

  Future<Object?> kcPostUserInfo({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation kcPostUserInfo(\$input: Map) { kcPostUserInfo(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostUserInfo'];
  }

  Future<Object?> kcPostSetPassword({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation kcPostSetPassword(\$input: Map) { kcPostSetPassword(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['kcPostSetPassword'];
  }

}
