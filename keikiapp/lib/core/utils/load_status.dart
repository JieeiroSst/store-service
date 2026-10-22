enum LoadStatus { initial, loading, success, failure }

extension LoadStatusX on LoadStatus {
  bool get isLoading => this == LoadStatus.loading;
  bool get isFailure => this == LoadStatus.failure;
  bool get isSuccess => this == LoadStatus.success;
}
