from __future__ import annotations

import logging

from app.config import settings
from app.container import Container

log = logging.getLogger(__name__)


def main() -> None:
    logging.basicConfig(level=settings.log_level)
    container = Container.from_settings(settings)
    container.startup()
    try:
        scenarios = container.scenario_service().seed_defaults()
        log.info("created scenarios: %s", ", ".join(scenarios) or "none (all exist)")
        entries = container.knowledge_service().seed_defaults()
        log.info("created %d knowledge entries", len(entries))
    finally:
        container.shutdown()


if __name__ == "__main__":
    main()
