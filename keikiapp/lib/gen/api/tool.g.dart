// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ToolApi {
  const ToolApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<ToolJob>?> toolJobs({String? selectionSet}) async {
    final data = await _executor.query('query toolJobs { toolJobs ${selectionSet ?? '{ id kind status progress started_at finished_at result error }'} }'
    );
    return (data['toolJobs'] as List?)?.map((e) => (e == null ? null : ToolJob.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ToolJob?> toolJob({required String id, String? selectionSet}) async {
    final data = await _executor.query('query toolJob(\$id: String!) { toolJob(id: \$id) ${selectionSet ?? '{ id kind status progress started_at finished_at result error }'} }', variables: {
      'id': id,
    }
    );
    return data['toolJob'] == null ? null : ToolJob.fromJson((data['toolJob'] as Map).cast<String, dynamic>());
  }

  Future<ToolMobileDevices?> toolMobileDevices({String? selectionSet}) async {
    final data = await _executor.query('query toolMobileDevices { toolMobileDevices ${selectionSet ?? '{ android { available error devices { serial state model product android_version sdk screen density_dpi } } ios { available error simulators { udid name state runtime } } }'} }'
    );
    return data['toolMobileDevices'] == null ? null : ToolMobileDevices.fromJson((data['toolMobileDevices'] as Map).cast<String, dynamic>());
  }

  Future<ToolStats?> toolLearningStats({String? selectionSet}) async {
    final data = await _executor.query('query toolLearningStats { toolLearningStats ${selectionSet ?? '{ active_model base_model by_model by_model_human_only total good bad pending approval_rate feedback_since_reflect playbook }'} }'
    );
    return data['toolLearningStats'] == null ? null : ToolStats.fromJson((data['toolLearningStats'] as Map).cast<String, dynamic>());
  }

  Future<List<ToolExperience>?> toolLearningPending({String? selectionSet}) async {
    final data = await _executor.query('query toolLearningPending { toolLearningPending ${selectionSet ?? '{ id kind input output model verdict source score note corrected signals embedding created_at }'} }'
    );
    return (data['toolLearningPending'] as List?)?.map((e) => (e == null ? null : ToolExperience.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ToolSnapshot>?> toolLearningHistory({String? selectionSet}) async {
    final data = await _executor.query('query toolLearningHistory { toolLearningHistory ${selectionSet ?? '{ at model approval_rate human_rate auto_rate reviews by_model }'} }'
    );
    return (data['toolLearningHistory'] as List?)?.map((e) => (e == null ? null : ToolSnapshot.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ToolSuiteResult?> toolSuite({required String name_, String? selectionSet}) async {
    final data = await _executor.query('query toolSuite(\$name: String!) { toolSuite(name: \$name) ${selectionSet ?? '{ suite { name cases { hash approved_at from_ai } ai_accepted ai_rejected } precision }'} }', variables: {
      'name': name_,
    }
    );
    return data['toolSuite'] == null ? null : ToolSuiteResult.fromJson((data['toolSuite'] as Map).cast<String, dynamic>());
  }

  Future<Object?> toolGenerateTestCases({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolGenerateTestCases(\$input: Map) { toolGenerateTestCases(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolGenerateTestCases'];
  }

  Future<Object?> toolRunTestCases({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolRunTestCases(\$input: Map) { toolRunTestCases(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolRunTestCases'];
  }

  Future<Object?> toolRunBenchmark({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolRunBenchmark(\$input: Map) { toolRunBenchmark(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolRunBenchmark'];
  }

  Future<Object?> toolAnalyzeCode({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolAnalyzeCode(\$input: Map) { toolAnalyzeCode(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolAnalyzeCode'];
  }

  Future<Object?> toolAnalyzeTask({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolAnalyzeTask(\$input: Map) { toolAnalyzeTask(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolAnalyzeTask'];
  }

  Future<Object?> toolQaScan({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolQaScan(\$input: Map) { toolQaScan(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolQaScan'];
  }

  Future<Object?> toolExportXlsx({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolExportXlsx(\$input: Map) { toolExportXlsx(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolExportXlsx'];
  }

  Future<Object?> toolQaDatabase({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolQaDatabase(\$input: Map) { toolQaDatabase(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolQaDatabase'];
  }

  Future<Object?> toolQaAsync({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolQaAsync(\$input: Map) { toolQaAsync(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolQaAsync'];
  }

  Future<Object?> toolQaQueue({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolQaQueue(\$input: Map) { toolQaQueue(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolQaQueue'];
  }

  Future<Object?> toolQaBrowser({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolQaBrowser(\$input: Map) { toolQaBrowser(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolQaBrowser'];
  }

  Future<Object?> toolQaActive({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolQaActive(\$input: Map) { toolQaActive(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolQaActive'];
  }

  Future<Object?> toolJobStress({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolJobStress(\$input: Map) { toolJobStress(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolJobStress'];
  }

  Future<Object?> toolJobSoak({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolJobSoak(\$input: Map) { toolJobSoak(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolJobSoak'];
  }

  Future<Object?> toolMobileStatic({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolMobileStatic(\$input: Map) { toolMobileStatic(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolMobileStatic'];
  }

  Future<Object?> toolMobileAndroid({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolMobileAndroid(\$input: Map) { toolMobileAndroid(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolMobileAndroid'];
  }

  Future<Object?> toolMobileIos({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolMobileIos(\$input: Map) { toolMobileIos(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolMobileIos'];
  }

  Future<Object?> toolQcRun({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolQcRun(\$input: Map) { toolQcRun(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolQcRun'];
  }

  Future<Object?> toolLearningFeedback({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolLearningFeedback(\$input: Map) { toolLearningFeedback(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolLearningFeedback'];
  }

  Future<Object?> toolLearningReflect({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolLearningReflect(\$input: Map) { toolLearningReflect(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolLearningReflect'];
  }

  Future<Object?> toolLearningEvolve({Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolLearningEvolve(\$input: Map) { toolLearningEvolve(input: \$input) }', variables: {
      'input': ?input,
    }
    );
    return data['toolLearningEvolve'];
  }

  Future<Object?> toolCancelJob({required String id}) async {
    final data = await _executor.mutation('mutation toolCancelJob(\$id: String!) { toolCancelJob(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['toolCancelJob'];
  }

  Future<Object?> toolApproveCases({required String name_, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolApproveCases(\$name: String!, \$input: Map) { toolApproveCases(name: \$name, input: \$input) }', variables: {
      'name': name_,
      'input': ?input,
    }
    );
    return data['toolApproveCases'];
  }

  Future<Object?> toolRemoveCase({required String name_, required String hash}) async {
    final data = await _executor.mutation('mutation toolRemoveCase(\$name: String!, \$hash: String!) { toolRemoveCase(name: \$name, hash: \$hash) }', variables: {
      'name': name_,
      'hash': hash,
    }
    );
    return data['toolRemoveCase'];
  }

  Future<Object?> toolRunSuite({required String name_, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation toolRunSuite(\$name: String!, \$input: Map) { toolRunSuite(name: \$name, input: \$input) }', variables: {
      'name': name_,
      'input': ?input,
    }
    );
    return data['toolRunSuite'];
  }

}
