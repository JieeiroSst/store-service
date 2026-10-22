// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class CrmApi {
  const CrmApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<CrmCampaign>?> crmCampaigns({int? offset, int? limit, String? q, String? selectionSet}) async {
    final data = await _executor.query('query crmCampaigns(\$offset: Int, \$limit: Int, \$q: String) { crmCampaigns(offset: \$offset, limit: \$limit, q: \$q) ${selectionSet ?? '{ id created_at updated_at campaign_name campaign_objectives campaign_sponsor campaign_start_date campaign_end_date campaign_other_details }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
    }
    );
    return (data['crmCampaigns'] as List?)?.map((e) => (e == null ? null : CrmCampaign.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmCampaign?> crmCampaign({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmCampaign(\$id: Int!) { crmCampaign(id: \$id) ${selectionSet ?? '{ id created_at updated_at campaign_name campaign_objectives campaign_sponsor campaign_start_date campaign_end_date campaign_other_details }'} }', variables: {
      'id': id,
    }
    );
    return data['crmCampaign'] == null ? null : CrmCampaign.fromJson((data['crmCampaign'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmLead>?> crmLeads({int? offset, int? limit, String? q, String? status, String? source, String? selectionSet}) async {
    final data = await _executor.query('query crmLeads(\$offset: Int, \$limit: Int, \$q: String, \$status: String, \$source: String) { crmLeads(offset: \$offset, limit: \$limit, q: \$q, status: \$status, source: \$source) ${selectionSet ?? '{ id created_at updated_at lead_firstname lead_surname email phone source lead_other_details status converted_account_id converted_contact_id }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
      'status': ?status,
      'source': ?source,
    }
    );
    return (data['crmLeads'] as List?)?.map((e) => (e == null ? null : CrmLead.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmLead?> crmLead({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmLead(\$id: Int!) { crmLead(id: \$id) ${selectionSet ?? '{ id created_at updated_at lead_firstname lead_surname email phone source lead_other_details status converted_account_id converted_contact_id }'} }', variables: {
      'id': id,
    }
    );
    return data['crmLead'] == null ? null : CrmLead.fromJson((data['crmLead'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmAccount>?> crmAccounts({int? offset, int? limit, String? q, String? selectionSet}) async {
    final data = await _executor.query('query crmAccounts(\$offset: Int, \$limit: Int, \$q: String) { crmAccounts(offset: \$offset, limit: \$limit, q: \$q) ${selectionSet ?? '{ id created_at updated_at account_name account_description account_phone billing_address email tax_code representative representative_title }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
    }
    );
    return (data['crmAccounts'] as List?)?.map((e) => (e == null ? null : CrmAccount.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmAccount?> crmAccount({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmAccount(\$id: Int!) { crmAccount(id: \$id) ${selectionSet ?? '{ id created_at updated_at account_name account_description account_phone billing_address email tax_code representative representative_title }'} }', variables: {
      'id': id,
    }
    );
    return data['crmAccount'] == null ? null : CrmAccount.fromJson((data['crmAccount'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmContact>?> crmContacts({int? offset, int? limit, String? q, int? accountId, String? selectionSet}) async {
    final data = await _executor.query('query crmContacts(\$offset: Int, \$limit: Int, \$q: String, \$account_id: Int) { crmContacts(offset: \$offset, limit: \$limit, q: \$q, account_id: \$account_id) ${selectionSet ?? '{ id created_at updated_at account_id name email phone contact_address contact_contact_details }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
      'account_id': ?accountId,
    }
    );
    return (data['crmContacts'] as List?)?.map((e) => (e == null ? null : CrmContact.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmContact?> crmContact({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmContact(\$id: Int!) { crmContact(id: \$id) ${selectionSet ?? '{ id created_at updated_at account_id name email phone contact_address contact_contact_details }'} }', variables: {
      'id': id,
    }
    );
    return data['crmContact'] == null ? null : CrmContact.fromJson((data['crmContact'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmCampaignMember>?> crmCampaignMembers({int? offset, int? limit, String? q, int? campaignId, int? leadId, int? contactId, String? selectionSet}) async {
    final data = await _executor.query('query crmCampaignMembers(\$offset: Int, \$limit: Int, \$q: String, \$campaign_id: Int, \$lead_id: Int, \$contact_id: Int) { crmCampaignMembers(offset: \$offset, limit: \$limit, q: \$q, campaign_id: \$campaign_id, lead_id: \$lead_id, contact_id: \$contact_id) ${selectionSet ?? '{ id created_at updated_at campaign_id lead_id contact_id }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
      'campaign_id': ?campaignId,
      'lead_id': ?leadId,
      'contact_id': ?contactId,
    }
    );
    return (data['crmCampaignMembers'] as List?)?.map((e) => (e == null ? null : CrmCampaignMember.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmCampaignMember?> crmCampaignMember({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmCampaignMember(\$id: Int!) { crmCampaignMember(id: \$id) ${selectionSet ?? '{ id created_at updated_at campaign_id lead_id contact_id }'} }', variables: {
      'id': id,
    }
    );
    return data['crmCampaignMember'] == null ? null : CrmCampaignMember.fromJson((data['crmCampaignMember'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmCase>?> crmCases({int? offset, int? limit, String? q, int? contactId, String? status, String? priority, String? selectionSet}) async {
    final data = await _executor.query('query crmCases(\$offset: Int, \$limit: Int, \$q: String, \$contact_id: Int, \$status: String, \$priority: String) { crmCases(offset: \$offset, limit: \$limit, q: \$q, contact_id: \$contact_id, status: \$status, priority: \$priority) ${selectionSet ?? '{ id created_at updated_at contact_id subject description priority status closed_at }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
      'contact_id': ?contactId,
      'status': ?status,
      'priority': ?priority,
    }
    );
    return (data['crmCases'] as List?)?.map((e) => (e == null ? null : CrmCase.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmCase?> crmCase({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmCase(\$id: Int!) { crmCase(id: \$id) ${selectionSet ?? '{ id created_at updated_at contact_id subject description priority status closed_at }'} }', variables: {
      'id': id,
    }
    );
    return data['crmCase'] == null ? null : CrmCase.fromJson((data['crmCase'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmContract>?> crmContracts({int? offset, int? limit, String? q, int? accountId, String? contractStatus, String? selectionSet}) async {
    final data = await _executor.query('query crmContracts(\$offset: Int, \$limit: Int, \$q: String, \$account_id: Int, \$contract_status: String) { crmContracts(offset: \$offset, limit: \$limit, q: \$q, account_id: \$account_id, contract_status: \$contract_status) ${selectionSet ?? '{ id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
      'account_id': ?accountId,
      'contract_status': ?contractStatus,
    }
    );
    return (data['crmContracts'] as List?)?.map((e) => (e == null ? null : CrmContract.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmContract?> crmContract({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmContract(\$id: Int!) { crmContract(id: \$id) ${selectionSet ?? '{ id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority }'} }', variables: {
      'id': id,
    }
    );
    return data['crmContract'] == null ? null : CrmContract.fromJson((data['crmContract'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmAccountContactRole>?> crmAccountContactRoles({int? offset, int? limit, String? q, int? contactId, int? accountId, String? selectionSet}) async {
    final data = await _executor.query('query crmAccountContactRoles(\$offset: Int, \$limit: Int, \$q: String, \$contact_id: Int, \$account_id: Int) { crmAccountContactRoles(offset: \$offset, limit: \$limit, q: \$q, contact_id: \$contact_id, account_id: \$account_id) ${selectionSet ?? '{ id created_at updated_at contact_id account_id }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
      'contact_id': ?contactId,
      'account_id': ?accountId,
    }
    );
    return (data['crmAccountContactRoles'] as List?)?.map((e) => (e == null ? null : CrmAccountContactRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmAccountContactRole?> crmAccountContactRole({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmAccountContactRole(\$id: Int!) { crmAccountContactRole(id: \$id) ${selectionSet ?? '{ id created_at updated_at contact_id account_id }'} }', variables: {
      'id': id,
    }
    );
    return data['crmAccountContactRole'] == null ? null : CrmAccountContactRole.fromJson((data['crmAccountContactRole'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmOpportunity>?> crmOpportunities({int? offset, int? limit, String? q, int? accountId, String? opportunityStage, String? selectionSet}) async {
    final data = await _executor.query('query crmOpportunities(\$offset: Int, \$limit: Int, \$q: String, \$account_id: Int, \$opportunity_stage: String) { crmOpportunities(offset: \$offset, limit: \$limit, q: \$q, account_id: \$account_id, opportunity_stage: \$opportunity_stage) ${selectionSet ?? '{ id created_at updated_at account_id opportunity_description opportunity_details amount close_date opportunity_stage }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
      'account_id': ?accountId,
      'opportunity_stage': ?opportunityStage,
    }
    );
    return (data['crmOpportunities'] as List?)?.map((e) => (e == null ? null : CrmOpportunity.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmOpportunity?> crmOpportunity({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmOpportunity(\$id: Int!) { crmOpportunity(id: \$id) ${selectionSet ?? '{ id created_at updated_at account_id opportunity_description opportunity_details amount close_date opportunity_stage }'} }', variables: {
      'id': id,
    }
    );
    return data['crmOpportunity'] == null ? null : CrmOpportunity.fromJson((data['crmOpportunity'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmOpportunityContactRole>?> crmOpportunityContactRoles({int? offset, int? limit, String? q, int? contactId, int? opportunityId, String? selectionSet}) async {
    final data = await _executor.query('query crmOpportunityContactRoles(\$offset: Int, \$limit: Int, \$q: String, \$contact_id: Int, \$opportunity_id: Int) { crmOpportunityContactRoles(offset: \$offset, limit: \$limit, q: \$q, contact_id: \$contact_id, opportunity_id: \$opportunity_id) ${selectionSet ?? '{ id created_at updated_at contact_id opportunity_id date_time other_details }'} }', variables: {
      'offset': ?offset,
      'limit': ?limit,
      'q': ?q,
      'contact_id': ?contactId,
      'opportunity_id': ?opportunityId,
    }
    );
    return (data['crmOpportunityContactRoles'] as List?)?.map((e) => (e == null ? null : CrmOpportunityContactRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmOpportunityContactRole?> crmOpportunityContactRole({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmOpportunityContactRole(\$id: Int!) { crmOpportunityContactRole(id: \$id) ${selectionSet ?? '{ id created_at updated_at contact_id opportunity_id date_time other_details }'} }', variables: {
      'id': id,
    }
    );
    return data['crmOpportunityContactRole'] == null ? null : CrmOpportunityContactRole.fromJson((data['crmOpportunityContactRole'] as Map).cast<String, dynamic>());
  }

  Future<Object?> crmContractSigningPayload({required int id, required String endDate, required String signedBy, String? signingTime}) async {
    final data = await _executor.query('query crmContractSigningPayload(\$id: Int!, \$end_date: String!, \$signed_by: String!, \$signing_time: String) { crmContractSigningPayload(id: \$id, end_date: \$end_date, signed_by: \$signed_by, signing_time: \$signing_time) }', variables: {
      'id': id,
      'end_date': endDate,
      'signed_by': signedBy,
      'signing_time': ?signingTime,
    }
    );
    return data['crmContractSigningPayload'];
  }

  Future<CrmAccountOverview?> crmAccountOverview({required int id, String? selectionSet}) async {
    final data = await _executor.query('query crmAccountOverview(\$id: Int!) { crmAccountOverview(id: \$id) ${selectionSet ?? '{ account { id created_at updated_at account_name account_description account_phone billing_address email tax_code representative representative_title } contacts { id created_at updated_at account_id name email phone contact_address contact_contact_details } contracts { id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority } opportunities { id created_at updated_at account_id opportunity_description opportunity_details amount close_date opportunity_stage } }'} }', variables: {
      'id': id,
    }
    );
    return data['crmAccountOverview'] == null ? null : CrmAccountOverview.fromJson((data['crmAccountOverview'] as Map).cast<String, dynamic>());
  }

  Future<CrmSummary?> crmSummary({String? selectionSet}) async {
    final data = await _executor.query('query crmSummary { crmSummary ${selectionSet ?? '{ counts leads_by_status open_cases pipeline { stage count amount } }'} }'
    );
    return data['crmSummary'] == null ? null : CrmSummary.fromJson((data['crmSummary'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmFileKind>?> crmContractFileKinds({String? selectionSet}) async {
    final data = await _executor.query('query crmContractFileKinds { crmContractFileKinds ${selectionSet ?? '{ kind label formats system }'} }'
    );
    return (data['crmContractFileKinds'] as List?)?.map((e) => (e == null ? null : CrmFileKind.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<CrmContractFile>?> crmContractFiles({required int id, String? kind, bool? allVersions, int? offset, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query crmContractFiles(\$id: Int!, \$kind: String, \$all_versions: Boolean, \$offset: Int, \$limit: Int) { crmContractFiles(id: \$id, kind: \$kind, all_versions: \$all_versions, offset: \$offset, limit: \$limit) ${selectionSet ?? '{ id created_at updated_at contract_id kind name content_type format size sha256 description uploaded_by uploader_id source lineage_id version latest replaces_id scan_status scan_engine scanned_at }'} }', variables: {
      'id': id,
      'kind': ?kind,
      'all_versions': ?allVersions,
      'offset': ?offset,
      'limit': ?limit,
    }
    );
    return (data['crmContractFiles'] as List?)?.map((e) => (e == null ? null : CrmContractFile.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmContractFile?> crmContractFile({required int id, required int fileId, String? selectionSet}) async {
    final data = await _executor.query('query crmContractFile(\$id: Int!, \$file_id: Int!) { crmContractFile(id: \$id, file_id: \$file_id) ${selectionSet ?? '{ id created_at updated_at contract_id kind name content_type format size sha256 description uploaded_by uploader_id source lineage_id version latest replaces_id scan_status scan_engine scanned_at }'} }', variables: {
      'id': id,
      'file_id': fileId,
    }
    );
    return data['crmContractFile'] == null ? null : CrmContractFile.fromJson((data['crmContractFile'] as Map).cast<String, dynamic>());
  }

  Future<CrmFileHistory?> crmContractFileHistory({required int id, required int fileId, String? selectionSet}) async {
    final data = await _executor.query('query crmContractFileHistory(\$id: Int!, \$file_id: Int!) { crmContractFileHistory(id: \$id, file_id: \$file_id) ${selectionSet ?? '{ versions { id created_at updated_at contract_id kind name content_type format size sha256 description uploaded_by uploader_id source lineage_id version latest replaces_id scan_status scan_engine scanned_at } events { id at contract_id file_id lineage_id version file_name kind action actor_id actor_name actor_roles actor_service ip detail } }'} }', variables: {
      'id': id,
      'file_id': fileId,
    }
    );
    return data['crmContractFileHistory'] == null ? null : CrmFileHistory.fromJson((data['crmContractFileHistory'] as Map).cast<String, dynamic>());
  }

  Future<List<CrmContractFileEvent>?> crmContractFileEvents({required int id, String? action, int? fileId, int? offset, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query crmContractFileEvents(\$id: Int!, \$action: String, \$file_id: Int, \$offset: Int, \$limit: Int) { crmContractFileEvents(id: \$id, action: \$action, file_id: \$file_id, offset: \$offset, limit: \$limit) ${selectionSet ?? '{ id at contract_id file_id lineage_id version file_name kind action actor_id actor_name actor_roles actor_service ip detail }'} }', variables: {
      'id': id,
      'action': ?action,
      'file_id': ?fileId,
      'offset': ?offset,
      'limit': ?limit,
    }
    );
    return (data['crmContractFileEvents'] as List?)?.map((e) => (e == null ? null : CrmContractFileEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CrmCampaign?> crmCreateCampaign({required CrmCampaignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateCampaign(\$input: CrmCampaignInput!) { crmCreateCampaign(input: \$input) ${selectionSet ?? '{ id created_at updated_at campaign_name campaign_objectives campaign_sponsor campaign_start_date campaign_end_date campaign_other_details }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateCampaign'] == null ? null : CrmCampaign.fromJson((data['crmCreateCampaign'] as Map).cast<String, dynamic>());
  }

  Future<CrmCampaign?> crmUpdateCampaign({required int id, required CrmCampaignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateCampaign(\$id: Int!, \$input: CrmCampaignInput!) { crmUpdateCampaign(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at campaign_name campaign_objectives campaign_sponsor campaign_start_date campaign_end_date campaign_other_details }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateCampaign'] == null ? null : CrmCampaign.fromJson((data['crmUpdateCampaign'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteCampaign({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteCampaign(\$id: Int!) { crmDeleteCampaign(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteCampaign'] as bool?;
  }

  Future<CrmLead?> crmCreateLead({required CrmLeadInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateLead(\$input: CrmLeadInput!) { crmCreateLead(input: \$input) ${selectionSet ?? '{ id created_at updated_at lead_firstname lead_surname email phone source lead_other_details status converted_account_id converted_contact_id }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateLead'] == null ? null : CrmLead.fromJson((data['crmCreateLead'] as Map).cast<String, dynamic>());
  }

  Future<CrmLead?> crmUpdateLead({required int id, required CrmLeadInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateLead(\$id: Int!, \$input: CrmLeadInput!) { crmUpdateLead(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at lead_firstname lead_surname email phone source lead_other_details status converted_account_id converted_contact_id }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateLead'] == null ? null : CrmLead.fromJson((data['crmUpdateLead'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteLead({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteLead(\$id: Int!) { crmDeleteLead(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteLead'] as bool?;
  }

  Future<CrmAccount?> crmCreateAccount({required CrmAccountInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateAccount(\$input: CrmAccountInput!) { crmCreateAccount(input: \$input) ${selectionSet ?? '{ id created_at updated_at account_name account_description account_phone billing_address email tax_code representative representative_title }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateAccount'] == null ? null : CrmAccount.fromJson((data['crmCreateAccount'] as Map).cast<String, dynamic>());
  }

  Future<CrmAccount?> crmUpdateAccount({required int id, required CrmAccountInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateAccount(\$id: Int!, \$input: CrmAccountInput!) { crmUpdateAccount(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at account_name account_description account_phone billing_address email tax_code representative representative_title }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateAccount'] == null ? null : CrmAccount.fromJson((data['crmUpdateAccount'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteAccount({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteAccount(\$id: Int!) { crmDeleteAccount(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteAccount'] as bool?;
  }

  Future<CrmContact?> crmCreateContact({required CrmContactInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateContact(\$input: CrmContactInput!) { crmCreateContact(input: \$input) ${selectionSet ?? '{ id created_at updated_at account_id name email phone contact_address contact_contact_details }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateContact'] == null ? null : CrmContact.fromJson((data['crmCreateContact'] as Map).cast<String, dynamic>());
  }

  Future<CrmContact?> crmUpdateContact({required int id, required CrmContactInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateContact(\$id: Int!, \$input: CrmContactInput!) { crmUpdateContact(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at account_id name email phone contact_address contact_contact_details }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateContact'] == null ? null : CrmContact.fromJson((data['crmUpdateContact'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteContact({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteContact(\$id: Int!) { crmDeleteContact(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteContact'] as bool?;
  }

  Future<CrmCampaignMember?> crmCreateCampaignMember({required CrmCampaignMemberInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateCampaignMember(\$input: CrmCampaignMemberInput!) { crmCreateCampaignMember(input: \$input) ${selectionSet ?? '{ id created_at updated_at campaign_id lead_id contact_id }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateCampaignMember'] == null ? null : CrmCampaignMember.fromJson((data['crmCreateCampaignMember'] as Map).cast<String, dynamic>());
  }

  Future<CrmCampaignMember?> crmUpdateCampaignMember({required int id, required CrmCampaignMemberInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateCampaignMember(\$id: Int!, \$input: CrmCampaignMemberInput!) { crmUpdateCampaignMember(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at campaign_id lead_id contact_id }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateCampaignMember'] == null ? null : CrmCampaignMember.fromJson((data['crmUpdateCampaignMember'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteCampaignMember({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteCampaignMember(\$id: Int!) { crmDeleteCampaignMember(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteCampaignMember'] as bool?;
  }

  Future<CrmCase?> crmCreateCase({required CrmCaseInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateCase(\$input: CrmCaseInput!) { crmCreateCase(input: \$input) ${selectionSet ?? '{ id created_at updated_at contact_id subject description priority status closed_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateCase'] == null ? null : CrmCase.fromJson((data['crmCreateCase'] as Map).cast<String, dynamic>());
  }

  Future<CrmCase?> crmUpdateCase({required int id, required CrmCaseInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateCase(\$id: Int!, \$input: CrmCaseInput!) { crmUpdateCase(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at contact_id subject description priority status closed_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateCase'] == null ? null : CrmCase.fromJson((data['crmUpdateCase'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteCase({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteCase(\$id: Int!) { crmDeleteCase(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteCase'] as bool?;
  }

  Future<CrmContract?> crmCreateContract({required CrmContractInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateContract(\$input: CrmContractInput!) { crmCreateContract(input: \$input) ${selectionSet ?? '{ id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateContract'] == null ? null : CrmContract.fromJson((data['crmCreateContract'] as Map).cast<String, dynamic>());
  }

  Future<CrmContract?> crmUpdateContract({required int id, required CrmContractInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateContract(\$id: Int!, \$input: CrmContractInput!) { crmUpdateContract(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateContract'] == null ? null : CrmContract.fromJson((data['crmUpdateContract'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteContract({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteContract(\$id: Int!) { crmDeleteContract(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteContract'] as bool?;
  }

  Future<CrmAccountContactRole?> crmCreateAccountContactRole({required CrmAccountContactRoleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateAccountContactRole(\$input: CrmAccountContactRoleInput!) { crmCreateAccountContactRole(input: \$input) ${selectionSet ?? '{ id created_at updated_at contact_id account_id }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateAccountContactRole'] == null ? null : CrmAccountContactRole.fromJson((data['crmCreateAccountContactRole'] as Map).cast<String, dynamic>());
  }

  Future<CrmAccountContactRole?> crmUpdateAccountContactRole({required int id, required CrmAccountContactRoleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateAccountContactRole(\$id: Int!, \$input: CrmAccountContactRoleInput!) { crmUpdateAccountContactRole(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at contact_id account_id }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateAccountContactRole'] == null ? null : CrmAccountContactRole.fromJson((data['crmUpdateAccountContactRole'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteAccountContactRole({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteAccountContactRole(\$id: Int!) { crmDeleteAccountContactRole(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteAccountContactRole'] as bool?;
  }

  Future<CrmOpportunity?> crmCreateOpportunity({required CrmOpportunityInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateOpportunity(\$input: CrmOpportunityInput!) { crmCreateOpportunity(input: \$input) ${selectionSet ?? '{ id created_at updated_at account_id opportunity_description opportunity_details amount close_date opportunity_stage }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateOpportunity'] == null ? null : CrmOpportunity.fromJson((data['crmCreateOpportunity'] as Map).cast<String, dynamic>());
  }

  Future<CrmOpportunity?> crmUpdateOpportunity({required int id, required CrmOpportunityInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateOpportunity(\$id: Int!, \$input: CrmOpportunityInput!) { crmUpdateOpportunity(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at account_id opportunity_description opportunity_details amount close_date opportunity_stage }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateOpportunity'] == null ? null : CrmOpportunity.fromJson((data['crmUpdateOpportunity'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteOpportunity({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteOpportunity(\$id: Int!) { crmDeleteOpportunity(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteOpportunity'] as bool?;
  }

  Future<CrmOpportunityContactRole?> crmCreateOpportunityContactRole({required CrmOpportunityContactRoleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCreateOpportunityContactRole(\$input: CrmOpportunityContactRoleInput!) { crmCreateOpportunityContactRole(input: \$input) ${selectionSet ?? '{ id created_at updated_at contact_id opportunity_id date_time other_details }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['crmCreateOpportunityContactRole'] == null ? null : CrmOpportunityContactRole.fromJson((data['crmCreateOpportunityContactRole'] as Map).cast<String, dynamic>());
  }

  Future<CrmOpportunityContactRole?> crmUpdateOpportunityContactRole({required int id, required CrmOpportunityContactRoleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUpdateOpportunityContactRole(\$id: Int!, \$input: CrmOpportunityContactRoleInput!) { crmUpdateOpportunityContactRole(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at contact_id opportunity_id date_time other_details }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmUpdateOpportunityContactRole'] == null ? null : CrmOpportunityContactRole.fromJson((data['crmUpdateOpportunityContactRole'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteOpportunityContactRole({required int id}) async {
    final data = await _executor.mutation('mutation crmDeleteOpportunityContactRole(\$id: Int!) { crmDeleteOpportunityContactRole(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['crmDeleteOpportunityContactRole'] as bool?;
  }

  Future<CrmConvertLeadResult?> crmConvertLead({required int id, CrmConvertLeadInput? input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmConvertLead(\$id: Int!, \$input: CrmConvertLeadInput) { crmConvertLead(id: \$id, input: \$input) ${selectionSet ?? '{ lead { id created_at updated_at lead_firstname lead_surname email phone source lead_other_details status converted_account_id converted_contact_id } account { id created_at updated_at account_name account_description account_phone billing_address email tax_code representative representative_title } contact { id created_at updated_at account_id name email phone contact_address contact_contact_details } opportunity { id created_at updated_at account_id opportunity_description opportunity_details amount close_date opportunity_stage } }'} }', variables: {
      'id': id,
      if (input != null) 'input': input.toJson(),
    }
    );
    return data['crmConvertLead'] == null ? null : CrmConvertLeadResult.fromJson((data['crmConvertLead'] as Map).cast<String, dynamic>());
  }

  Future<CrmOpportunity?> crmSetOpportunityStage({required int id, required String stage, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmSetOpportunityStage(\$id: Int!, \$stage: String!) { crmSetOpportunityStage(id: \$id, stage: \$stage) ${selectionSet ?? '{ id created_at updated_at account_id opportunity_description opportunity_details amount close_date opportunity_stage }'} }', variables: {
      'id': id,
      'stage': stage,
    }
    );
    return data['crmSetOpportunityStage'] == null ? null : CrmOpportunity.fromJson((data['crmSetOpportunityStage'] as Map).cast<String, dynamic>());
  }

  Future<CrmContract?> crmSignContract({required int id, required CrmSignContractInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmSignContract(\$id: Int!, \$input: CrmSignContractInput!) { crmSignContract(id: \$id, input: \$input) ${selectionSet ?? '{ id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['crmSignContract'] == null ? null : CrmContract.fromJson((data['crmSignContract'] as Map).cast<String, dynamic>());
  }

  Future<CrmContract?> crmSubmitContract({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmSubmitContract(\$id: Int!) { crmSubmitContract(id: \$id) ${selectionSet ?? '{ id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority }'} }', variables: {
      'id': id,
    }
    );
    return data['crmSubmitContract'] == null ? null : CrmContract.fromJson((data['crmSubmitContract'] as Map).cast<String, dynamic>());
  }

  Future<CrmContract?> crmApproveContract({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmApproveContract(\$id: Int!) { crmApproveContract(id: \$id) ${selectionSet ?? '{ id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority }'} }', variables: {
      'id': id,
    }
    );
    return data['crmApproveContract'] == null ? null : CrmContract.fromJson((data['crmApproveContract'] as Map).cast<String, dynamic>());
  }

  Future<CrmContract?> crmRejectContract({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmRejectContract(\$id: Int!) { crmRejectContract(id: \$id) ${selectionSet ?? '{ id created_at updated_at account_id contract_number contract_title end_date contract_status contract_approval signed_by signed_at signature_algorithm signature signer_certificate signer_subject signer_serial signer_fingerprint signature_payload_hash signature_format revocation_status countersigner_subject countersigner_serial countersigner_fingerprint countersigned_at timestamp_token timestamp_at timestamp_authority }'} }', variables: {
      'id': id,
    }
    );
    return data['crmRejectContract'] == null ? null : CrmContract.fromJson((data['crmRejectContract'] as Map).cast<String, dynamic>());
  }

  Future<CrmCase?> crmCloseCase({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmCloseCase(\$id: Int!) { crmCloseCase(id: \$id) ${selectionSet ?? '{ id created_at updated_at contact_id subject description priority status closed_at }'} }', variables: {
      'id': id,
    }
    );
    return data['crmCloseCase'] == null ? null : CrmCase.fromJson((data['crmCloseCase'] as Map).cast<String, dynamic>());
  }

  Future<CrmContractFile?> crmUploadContractFile({required int id, required GqlUpload file, String? kind, String? description, int? replaces, String? selectionSet}) async {
    final data = await _executor.mutation('mutation crmUploadContractFile(\$id: Int!, \$file: Upload!, \$kind: String, \$description: String, \$replaces: Int) { crmUploadContractFile(id: \$id, file: \$file, kind: \$kind, description: \$description, replaces: \$replaces) ${selectionSet ?? '{ id created_at updated_at contract_id kind name content_type format size sha256 description uploaded_by uploader_id source lineage_id version latest replaces_id scan_status scan_engine scanned_at }'} }', variables: {
      'id': id,
      'file': file,
      'kind': ?kind,
      'description': ?description,
      'replaces': ?replaces,
    }
    );
    return data['crmUploadContractFile'] == null ? null : CrmContractFile.fromJson((data['crmUploadContractFile'] as Map).cast<String, dynamic>());
  }

  Future<bool?> crmDeleteContractFile({required int id, required int fileId}) async {
    final data = await _executor.mutation('mutation crmDeleteContractFile(\$id: Int!, \$file_id: Int!) { crmDeleteContractFile(id: \$id, file_id: \$file_id) }', variables: {
      'id': id,
      'file_id': fileId,
    }
    );
    return data['crmDeleteContractFile'] as bool?;
  }

}
