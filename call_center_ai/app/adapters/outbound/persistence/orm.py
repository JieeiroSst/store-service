from __future__ import annotations

from datetime import datetime

from sqlalchemy import Column, DateTime, Enum, Float, ForeignKey, Integer, String, Text
from sqlalchemy.orm import declarative_base, relationship

from app.domain.models import CallOutcome, CallStatus, MessageRole

Base = declarative_base()


class CallRow(Base):
    __tablename__ = "calls"

    id = Column(Integer, primary_key=True, index=True)
    call_sid = Column(String(255), unique=True, index=True, nullable=False)
    from_number = Column(String(20), nullable=False, index=True)
    to_number = Column(String(20), nullable=False)
    status = Column(Enum(CallStatus), default=CallStatus.INITIATED)
    duration = Column(Integer, default=0)
    start_time = Column(DateTime, default=datetime.utcnow)
    end_time = Column(DateTime, nullable=True)
    recording_url = Column(String(512), nullable=True)
    notes = Column(Text, nullable=True)
    scenario = Column(String(255), nullable=True)
    outcome = Column(Enum(CallOutcome), nullable=True, index=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    messages = relationship(
        "MessageRow",
        back_populates="call",
        cascade="all, delete-orphan",
        order_by="MessageRow.id",
    )


class MessageRow(Base):
    __tablename__ = "messages"

    id = Column(Integer, primary_key=True, index=True)
    call_id = Column(Integer, ForeignKey("calls.id"), nullable=False, index=True)
    role = Column(Enum(MessageRole), nullable=False)
    content = Column(Text, nullable=False)
    audio_url = Column(String(512), nullable=True)
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)
    processing_time = Column(Float, nullable=True)

    call = relationship("CallRow", back_populates="messages")


class ScenarioRow(Base):
    __tablename__ = "scenarios"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False, unique=True, index=True)
    description = Column(Text, nullable=True)
    prompt = Column(Text, nullable=False)
    greeting = Column(Text, nullable=True)
    phone_number = Column(String(20), nullable=True, index=True)
    is_default = Column(Integer, default=0)
    is_active = Column(Integer, default=1)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class KnowledgeRow(Base):
    __tablename__ = "knowledge_entries"

    id = Column(Integer, primary_key=True, index=True)
    question = Column(Text, nullable=False)
    answer = Column(Text, nullable=False)
    keywords = Column(Text, nullable=True)
    category = Column(String(100), nullable=True, index=True)
    scenario = Column(String(255), nullable=True, index=True)
    is_active = Column(Integer, default=1)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class CustomerRow(Base):
    __tablename__ = "customers"

    id = Column(Integer, primary_key=True, index=True)
    phone_number = Column(String(20), unique=True, nullable=False, index=True)
    name = Column(String(255), nullable=True)
    email = Column(String(255), nullable=True)
    notes = Column(Text, nullable=True)
    total_calls = Column(Integer, default=0)
    last_call_date = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
