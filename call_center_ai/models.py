from sqlalchemy import Column, Integer, String, DateTime, Text, Float, ForeignKey, Enum
from sqlalchemy.orm import relationship
from datetime import datetime
from database import Base
import enum

class CallStatus(str, enum.Enum):
    """Trạng thái cuộc gọi"""
    INITIATED = "initiated"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"
    NO_ANSWER = "no_answer"
    BUSY = "busy"

class MessageRole(str, enum.Enum):
    """Vai trò trong cuộc hội thoại"""
    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"

class Call(Base):
    """Model lưu thông tin cuộc gọi"""
    __tablename__ = "calls"

    id = Column(Integer, primary_key=True, index=True)
    call_sid = Column(String(255), unique=True, index=True, nullable=False)
    from_number = Column(String(20), nullable=False, index=True)
    to_number = Column(String(20), nullable=False)
    status = Column(Enum(CallStatus), default=CallStatus.INITIATED)
    duration = Column(Integer, default=0)  # Thời lượng cuộc gọi (giây)
    start_time = Column(DateTime, default=datetime.utcnow)
    end_time = Column(DateTime, nullable=True)
    recording_url = Column(String(512), nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    # Relationship
    messages = relationship("Message", back_populates="call", cascade="all, delete-orphan")
    
    def __repr__(self):
        return f"<Call(id={self.id}, call_sid={self.call_sid}, from={self.from_number})>"

class Message(Base):
    """Model lưu tin nhắn trong cuộc hội thoại"""
    __tablename__ = "messages"

    id = Column(Integer, primary_key=True, index=True)
    call_id = Column(Integer, ForeignKey("calls.id"), nullable=False, index=True)
    role = Column(Enum(MessageRole), nullable=False)
    content = Column(Text, nullable=False)
    audio_url = Column(String(512), nullable=True)  # URL file audio nếu có
    timestamp = Column(DateTime, default=datetime.utcnow, index=True)
    processing_time = Column(Float, nullable=True)  # Thời gian xử lý (giây)
    
    # Relationship
    call = relationship("Call", back_populates="messages")
    
    def __repr__(self):
        return f"<Message(id={self.id}, call_id={self.call_id}, role={self.role})>"

class Scenario(Base):
    """Model lưu kịch bản hội thoại"""
    __tablename__ = "scenarios"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False, unique=True, index=True)
    description = Column(Text, nullable=True)
    prompt = Column(Text, nullable=False)  # System prompt cho AI
    is_active = Column(Integer, default=1)  # 1: active, 0: inactive
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    def __repr__(self):
        return f"<Scenario(id={self.id}, name={self.name})>"

class Customer(Base):
    """Model lưu thông tin khách hàng"""
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
    
    def __repr__(self):
        return f"<Customer(id={self.id}, phone={self.phone_number})>"
