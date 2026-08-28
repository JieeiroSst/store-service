"""Use case: dò thưởng theo (ngày, tỉnh, số vé)."""
from __future__ import annotations

from datetime import date
from typing import Optional

from app.domain.models import WinningCheckResult
from app.domain.ports.result_repository import ResultRepositoryPort
from app.domain.prize_rules import PrizeRuleConfig, check_winning


class CheckWinningUseCase:
    def __init__(self, repository: ResultRepositoryPort, rule_config: Optional[PrizeRuleConfig] = None) -> None:
        self._repository = repository
        self._rule_config = rule_config or PrizeRuleConfig()

    async def execute(self, day: date, province_code: str, ticket_number: str) -> WinningCheckResult:
        result = await self._repository.find_by_date_province(day, province_code)
        if result is None:
            return WinningCheckResult(
                won=False,
                prizes_won=[],
                detail=[],
                result_found=False,
                highest_prize=None,
            )

        return check_winning(
            region=result.region,
            prizes=result.prizes,
            ticket_number=ticket_number,
            config=self._rule_config,
        )
