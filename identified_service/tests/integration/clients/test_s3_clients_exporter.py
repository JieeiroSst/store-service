from io import StringIO

import pytest

from identified_service.building_blocks.custom_types import S3SdkClient
from identified_service.clients.domain.errors import ExportError
from identified_service.clients.domain.report import Report
from identified_service.clients.infrastructure.s3_clients_exporter import S3ClientsExporter


def test_can_export_clients_report(s3_mock: S3SdkClient) -> None:
    # given
    exporter = S3ClientsExporter(s3_mock, "client_reports")

    # when
    exporter.export(Report("test_report.csv", StringIO("first_name, last_name\nJohn, Doe")))

    # then
    s3_object = s3_mock.get_object(Bucket="client_reports", Key="reports/test_report.csv")
    assert "Body" in s3_object


def test_should_raise_an_error_if_upload_fails(s3_mock: S3SdkClient) -> None:
    # given
    exporter = S3ClientsExporter(s3_mock, "not_existing_bucket")

    # expect
    with pytest.raises(ExportError):
        exporter.export(Report("test_report.csv", StringIO("first_name, last_name\nJohn, Doe")))