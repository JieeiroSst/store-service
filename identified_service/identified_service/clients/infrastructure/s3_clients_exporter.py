from botocore.exceptions import ClientError

from identified_service.building_blocks.custom_types import S3SdkClient
from identified_service.clients.domain.clients_exporter import IClientsExporter
from identified_service.clients.domain.errors import ExportError
from identified_service.clients.domain.report import Report


class S3ClientsExporter(IClientsExporter):
    def __init__(self, s3_sdk_client: S3SdkClient, bucket_name: str) -> None:
        self._s3_sdk_client = s3_sdk_client
        self._bucket_name = bucket_name

    def export(self, report: Report) -> None:
        try:
            self._s3_sdk_client.put_object(
                Bucket=self._bucket_name, Key=f"reports/{report.file_name}", Body=report.content.read()
            )
        except ClientError as error:
            raise ExportError("Can not export clients report!") from error