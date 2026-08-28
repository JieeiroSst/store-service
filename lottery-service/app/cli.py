"""CLI backfill lịch sử & cào 1 ngày. Ví dụ:

    python -m app.cli backfill --start 2026-01-01 --end 2026-01-31
    python -m app.cli backfill --start 2026-01-01 --end 2026-01-31 --force
    python -m app.cli scrape --date 2026-08-26
    python -m app.cli scrape                        # mặc định: hôm nay theo giờ VN

Lệnh `scrape` không tham số ngày là lệnh mà k8s CronJob
(chart/lottery-service/templates/scrape-cronjob.yaml) chạy mỗi ngày.
"""
from __future__ import annotations

import argparse
import asyncio
import logging
from datetime import date, datetime
from zoneinfo import ZoneInfo

from app.config import settings
from app.container import Container

logging.basicConfig(level=settings.log_level, format="%(asctime)s %(levelname)s [%(name)s] %(message)s")
logger = logging.getLogger(__name__)


def _parse_date(value: str) -> date:
    return datetime.strptime(value, "%Y-%m-%d").date()


async def _run_backfill(start: date, end: date, force: bool) -> None:
    container = Container(settings)
    await container.startup()
    try:
        use_case = container.build_backfill_use_case()
        summary = await use_case.execute(start, end, force=force)
        logger.info(
            "Backfill xong: processed=%d skipped=%d errors=%d",
            summary.days_processed,
            summary.days_skipped,
            len(summary.errors),
        )
        for err in summary.errors:
            logger.warning("Backfill error: %s", err)
    finally:
        await container.disconnect()


async def _run_scrape(day: date) -> None:
    container = Container(settings)
    await container.startup()
    try:
        use_case = container.build_scrape_results_use_case()
        summary = await use_case.execute(day)
        logger.info(
            "Scrape xong: ok=%s failed=%s provinces_saved=%d",
            summary.regions_ok,
            summary.regions_failed,
            summary.provinces_saved,
        )
    finally:
        await container.disconnect()


def main() -> None:
    parser = argparse.ArgumentParser(description="Lottery service CLI")
    subparsers = parser.add_subparsers(dest="command", required=True)

    backfill_parser = subparsers.add_parser("backfill", help="Cào lịch sử theo khoảng ngày")
    backfill_parser.add_argument("--start", required=True, type=_parse_date)
    backfill_parser.add_argument("--end", required=True, type=_parse_date)
    backfill_parser.add_argument("--force", action="store_true", help="Cào lại kể cả ngày đã có dữ liệu")

    scrape_parser = subparsers.add_parser("scrape", help="Cào 1 ngày cụ thể (mặc định: hôm nay giờ VN)")
    scrape_parser.add_argument("--date", required=False, default=None, type=_parse_date)

    args = parser.parse_args()

    if args.command == "backfill":
        asyncio.run(_run_backfill(args.start, args.end, args.force))
    elif args.command == "scrape":
        day = args.date or datetime.now(ZoneInfo(settings.scheduler_timezone)).date()
        asyncio.run(_run_scrape(day))


if __name__ == "__main__":
    main()
