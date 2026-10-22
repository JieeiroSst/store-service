// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RecruitmentApi {
  const RecruitmentApi(this._executor);

  final GatewayExecutor _executor;

  Future<RecruitJobPage?> recruitJobs({int? page, int? limit, String? status, String? departmentId, int? recruiterId, String? workMode, List<String>? skills, String? search, String? selectionSet}) async {
    final data = await _executor.query('query recruitJobs(\$page: Int, \$limit: Int, \$status: String, \$department_id: String, \$recruiter_id: Int, \$work_mode: String, \$skills: [String!], \$search: String) { recruitJobs(page: \$page, limit: \$limit, status: \$status, department_id: \$department_id, recruiter_id: \$recruiter_id, work_mode: \$work_mode, skills: \$skills, search: \$search) ${selectionSet ?? '{ items { title code department_id hiring_manager_id recruiter_ids description requirements nice_to_have skills job_type work_mode location { street city country } salary_min { amount currency } salary_max { amount currency } salary_visible headcount pipeline_stages { id name order is_system } status priority tags total_applications open_applications id created_at updated_at deleted_at } total page limit total_pages }'} }', variables: {
      'page': ?page,
      'limit': ?limit,
      'status': ?status,
      'department_id': ?departmentId,
      'recruiter_id': ?recruiterId,
      'work_mode': ?workMode,
      'skills': ?skills,
      'search': ?search,
    }
    );
    return data['recruitJobs'] == null ? null : RecruitJobPage.fromJson((data['recruitJobs'] as Map).cast<String, dynamic>());
  }

  Future<RecruitJob?> recruitJob({required String id, String? selectionSet}) async {
    final data = await _executor.query('query recruitJob(\$id: String!) { recruitJob(id: \$id) ${selectionSet ?? '{ title code department_id hiring_manager_id recruiter_ids description requirements nice_to_have skills job_type work_mode location { street city country } salary_min { amount currency } salary_max { amount currency } salary_visible headcount pipeline_stages { id name order is_system } status priority tags total_applications open_applications id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
    }
    );
    return data['recruitJob'] == null ? null : RecruitJob.fromJson((data['recruitJob'] as Map).cast<String, dynamic>());
  }

  Future<List<RecruitCandidate>?> recruitRecommendedCandidates({required String id, String? selectionSet}) async {
    final data = await _executor.query('query recruitRecommendedCandidates(\$id: String!) { recruitRecommendedCandidates(id: \$id) ${selectionSet ?? '{ full_name email phone avatar_url linkedin_url resume_url location { street city country } current_title current_company years_of_experience experience_level skills expected_salary { amount currency } notice_period_days status source referred_by_id tags notes ai_score { score confidence breakdown scored_at model_id } experiences { company title start_date end_date description } educations { institution degree major graduated_at } id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['recruitRecommendedCandidates'] as List?)?.map((e) => (e == null ? null : RecruitCandidate.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RecruitApplicationPage?> recruitApplications({int? page, int? limit, String? jobId, String? candidateId, int? recruiterId, String? status, String? selectionSet}) async {
    final data = await _executor.query('query recruitApplications(\$page: Int, \$limit: Int, \$job_id: String, \$candidate_id: String, \$recruiter_id: Int, \$status: String) { recruitApplications(page: \$page, limit: \$limit, job_id: \$job_id, candidate_id: \$candidate_id, recruiter_id: \$recruiter_id, status: \$status) ${selectionSet ?? '{ items { job_id candidate_id recruiter_id status current_stage_id rejection_reason withdraw_reason interviews { id round title interviewer_ids scheduled_at duration_min meeting_url type } offer { id start_date title benefits offer_letter_url expires_at sent_at responded_at } match_score { score confidence breakdown scored_at model_id } priority referred_by_partner_id days_in_stage last_moved_at id created_at updated_at deleted_at } total page limit total_pages }'} }', variables: {
      'page': ?page,
      'limit': ?limit,
      'job_id': ?jobId,
      'candidate_id': ?candidateId,
      'recruiter_id': ?recruiterId,
      'status': ?status,
    }
    );
    return data['recruitApplications'] == null ? null : RecruitApplicationPage.fromJson((data['recruitApplications'] as Map).cast<String, dynamic>());
  }

  Future<RecruitApplication?> recruitApplication({required String id, String? selectionSet}) async {
    final data = await _executor.query('query recruitApplication(\$id: String!) { recruitApplication(id: \$id) ${selectionSet ?? '{ job_id candidate_id recruiter_id status current_stage_id rejection_reason withdraw_reason interviews { id round title interviewer_ids scheduled_at duration_min meeting_url type feedback { submitted_by decision score strengths weaknesses notes submitted_at } } offer { id salary { amount currency } start_date title benefits offer_letter_url expires_at sent_at responded_at } match_score { score confidence breakdown scored_at model_id } priority referred_by_partner_id days_in_stage last_moved_at id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
    }
    );
    return data['recruitApplication'] == null ? null : RecruitApplication.fromJson((data['recruitApplication'] as Map).cast<String, dynamic>());
  }

  Future<RecruitCandidatePage?> recruitCandidates({int? page, int? limit, String? status, String? source, String? experienceLevel, List<String>? skills, double? minScore, String? search, String? selectionSet}) async {
    final data = await _executor.query('query recruitCandidates(\$page: Int, \$limit: Int, \$status: String, \$source: String, \$experience_level: String, \$skills: [String!], \$min_score: Float, \$search: String) { recruitCandidates(page: \$page, limit: \$limit, status: \$status, source: \$source, experience_level: \$experience_level, skills: \$skills, min_score: \$min_score, search: \$search) ${selectionSet ?? '{ items { full_name email phone avatar_url linkedin_url resume_url location { street city country } current_title current_company years_of_experience experience_level skills expected_salary { amount currency } notice_period_days status source referred_by_id tags notes ai_score { score confidence breakdown scored_at model_id } experiences { company title start_date end_date description } educations { institution degree major graduated_at } id created_at updated_at deleted_at } total page limit total_pages }'} }', variables: {
      'page': ?page,
      'limit': ?limit,
      'status': ?status,
      'source': ?source,
      'experience_level': ?experienceLevel,
      'skills': ?skills,
      'min_score': ?minScore,
      'search': ?search,
    }
    );
    return data['recruitCandidates'] == null ? null : RecruitCandidatePage.fromJson((data['recruitCandidates'] as Map).cast<String, dynamic>());
  }

  Future<RecruitCandidate?> recruitCandidate({required String id, String? selectionSet}) async {
    final data = await _executor.query('query recruitCandidate(\$id: String!) { recruitCandidate(id: \$id) ${selectionSet ?? '{ full_name email phone avatar_url linkedin_url resume_url location { street city country } current_title current_company years_of_experience experience_level skills expected_salary { amount currency } notice_period_days status source referred_by_id tags notes ai_score { score confidence breakdown scored_at model_id } experiences { company title start_date end_date description } educations { institution degree major graduated_at } id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
    }
    );
    return data['recruitCandidate'] == null ? null : RecruitCandidate.fromJson((data['recruitCandidate'] as Map).cast<String, dynamic>());
  }

  Future<RecruitPartnerStats?> recruitPartnerStats({required String id, String? selectionSet}) async {
    final data = await _executor.query('query recruitPartnerStats(\$id: String!) { recruitPartnerStats(id: \$id) ${selectionSet ?? '{ partner { user_id full_name email phone company bio referred_by_partner_id network_depth status tier commission { type percentage payout_trigger probation_days } total_referrals hired_referrals conversion_rate total_earned { amount currency } pending_payout { amount currency } id created_at updated_at deleted_at } total_referrals hired_count conversion_rate pending_payout { amount currency } total_earned { amount currency } rank }'} }', variables: {
      'id': id,
    }
    );
    return data['recruitPartnerStats'] == null ? null : RecruitPartnerStats.fromJson((data['recruitPartnerStats'] as Map).cast<String, dynamic>());
  }

  Future<RecruitPartnerNetwork?> recruitPartnerNetwork({required String id, String? selectionSet}) async {
    final data = await _executor.query('query recruitPartnerNetwork(\$id: String!) { recruitPartnerNetwork(id: \$id) ${selectionSet ?? '{ partner_id tier network_depth total_referrals hired_count }'} }', variables: {
      'id': id,
    }
    );
    return data['recruitPartnerNetwork'] == null ? null : RecruitPartnerNetwork.fromJson((data['recruitPartnerNetwork'] as Map).cast<String, dynamic>());
  }

  Future<List<RecruitPartnerStats>?> recruitReferralLeaderboard({String? selectionSet}) async {
    final data = await _executor.query('query recruitReferralLeaderboard { recruitReferralLeaderboard ${selectionSet ?? '{ partner { user_id full_name email phone company bio referred_by_partner_id network_depth status tier commission { type percentage payout_trigger probation_days } total_referrals hired_referrals conversion_rate total_earned { amount currency } pending_payout { amount currency } id created_at updated_at deleted_at } total_referrals hired_count conversion_rate pending_payout { amount currency } total_earned { amount currency } rank }'} }'
    );
    return (data['recruitReferralLeaderboard'] as List?)?.map((e) => (e == null ? null : RecruitPartnerStats.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RecruitFunnel?> recruitFunnel({String? selectionSet}) async {
    final data = await _executor.query('query recruitFunnel { recruitFunnel ${selectionSet ?? '{ funnel { job_id job_title job_code total_applications cv_review phone_screen technical final_round offer hired rejected withdrawn offer_rate_pct avg_days_to_hire } total }'} }'
    );
    return data['recruitFunnel'] == null ? null : RecruitFunnel.fromJson((data['recruitFunnel'] as Map).cast<String, dynamic>());
  }

  Future<RecruitJobFunnel?> recruitFunnelByJob({required String jobId, String? selectionSet}) async {
    final data = await _executor.query('query recruitFunnelByJob(\$job_id: String!) { recruitFunnelByJob(job_id: \$job_id) ${selectionSet ?? '{ job_id job_title total_applications cv_review phone_screen technical final_round offer hired rejected offer_rate_pct avg_days_to_hire }'} }', variables: {
      'job_id': jobId,
    }
    );
    return data['recruitFunnelByJob'] == null ? null : RecruitJobFunnel.fromJson((data['recruitFunnelByJob'] as Map).cast<String, dynamic>());
  }

  Future<List<RecruitRecruiterRow>?> recruiterPerformance({String? selectionSet}) async {
    final data = await _executor.query('query recruiterPerformance { recruiterPerformance ${selectionSet ?? '{ recruiter_id total_managed hired active_jobs avg_time_to_close_days stale_applications }'} }'
    );
    return (data['recruiterPerformance'] as List?)?.map((e) => (e == null ? null : RecruitRecruiterRow.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RecruitSLABreaches?> recruitSlaBreaches({bool? breachedOnly, String? selectionSet}) async {
    final data = await _executor.query('query recruitSlaBreaches(\$breached_only: Boolean) { recruitSlaBreaches(breached_only: \$breached_only) ${selectionSet ?? '{ applications { application_id job_id candidate_id recruiter_id status days_in_stage sla_days is_breached } total breached_count }'} }', variables: {
      'breached_only': ?breachedOnly,
    }
    );
    return data['recruitSlaBreaches'] == null ? null : RecruitSLABreaches.fromJson((data['recruitSlaBreaches'] as Map).cast<String, dynamic>());
  }

  Future<List<RecruitScoreRow>?> recruitAiScoreDistribution({String? selectionSet}) async {
    final data = await _executor.query('query recruitAiScoreDistribution { recruitAiScoreDistribution ${selectionSet ?? '{ job_id job_title scored_applications avg_score min_score max_score high_match medium_match low_match }'} }'
    );
    return (data['recruitAiScoreDistribution'] as List?)?.map((e) => (e == null ? null : RecruitScoreRow.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<RecruitNetworkRow>?> recruitReferralNetworkStats({String? selectionSet}) async {
    final data = await _executor.query('query recruitReferralNetworkStats { recruitReferralNetworkStats ${selectionSet ?? '{ partner_id full_name network_level hired_referrals direct_downline }'} }'
    );
    return (data['recruitReferralNetworkStats'] as List?)?.map((e) => (e == null ? null : RecruitNetworkRow.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RecruitJob?> recruitCreateJob({required RecruitJobInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitCreateJob(\$input: RecruitJobInput!) { recruitCreateJob(input: \$input) ${selectionSet ?? '{ title code department_id hiring_manager_id recruiter_ids description requirements nice_to_have skills job_type work_mode location { street city country } salary_min { amount currency } salary_max { amount currency } salary_visible headcount pipeline_stages { id name order is_system } status priority tags total_applications open_applications id created_at updated_at deleted_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['recruitCreateJob'] == null ? null : RecruitJob.fromJson((data['recruitCreateJob'] as Map).cast<String, dynamic>());
  }

  Future<RecruitJob?> recruitUpdateJob({required String id, required RecruitJobInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitUpdateJob(\$id: String!, \$input: RecruitJobInput!) { recruitUpdateJob(id: \$id, input: \$input) ${selectionSet ?? '{ title code department_id hiring_manager_id recruiter_ids description requirements nice_to_have skills job_type work_mode location { street city country } salary_min { amount currency } salary_max { amount currency } salary_visible headcount pipeline_stages { id name order is_system } status priority tags total_applications open_applications id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['recruitUpdateJob'] == null ? null : RecruitJob.fromJson((data['recruitUpdateJob'] as Map).cast<String, dynamic>());
  }

  Future<RecruitJob?> recruitPublishJob({required String id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitPublishJob(\$id: String!, \$reason: String) { recruitPublishJob(id: \$id, reason: \$reason) ${selectionSet ?? '{ title code department_id hiring_manager_id recruiter_ids description requirements nice_to_have skills job_type work_mode location { street city country } salary_min { amount currency } salary_max { amount currency } salary_visible headcount pipeline_stages { id name order is_system } status priority tags total_applications open_applications id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['recruitPublishJob'] == null ? null : RecruitJob.fromJson((data['recruitPublishJob'] as Map).cast<String, dynamic>());
  }

  Future<RecruitJob?> recruitPauseJob({required String id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitPauseJob(\$id: String!, \$reason: String) { recruitPauseJob(id: \$id, reason: \$reason) ${selectionSet ?? '{ title code department_id hiring_manager_id recruiter_ids description requirements nice_to_have skills job_type work_mode location { street city country } salary_min { amount currency } salary_max { amount currency } salary_visible headcount pipeline_stages { id name order is_system } status priority tags total_applications open_applications id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['recruitPauseJob'] == null ? null : RecruitJob.fromJson((data['recruitPauseJob'] as Map).cast<String, dynamic>());
  }

  Future<RecruitJob?> recruitCloseJob({required String id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitCloseJob(\$id: String!, \$reason: String) { recruitCloseJob(id: \$id, reason: \$reason) ${selectionSet ?? '{ title code department_id hiring_manager_id recruiter_ids description requirements nice_to_have skills job_type work_mode location { street city country } salary_min { amount currency } salary_max { amount currency } salary_visible headcount pipeline_stages { id name order is_system } status priority tags total_applications open_applications id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['recruitCloseJob'] == null ? null : RecruitJob.fromJson((data['recruitCloseJob'] as Map).cast<String, dynamic>());
  }

  Future<RecruitApplication?> recruitApply({required String jobId, required String candidateId, String? referredByPartnerId, String? coverLetter, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitApply(\$job_id: String!, \$candidate_id: String!, \$referred_by_partner_id: String, \$cover_letter: String) { recruitApply(job_id: \$job_id, candidate_id: \$candidate_id, referred_by_partner_id: \$referred_by_partner_id, cover_letter: \$cover_letter) ${selectionSet ?? '{ job_id candidate_id recruiter_id status current_stage_id rejection_reason withdraw_reason interviews { id round title interviewer_ids scheduled_at duration_min meeting_url type feedback { submitted_by decision score strengths weaknesses notes submitted_at } } offer { id salary { amount currency } start_date title benefits offer_letter_url expires_at sent_at responded_at } match_score { score confidence breakdown scored_at model_id } priority referred_by_partner_id days_in_stage last_moved_at id created_at updated_at deleted_at }'} }', variables: {
      'job_id': jobId,
      'candidate_id': candidateId,
      'referred_by_partner_id': ?referredByPartnerId,
      'cover_letter': ?coverLetter,
    }
    );
    return data['recruitApply'] == null ? null : RecruitApplication.fromJson((data['recruitApply'] as Map).cast<String, dynamic>());
  }

  Future<RecruitApplication?> recruitMoveStage({required String id, required String stageId, required String status, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitMoveStage(\$id: String!, \$stage_id: String!, \$status: String!) { recruitMoveStage(id: \$id, stage_id: \$stage_id, status: \$status) ${selectionSet ?? '{ job_id candidate_id recruiter_id status current_stage_id rejection_reason withdraw_reason interviews { id round title interviewer_ids scheduled_at duration_min meeting_url type feedback { submitted_by decision score strengths weaknesses notes submitted_at } } offer { id salary { amount currency } start_date title benefits offer_letter_url expires_at sent_at responded_at } match_score { score confidence breakdown scored_at model_id } priority referred_by_partner_id days_in_stage last_moved_at id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
      'stage_id': stageId,
      'status': status,
    }
    );
    return data['recruitMoveStage'] == null ? null : RecruitApplication.fromJson((data['recruitMoveStage'] as Map).cast<String, dynamic>());
  }

  Future<Object?> recruitScheduleInterview({required String id, required RecruitInterviewInput input}) async {
    final data = await _executor.mutation('mutation recruitScheduleInterview(\$id: String!, \$input: RecruitInterviewInput!) { recruitScheduleInterview(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['recruitScheduleInterview'];
  }

  Future<Object?> recruitSubmitFeedback({required String id, required String interviewId, required RecruitFeedbackInput input}) async {
    final data = await _executor.mutation('mutation recruitSubmitFeedback(\$id: String!, \$interviewId: String!, \$input: RecruitFeedbackInput!) { recruitSubmitFeedback(id: \$id, interviewId: \$interviewId, input: \$input) }', variables: {
      'id': id,
      'interviewId': interviewId,
      'input': input.toJson(),
    }
    );
    return data['recruitSubmitFeedback'];
  }

  Future<Object?> recruitExtendOffer({required String id, required RecruitOfferInput input}) async {
    final data = await _executor.mutation('mutation recruitExtendOffer(\$id: String!, \$input: RecruitOfferInput!) { recruitExtendOffer(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['recruitExtendOffer'];
  }

  Future<RecruitApplication?> recruitReject({required String id, required String reason, String? note, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitReject(\$id: String!, \$reason: String!, \$note: String) { recruitReject(id: \$id, reason: \$reason, note: \$note) ${selectionSet ?? '{ job_id candidate_id recruiter_id status current_stage_id rejection_reason withdraw_reason interviews { id round title interviewer_ids scheduled_at duration_min meeting_url type feedback { submitted_by decision score strengths weaknesses notes submitted_at } } offer { id salary { amount currency } start_date title benefits offer_letter_url expires_at sent_at responded_at } match_score { score confidence breakdown scored_at model_id } priority referred_by_partner_id days_in_stage last_moved_at id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
      'reason': reason,
      'note': ?note,
    }
    );
    return data['recruitReject'] == null ? null : RecruitApplication.fromJson((data['recruitReject'] as Map).cast<String, dynamic>());
  }

  Future<RecruitPartner?> recruitRegisterPartner({required RecruitPartnerInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitRegisterPartner(\$input: RecruitPartnerInput!) { recruitRegisterPartner(input: \$input) ${selectionSet ?? '{ user_id full_name email phone company bio referred_by_partner_id network_depth status tier commission { type fixed_amount { amount currency } percentage payout_trigger probation_days } total_referrals hired_referrals conversion_rate total_earned { amount currency } pending_payout { amount currency } id created_at updated_at deleted_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['recruitRegisterPartner'] == null ? null : RecruitPartner.fromJson((data['recruitRegisterPartner'] as Map).cast<String, dynamic>());
  }

  Future<Object?> recruitRequestPayout({required String id}) async {
    final data = await _executor.mutation('mutation recruitRequestPayout(\$id: String!) { recruitRequestPayout(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['recruitRequestPayout'];
  }

  Future<Object?> recruitGenerateLink({required String partnerId, String? jobId}) async {
    final data = await _executor.mutation('mutation recruitGenerateLink(\$partner_id: String!, \$job_id: String) { recruitGenerateLink(partner_id: \$partner_id, job_id: \$job_id) }', variables: {
      'partner_id': partnerId,
      'job_id': ?jobId,
    }
    );
    return data['recruitGenerateLink'];
  }

  Future<RecruitCandidate?> recruitCreateCandidate({required RecruitCandidateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitCreateCandidate(\$input: RecruitCandidateInput!) { recruitCreateCandidate(input: \$input) ${selectionSet ?? '{ full_name email phone avatar_url linkedin_url resume_url location { street city country } current_title current_company years_of_experience experience_level skills expected_salary { amount currency } notice_period_days status source referred_by_id tags notes ai_score { score confidence breakdown scored_at model_id } experiences { company title start_date end_date description } educations { institution degree major graduated_at } id created_at updated_at deleted_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['recruitCreateCandidate'] == null ? null : RecruitCandidate.fromJson((data['recruitCreateCandidate'] as Map).cast<String, dynamic>());
  }

  Future<RecruitCandidate?> recruitUpdateCandidate({required String id, required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitUpdateCandidate(\$id: String!, \$input: Map!) { recruitUpdateCandidate(id: \$id, input: \$input) ${selectionSet ?? '{ full_name email phone avatar_url linkedin_url resume_url location { street city country } current_title current_company years_of_experience experience_level skills expected_salary { amount currency } notice_period_days status source referred_by_id tags notes ai_score { score confidence breakdown scored_at model_id } experiences { company title start_date end_date description } educations { institution degree major graduated_at } id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['recruitUpdateCandidate'] == null ? null : RecruitCandidate.fromJson((data['recruitUpdateCandidate'] as Map).cast<String, dynamic>());
  }

  Future<RecruitCandidate?> recruitTransitionCandidate({required String id, required String status, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation recruitTransitionCandidate(\$id: String!, \$status: String!, \$reason: String) { recruitTransitionCandidate(id: \$id, status: \$status, reason: \$reason) ${selectionSet ?? '{ full_name email phone avatar_url linkedin_url resume_url location { street city country } current_title current_company years_of_experience experience_level skills expected_salary { amount currency } notice_period_days status source referred_by_id tags notes ai_score { score confidence breakdown scored_at model_id } experiences { company title start_date end_date description } educations { institution degree major graduated_at } id created_at updated_at deleted_at }'} }', variables: {
      'id': id,
      'status': status,
      'reason': ?reason,
    }
    );
    return data['recruitTransitionCandidate'] == null ? null : RecruitCandidate.fromJson((data['recruitTransitionCandidate'] as Map).cast<String, dynamic>());
  }

  Future<Object?> recruitEnrichCandidate({required String id}) async {
    final data = await _executor.mutation('mutation recruitEnrichCandidate(\$id: String!) { recruitEnrichCandidate(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['recruitEnrichCandidate'];
  }

}
