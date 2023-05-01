from abc import ABC, abstractmethod

from identified_service.clients.domain.report import Report


class IClientsExporter(ABC):
    @abstractmethod
    def export(self, report: Report) -> None:
        pass