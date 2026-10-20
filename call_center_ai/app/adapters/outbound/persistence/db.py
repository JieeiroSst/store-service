from __future__ import annotations

import logging

from sqlalchemy import create_engine, inspect, text
from sqlalchemy.engine import Engine
from sqlalchemy.orm import sessionmaker

from app.adapters.outbound.persistence.orm import Base

log = logging.getLogger(__name__)


def build_engine(url: str) -> Engine:
    if url.startswith("sqlite"):
        return create_engine(url, connect_args={"check_same_thread": False})
    return create_engine(
        url,
        pool_size=10,
        max_overflow=20,
        pool_pre_ping=True,
        pool_recycle=3600,
    )


def build_session_factory(engine: Engine) -> sessionmaker:
    return sessionmaker(bind=engine, autoflush=False, expire_on_commit=False)


def create_schema(engine: Engine) -> None:
    Base.metadata.create_all(bind=engine)
    add_missing_columns(engine)


def add_missing_columns(engine: Engine) -> None:
    inspector = inspect(engine)
    preparer = engine.dialect.identifier_preparer
    with engine.begin() as conn:
        for table in Base.metadata.sorted_tables:
            existing = {c["name"] for c in inspector.get_columns(table.name)}
            for column in table.columns:
                if column.name in existing:
                    continue
                ddl = "ALTER TABLE {} ADD COLUMN {} {} NULL".format(
                    preparer.quote(table.name),
                    preparer.quote(column.name),
                    column.type.compile(dialect=engine.dialect),
                )
                log.info("migrating schema: %s", ddl)
                conn.execute(text(ddl))
