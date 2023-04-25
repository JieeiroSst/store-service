from abc import ABC, abstractmethod
from datetime import datetime

class Clock(ABC):
    @abstractmethod
    def get_current_date(self) -> datetime:
        pass

class SystemClock(Clock):
    def get_current_date(self) -> datetime:
        return datetime.utcnow()
    
class FixedClock(Clock):
    def __init__(self, date: datetime) -> None:
        self._date = date
    
    def get_current_date(self) -> datetime:
        return self._date