from abc import ABC, abstractmethod

from identified_service.clients.domain.client import Client
from identified_service.clients.domain.client_id import ClientId


class IClientRepository(ABC):
    @abstractmethod
    def get(self, client_id: ClientId) -> Client:
        pass

    @abstractmethod
    def get_all(self) -> list[Client]:
        pass

    @abstractmethod
    def save(self, client: Client) -> None:
        pass