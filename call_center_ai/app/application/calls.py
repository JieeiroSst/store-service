from __future__ import annotations

from datetime import datetime, timedelta
from typing import Any, Callable, List, Mapping, Optional

from app.domain.errors import NotFoundError
from app.domain.models import Call, CallOutcome, CallStatistics, CallStatus, Customer
from app.domain.ports import CallRepositoryPort, CustomerRepositoryPort

CUSTOMER_EDITABLE_FIELDS = ("name", "email", "notes")


class CallService:
    def __init__(
        self,
        calls: CallRepositoryPort,
        customers: CustomerRepositoryPort,
        clock: Callable[[], datetime] = datetime.utcnow,
    ) -> None:
        self._calls = calls
        self._customers = customers
        self._clock = clock

    def list_calls(
        self,
        skip: int = 0,
        limit: int = 100,
        status: Optional[CallStatus] = None,
        from_number: Optional[str] = None,
        outcome: Optional[CallOutcome] = None,
    ) -> List[Call]:
        return self._calls.list(skip, limit, status=status, from_number=from_number, outcome=outcome)

    def get_call(self, call_id: int) -> Call:
        call = self._calls.get_with_messages(call_id)
        if call is None:
            raise NotFoundError("Call not found")
        return call

    def delete_call(self, call_id: int) -> None:
        if not self._calls.delete(call_id):
            raise NotFoundError("Call not found")

    def statistics(self, days: int = 30) -> CallStatistics:
        return self._calls.statistics(self._clock() - timedelta(days=days))

    def get_customer(self, phone_number: str) -> Customer:
        customer = self._customers.get_by_phone(phone_number)
        if customer is None:
            raise NotFoundError("Customer not found")
        return customer

    def update_customer(self, phone_number: str, changes: Mapping[str, Any]) -> Customer:
        customer = self.get_customer(phone_number)
        for name in CUSTOMER_EDITABLE_FIELDS:
            if name in changes:
                setattr(customer, name, changes[name])
        return self._customers.save(customer)
