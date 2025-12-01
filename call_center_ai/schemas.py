from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List
from models import CallStatus, MessageRole

# Call Schemas
class CallBase(BaseModel):
    from_number: str
    to_number: str
    notes: Optional[str] = None

class CallCreate(CallBase):
    call_sid: str

class CallUpdate(BaseModel):
    status: Optional[CallStatus] = None
    duration: Optional[int] = None
    end_time: Optional[datetime] = None
    recording_url: Optional[str] = None
    notes: Optional[str] = None

class CallResponse(CallBase):
    id: int
    call_sid: str
    status: CallStatus
    duration: int
    start_time: datetime
    end_time: Optional[datetime]
    recording_url: Optional[str]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

# Message Schemas
class MessageBase(BaseModel):
    role: MessageRole
    content: str
    audio_url: Optional[str] = None

class MessageCreate(MessageBase):
    call_id: int

class MessageResponse(MessageBase):
    id: int
    call_id: int
    timestamp: datetime
    processing_time: Optional[float]

    class Config:
        from_attributes = True

# Scenario Schemas
class ScenarioBase(BaseModel):
    name: str
    description: Optional[str] = None
    prompt: str
    is_active: int = 1

class ScenarioCreate(ScenarioBase):
    pass

class ScenarioUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    prompt: Optional[str] = None
    is_active: Optional[int] = None

class ScenarioResponse(ScenarioBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

# Customer Schemas
class CustomerBase(BaseModel):
    phone_number: str
    name: Optional[str] = None
    email: Optional[str] = None
    notes: Optional[str] = None

class CustomerCreate(CustomerBase):
    pass

class CustomerUpdate(BaseModel):
    name: Optional[str] = None
    email: Optional[str] = None
    notes: Optional[str] = None

class CustomerResponse(CustomerBase):
    id: int
    total_calls: int
    last_call_date: Optional[datetime]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

# Call History with Messages
class CallHistoryResponse(CallResponse):
    messages: List[MessageResponse] = []

    class Config:
        from_attributes = True

# Analytics Schemas
class CallStatistics(BaseModel):
    total_calls: int
    completed_calls: int
    failed_calls: int
    average_duration: float
    total_messages: int
