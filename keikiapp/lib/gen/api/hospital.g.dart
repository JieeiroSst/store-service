// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class HospitalApi {
  const HospitalApi(this._executor);

  final GatewayExecutor _executor;

  Future<HospitalDepartment?> hospitalGetDepartment({required HospitalDepartmentInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetDepartment(\$input: HospitalDepartmentInput!) { hospitalGetDepartment(input: \$input) ${selectionSet ?? '{ departmentId name description createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetDepartment'] == null ? null : HospitalDepartment.fromJson((data['hospitalGetDepartment'] as Map).cast<String, dynamic>());
  }

  Future<HospitalListDepartmentsResponse?> hospitalListDepartments({required HospitalListDepartmentsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalListDepartments(\$input: HospitalListDepartmentsRequestInput!) { hospitalListDepartments(input: \$input) ${selectionSet ?? '{ departments { departmentId name description createdAt updatedAt } totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalListDepartments'] == null ? null : HospitalListDepartmentsResponse.fromJson((data['hospitalListDepartments'] as Map).cast<String, dynamic>());
  }

  Future<HospitalStaff?> hospitalGetStaff({required HospitalStaffInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetStaff(\$input: HospitalStaffInput!) { hospitalGetStaff(input: \$input) ${selectionSet ?? '{ staffId departmentId firstName lastName role specialization email phone hireDate licenseNumber createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetStaff'] == null ? null : HospitalStaff.fromJson((data['hospitalGetStaff'] as Map).cast<String, dynamic>());
  }

  Future<HospitalListStaffResponse?> hospitalListStaff({required HospitalListStaffRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalListStaff(\$input: HospitalListStaffRequestInput!) { hospitalListStaff(input: \$input) ${selectionSet ?? '{ staff { staffId departmentId firstName lastName role specialization email phone hireDate licenseNumber createdAt updatedAt } totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalListStaff'] == null ? null : HospitalListStaffResponse.fromJson((data['hospitalListStaff'] as Map).cast<String, dynamic>());
  }

  Future<HospitalPatient?> hospitalGetPatient({required HospitalPatientInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetPatient(\$input: HospitalPatientInput!) { hospitalGetPatient(input: \$input) ${selectionSet ?? '{ patientId firstName lastName dateOfBirth gender bloodType address phone email emergencyContactName emergencyContactPhone insuranceProvider insurancePolicyNumber createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetPatient'] == null ? null : HospitalPatient.fromJson((data['hospitalGetPatient'] as Map).cast<String, dynamic>());
  }

  Future<HospitalListPatientsResponse?> hospitalListPatients({required HospitalListPatientsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalListPatients(\$input: HospitalListPatientsRequestInput!) { hospitalListPatients(input: \$input) ${selectionSet ?? '{ patients { patientId firstName lastName dateOfBirth gender bloodType address phone email emergencyContactName emergencyContactPhone insuranceProvider insurancePolicyNumber createdAt updatedAt } totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalListPatients'] == null ? null : HospitalListPatientsResponse.fromJson((data['hospitalListPatients'] as Map).cast<String, dynamic>());
  }

  Future<HospitalMedicalRecord?> hospitalGetMedicalRecord({required HospitalMedicalRecordInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetMedicalRecord(\$input: HospitalMedicalRecordInput!) { hospitalGetMedicalRecord(input: \$input) ${selectionSet ?? '{ recordId patientId diagnosis treatmentPlan notes createdBy createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetMedicalRecord'] == null ? null : HospitalMedicalRecord.fromJson((data['hospitalGetMedicalRecord'] as Map).cast<String, dynamic>());
  }

  Future<HospitalListMedicalRecordsResponse?> hospitalListMedicalRecords({required HospitalListMedicalRecordsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalListMedicalRecords(\$input: HospitalListMedicalRecordsRequestInput!) { hospitalListMedicalRecords(input: \$input) ${selectionSet ?? '{ records { recordId patientId diagnosis treatmentPlan notes createdBy createdAt updatedAt } totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalListMedicalRecords'] == null ? null : HospitalListMedicalRecordsResponse.fromJson((data['hospitalListMedicalRecords'] as Map).cast<String, dynamic>());
  }

  Future<HospitalAppointment?> hospitalGetAppointment({required HospitalAppointmentInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetAppointment(\$input: HospitalAppointmentInput!) { hospitalGetAppointment(input: \$input) ${selectionSet ?? '{ appointmentId patientId staffId departmentId appointmentDate status reason notes createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetAppointment'] == null ? null : HospitalAppointment.fromJson((data['hospitalGetAppointment'] as Map).cast<String, dynamic>());
  }

  Future<HospitalListAppointmentsResponse?> hospitalListAppointments({required HospitalListAppointmentsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalListAppointments(\$input: HospitalListAppointmentsRequestInput!) { hospitalListAppointments(input: \$input) ${selectionSet ?? '{ appointments { appointmentId patientId staffId departmentId appointmentDate status reason notes createdAt updatedAt } totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalListAppointments'] == null ? null : HospitalListAppointmentsResponse.fromJson((data['hospitalListAppointments'] as Map).cast<String, dynamic>());
  }

  Future<HospitalPrescription?> hospitalGetPrescription({required HospitalPrescriptionInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetPrescription(\$input: HospitalPrescriptionInput!) { hospitalGetPrescription(input: \$input) ${selectionSet ?? '{ prescriptionId patientId prescribedBy medicationName dosage frequency startDate endDate instructions createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetPrescription'] == null ? null : HospitalPrescription.fromJson((data['hospitalGetPrescription'] as Map).cast<String, dynamic>());
  }

  Future<HospitalListPrescriptionsResponse?> hospitalListPrescriptions({required HospitalListPrescriptionsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalListPrescriptions(\$input: HospitalListPrescriptionsRequestInput!) { hospitalListPrescriptions(input: \$input) ${selectionSet ?? '{ prescriptions { prescriptionId patientId prescribedBy medicationName dosage frequency startDate endDate instructions createdAt updatedAt } totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalListPrescriptions'] == null ? null : HospitalListPrescriptionsResponse.fromJson((data['hospitalListPrescriptions'] as Map).cast<String, dynamic>());
  }

  Future<HospitalLabResult?> hospitalGetLabResult({required HospitalLabResultInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetLabResult(\$input: HospitalLabResultInput!) { hospitalGetLabResult(input: \$input) ${selectionSet ?? '{ resultId patientId orderedBy testName testDate results normalRange notes createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetLabResult'] == null ? null : HospitalLabResult.fromJson((data['hospitalGetLabResult'] as Map).cast<String, dynamic>());
  }

  Future<HospitalListLabResultsResponse?> hospitalListLabResults({required HospitalListLabResultsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalListLabResults(\$input: HospitalListLabResultsRequestInput!) { hospitalListLabResults(input: \$input) ${selectionSet ?? '{ labResults { resultId patientId orderedBy testName testDate results normalRange notes createdAt updatedAt } totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalListLabResults'] == null ? null : HospitalListLabResultsResponse.fromJson((data['hospitalListLabResults'] as Map).cast<String, dynamic>());
  }

  Future<HospitalBilling?> hospitalGetBilling({required HospitalBillingInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetBilling(\$input: HospitalBillingInput!) { hospitalGetBilling(input: \$input) ${selectionSet ?? '{ billId patientId appointmentId amount insuranceCovered patientResponsibility status dueDate paidDate createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetBilling'] == null ? null : HospitalBilling.fromJson((data['hospitalGetBilling'] as Map).cast<String, dynamic>());
  }

  Future<HospitalListBillingsResponse?> hospitalListBillings({required HospitalListBillingsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalListBillings(\$input: HospitalListBillingsRequestInput!) { hospitalListBillings(input: \$input) ${selectionSet ?? '{ billings { billId patientId appointmentId amount insuranceCovered patientResponsibility status dueDate paidDate createdAt updatedAt } totalCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalListBillings'] == null ? null : HospitalListBillingsResponse.fromJson((data['hospitalListBillings'] as Map).cast<String, dynamic>());
  }

  Future<HospitalDepartmentStatsResponse?> hospitalGetDepartmentStats({String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetDepartmentStats { hospitalGetDepartmentStats ${selectionSet ?? '{ departmentStats { departmentId departmentName staffCount appointmentCount } }'} }'
    );
    return data['hospitalGetDepartmentStats'] == null ? null : HospitalDepartmentStatsResponse.fromJson((data['hospitalGetDepartmentStats'] as Map).cast<String, dynamic>());
  }

  Future<HospitalAppointmentStatsResponse?> hospitalGetAppointmentStats({required HospitalAppointmentStatsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetAppointmentStats(\$input: HospitalAppointmentStatsRequestInput!) { hospitalGetAppointmentStats(input: \$input) ${selectionSet ?? '{ totalAppointments completedAppointments cancelledAppointments noShowAppointments dailyStats { date appointmentCount } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetAppointmentStats'] == null ? null : HospitalAppointmentStatsResponse.fromJson((data['hospitalGetAppointmentStats'] as Map).cast<String, dynamic>());
  }

  Future<HospitalBillingStatsResponse?> hospitalGetBillingStats({required HospitalBillingStatsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query hospitalGetBillingStats(\$input: HospitalBillingStatsRequestInput!) { hospitalGetBillingStats(input: \$input) ${selectionSet ?? '{ totalBilled totalCollected totalOutstanding totalInsuranceCovered totalPatientResponsibility monthlyStats { month amountBilled amountCollected } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalGetBillingStats'] == null ? null : HospitalBillingStatsResponse.fromJson((data['hospitalGetBillingStats'] as Map).cast<String, dynamic>());
  }

  Future<HospitalDepartment?> hospitalCreateDepartment({required HospitalCreateDepartmentRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalCreateDepartment(\$input: HospitalCreateDepartmentRequestInput!) { hospitalCreateDepartment(input: \$input) ${selectionSet ?? '{ departmentId name description createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalCreateDepartment'] == null ? null : HospitalDepartment.fromJson((data['hospitalCreateDepartment'] as Map).cast<String, dynamic>());
  }

  Future<HospitalDepartment?> hospitalUpdateDepartment({required HospitalDepartmentInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdateDepartment(\$input: HospitalDepartmentInput!) { hospitalUpdateDepartment(input: \$input) ${selectionSet ?? '{ departmentId name description createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdateDepartment'] == null ? null : HospitalDepartment.fromJson((data['hospitalUpdateDepartment'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hospitalDeleteDepartment({required HospitalDepartmentInput input}) async {
    final data = await _executor.mutation('mutation hospitalDeleteDepartment(\$input: HospitalDepartmentInput!) { hospitalDeleteDepartment(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalDeleteDepartment'] as bool?;
  }

  Future<HospitalStaff?> hospitalCreateStaff({required HospitalCreateStaffRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalCreateStaff(\$input: HospitalCreateStaffRequestInput!) { hospitalCreateStaff(input: \$input) ${selectionSet ?? '{ staffId departmentId firstName lastName role specialization email phone hireDate licenseNumber createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalCreateStaff'] == null ? null : HospitalStaff.fromJson((data['hospitalCreateStaff'] as Map).cast<String, dynamic>());
  }

  Future<HospitalStaff?> hospitalUpdateStaff({required HospitalStaffInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdateStaff(\$input: HospitalStaffInput!) { hospitalUpdateStaff(input: \$input) ${selectionSet ?? '{ staffId departmentId firstName lastName role specialization email phone hireDate licenseNumber createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdateStaff'] == null ? null : HospitalStaff.fromJson((data['hospitalUpdateStaff'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hospitalDeleteStaff({required HospitalStaffInput input}) async {
    final data = await _executor.mutation('mutation hospitalDeleteStaff(\$input: HospitalStaffInput!) { hospitalDeleteStaff(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalDeleteStaff'] as bool?;
  }

  Future<HospitalPatient?> hospitalCreatePatient({required HospitalCreatePatientRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalCreatePatient(\$input: HospitalCreatePatientRequestInput!) { hospitalCreatePatient(input: \$input) ${selectionSet ?? '{ patientId firstName lastName dateOfBirth gender bloodType address phone email emergencyContactName emergencyContactPhone insuranceProvider insurancePolicyNumber createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalCreatePatient'] == null ? null : HospitalPatient.fromJson((data['hospitalCreatePatient'] as Map).cast<String, dynamic>());
  }

  Future<HospitalPatient?> hospitalUpdatePatient({required HospitalPatientInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdatePatient(\$input: HospitalPatientInput!) { hospitalUpdatePatient(input: \$input) ${selectionSet ?? '{ patientId firstName lastName dateOfBirth gender bloodType address phone email emergencyContactName emergencyContactPhone insuranceProvider insurancePolicyNumber createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdatePatient'] == null ? null : HospitalPatient.fromJson((data['hospitalUpdatePatient'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hospitalDeletePatient({required HospitalPatientInput input}) async {
    final data = await _executor.mutation('mutation hospitalDeletePatient(\$input: HospitalPatientInput!) { hospitalDeletePatient(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalDeletePatient'] as bool?;
  }

  Future<HospitalMedicalRecord?> hospitalCreateMedicalRecord({required HospitalCreateMedicalRecordRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalCreateMedicalRecord(\$input: HospitalCreateMedicalRecordRequestInput!) { hospitalCreateMedicalRecord(input: \$input) ${selectionSet ?? '{ recordId patientId diagnosis treatmentPlan notes createdBy createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalCreateMedicalRecord'] == null ? null : HospitalMedicalRecord.fromJson((data['hospitalCreateMedicalRecord'] as Map).cast<String, dynamic>());
  }

  Future<HospitalMedicalRecord?> hospitalUpdateMedicalRecord({required HospitalMedicalRecordInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdateMedicalRecord(\$input: HospitalMedicalRecordInput!) { hospitalUpdateMedicalRecord(input: \$input) ${selectionSet ?? '{ recordId patientId diagnosis treatmentPlan notes createdBy createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdateMedicalRecord'] == null ? null : HospitalMedicalRecord.fromJson((data['hospitalUpdateMedicalRecord'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hospitalDeleteMedicalRecord({required HospitalMedicalRecordInput input}) async {
    final data = await _executor.mutation('mutation hospitalDeleteMedicalRecord(\$input: HospitalMedicalRecordInput!) { hospitalDeleteMedicalRecord(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalDeleteMedicalRecord'] as bool?;
  }

  Future<HospitalAppointment?> hospitalCreateAppointment({required HospitalCreateAppointmentRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalCreateAppointment(\$input: HospitalCreateAppointmentRequestInput!) { hospitalCreateAppointment(input: \$input) ${selectionSet ?? '{ appointmentId patientId staffId departmentId appointmentDate status reason notes createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalCreateAppointment'] == null ? null : HospitalAppointment.fromJson((data['hospitalCreateAppointment'] as Map).cast<String, dynamic>());
  }

  Future<HospitalAppointment?> hospitalUpdateAppointmentStatus({required HospitalUpdateAppointmentStatusRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdateAppointmentStatus(\$input: HospitalUpdateAppointmentStatusRequestInput!) { hospitalUpdateAppointmentStatus(input: \$input) ${selectionSet ?? '{ appointmentId patientId staffId departmentId appointmentDate status reason notes createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdateAppointmentStatus'] == null ? null : HospitalAppointment.fromJson((data['hospitalUpdateAppointmentStatus'] as Map).cast<String, dynamic>());
  }

  Future<HospitalAppointment?> hospitalUpdateAppointment({required HospitalAppointmentInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdateAppointment(\$input: HospitalAppointmentInput!) { hospitalUpdateAppointment(input: \$input) ${selectionSet ?? '{ appointmentId patientId staffId departmentId appointmentDate status reason notes createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdateAppointment'] == null ? null : HospitalAppointment.fromJson((data['hospitalUpdateAppointment'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hospitalDeleteAppointment({required HospitalAppointmentInput input}) async {
    final data = await _executor.mutation('mutation hospitalDeleteAppointment(\$input: HospitalAppointmentInput!) { hospitalDeleteAppointment(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalDeleteAppointment'] as bool?;
  }

  Future<HospitalPrescription?> hospitalCreatePrescription({required HospitalCreatePrescriptionRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalCreatePrescription(\$input: HospitalCreatePrescriptionRequestInput!) { hospitalCreatePrescription(input: \$input) ${selectionSet ?? '{ prescriptionId patientId prescribedBy medicationName dosage frequency startDate endDate instructions createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalCreatePrescription'] == null ? null : HospitalPrescription.fromJson((data['hospitalCreatePrescription'] as Map).cast<String, dynamic>());
  }

  Future<HospitalPrescription?> hospitalUpdatePrescription({required HospitalPrescriptionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdatePrescription(\$input: HospitalPrescriptionInput!) { hospitalUpdatePrescription(input: \$input) ${selectionSet ?? '{ prescriptionId patientId prescribedBy medicationName dosage frequency startDate endDate instructions createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdatePrescription'] == null ? null : HospitalPrescription.fromJson((data['hospitalUpdatePrescription'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hospitalDeletePrescription({required HospitalPrescriptionInput input}) async {
    final data = await _executor.mutation('mutation hospitalDeletePrescription(\$input: HospitalPrescriptionInput!) { hospitalDeletePrescription(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalDeletePrescription'] as bool?;
  }

  Future<HospitalLabResult?> hospitalCreateLabResult({required HospitalCreateLabResultRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalCreateLabResult(\$input: HospitalCreateLabResultRequestInput!) { hospitalCreateLabResult(input: \$input) ${selectionSet ?? '{ resultId patientId orderedBy testName testDate results normalRange notes createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalCreateLabResult'] == null ? null : HospitalLabResult.fromJson((data['hospitalCreateLabResult'] as Map).cast<String, dynamic>());
  }

  Future<HospitalLabResult?> hospitalUpdateLabResult({required HospitalLabResultInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdateLabResult(\$input: HospitalLabResultInput!) { hospitalUpdateLabResult(input: \$input) ${selectionSet ?? '{ resultId patientId orderedBy testName testDate results normalRange notes createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdateLabResult'] == null ? null : HospitalLabResult.fromJson((data['hospitalUpdateLabResult'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hospitalDeleteLabResult({required HospitalLabResultInput input}) async {
    final data = await _executor.mutation('mutation hospitalDeleteLabResult(\$input: HospitalLabResultInput!) { hospitalDeleteLabResult(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalDeleteLabResult'] as bool?;
  }

  Future<HospitalBilling?> hospitalCreateBilling({required HospitalCreateBillingRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalCreateBilling(\$input: HospitalCreateBillingRequestInput!) { hospitalCreateBilling(input: \$input) ${selectionSet ?? '{ billId patientId appointmentId amount insuranceCovered patientResponsibility status dueDate paidDate createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalCreateBilling'] == null ? null : HospitalBilling.fromJson((data['hospitalCreateBilling'] as Map).cast<String, dynamic>());
  }

  Future<HospitalBilling?> hospitalUpdateBillingStatus({required HospitalUpdateBillingStatusRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdateBillingStatus(\$input: HospitalUpdateBillingStatusRequestInput!) { hospitalUpdateBillingStatus(input: \$input) ${selectionSet ?? '{ billId patientId appointmentId amount insuranceCovered patientResponsibility status dueDate paidDate createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdateBillingStatus'] == null ? null : HospitalBilling.fromJson((data['hospitalUpdateBillingStatus'] as Map).cast<String, dynamic>());
  }

  Future<HospitalBilling?> hospitalUpdateBilling({required HospitalBillingInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation hospitalUpdateBilling(\$input: HospitalBillingInput!) { hospitalUpdateBilling(input: \$input) ${selectionSet ?? '{ billId patientId appointmentId amount insuranceCovered patientResponsibility status dueDate paidDate createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalUpdateBilling'] == null ? null : HospitalBilling.fromJson((data['hospitalUpdateBilling'] as Map).cast<String, dynamic>());
  }

  Future<bool?> hospitalDeleteBilling({required HospitalBillingInput input}) async {
    final data = await _executor.mutation('mutation hospitalDeleteBilling(\$input: HospitalBillingInput!) { hospitalDeleteBilling(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['hospitalDeleteBilling'] as bool?;
  }

}
