from kink import di

from identified_service.building_blocks.db import get_mongo_database
from identified_service.clients.application.client_service import ClientService
from identified_service.clients.application.exporter_factory import Destination, ExporterFactory
from identified_service.clients.domain.client_repository import IClientRepository
from identified_service.clients.domain.clients_exporter import IClientsExporter
from identified_service.clients.infrastructure.mongo_client_repository import MongoDBClientRepository
from identified_service.gym_passes.facade import GymPassFacade


def bootstrap_di() -> None:
    repository = MongoDBClientRepository(get_mongo_database())
    clients_exporter = ExporterFactory.build(Destination.S3)

    di[IClientRepository] = repository
    di[IClientsExporter] = clients_exporter
    di[ClientService] = ClientService(repository, di[GymPassFacade], clients_exporter)