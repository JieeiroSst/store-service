from sqlalchemy.orm import Session
from sqlalchemy import func, and_
from typing import List, Optional
from datetime import datetime, timedelta
import models
import schemas

# Call CRUD operations
def create_call(db: Session, call: schemas.CallCreate) -> models.Call:
    """Tạo cuộc gọi mới"""
    db_call = models.Call(**call.dict())
    db.add(db_call)
    db.commit()
    db.refresh(db_call)
    return db_call

def get_call_by_sid(db: Session, call_sid: str) -> Optional[models.Call]:
    """Lấy cuộc gọi theo call_sid"""
    return db.query(models.Call).filter(models.Call.call_sid == call_sid).first()

def get_call(db: Session, call_id: int) -> Optional[models.Call]:
    """Lấy cuộc gọi theo ID"""
    return db.query(models.Call).filter(models.Call.id == call_id).first()

def get_calls(
    db: Session, 
    skip: int = 0, 
    limit: int = 100,
    status: Optional[models.CallStatus] = None,
    from_number: Optional[str] = None
) -> List[models.Call]:
    """Lấy danh sách cuộc gọi với filter"""
    query = db.query(models.Call)
    
    if status:
        query = query.filter(models.Call.status == status)
    
    if from_number:
        query = query.filter(models.Call.from_number == from_number)
    
    return query.order_by(models.Call.created_at.desc()).offset(skip).limit(limit).all()

def update_call(db: Session, call_id: int, call_update: schemas.CallUpdate) -> Optional[models.Call]:
    """Cập nhật thông tin cuộc gọi"""
    db_call = get_call(db, call_id)
    if not db_call:
        return None
    
    update_data = call_update.dict(exclude_unset=True)
    for field, value in update_data.items():
        setattr(db_call, field, value)
    
    db_call.updated_at = datetime.utcnow()
    db.commit()
    db.refresh(db_call)
    return db_call

def delete_call(db: Session, call_id: int) -> bool:
    """Xóa cuộc gọi"""
    db_call = get_call(db, call_id)
    if not db_call:
        return False
    
    db.delete(db_call)
    db.commit()
    return True

# Message CRUD operations
def create_message(db: Session, message: schemas.MessageCreate) -> models.Message:
    """Tạo message mới"""
    db_message = models.Message(**message.dict())
    db.add(db_message)
    db.commit()
    db.refresh(db_message)
    return db_message

def get_messages_by_call(db: Session, call_id: int) -> List[models.Message]:
    """Lấy tất cả messages của một cuộc gọi"""
    return db.query(models.Message)\
        .filter(models.Message.call_id == call_id)\
        .order_by(models.Message.timestamp.asc())\
        .all()

def get_call_with_messages(db: Session, call_id: int) -> Optional[models.Call]:
    """Lấy cuộc gọi kèm tất cả messages"""
    return db.query(models.Call)\
        .filter(models.Call.id == call_id)\
        .first()

# Scenario CRUD operations
def create_scenario(db: Session, scenario: schemas.ScenarioCreate) -> models.Scenario:
    """Tạo kịch bản mới"""
    db_scenario = models.Scenario(**scenario.dict())
    db.add(db_scenario)
    db.commit()
    db.refresh(db_scenario)
    return db_scenario

def get_scenario(db: Session, scenario_id: int) -> Optional[models.Scenario]:
    """Lấy kịch bản theo ID"""
    return db.query(models.Scenario).filter(models.Scenario.id == scenario_id).first()

def get_scenario_by_name(db: Session, name: str) -> Optional[models.Scenario]:
    """Lấy kịch bản theo tên"""
    return db.query(models.Scenario).filter(models.Scenario.name == name).first()

def get_scenarios(db: Session, skip: int = 0, limit: int = 100, active_only: bool = True) -> List[models.Scenario]:
    """Lấy danh sách kịch bản"""
    query = db.query(models.Scenario)
    
    if active_only:
        query = query.filter(models.Scenario.is_active == 1)
    
    return query.order_by(models.Scenario.created_at.desc()).offset(skip).limit(limit).all()

def update_scenario(db: Session, scenario_id: int, scenario_update: schemas.ScenarioUpdate) -> Optional[models.Scenario]:
    """Cập nhật kịch bản"""
    db_scenario = get_scenario(db, scenario_id)
    if not db_scenario:
        return None
    
    update_data = scenario_update.dict(exclude_unset=True)
    for field, value in update_data.items():
        setattr(db_scenario, field, value)
    
    db_scenario.updated_at = datetime.utcnow()
    db.commit()
    db.refresh(db_scenario)
    return db_scenario

def delete_scenario(db: Session, scenario_id: int) -> bool:
    """Xóa kịch bản"""
    db_scenario = get_scenario(db, scenario_id)
    if not db_scenario:
        return False
    
    db.delete(db_scenario)
    db.commit()
    return True

# Customer CRUD operations
def create_customer(db: Session, customer: schemas.CustomerCreate) -> models.Customer:
    """Tạo khách hàng mới"""
    db_customer = models.Customer(**customer.dict())
    db.add(db_customer)
    db.commit()
    db.refresh(db_customer)
    return db_customer

def get_customer_by_phone(db: Session, phone_number: str) -> Optional[models.Customer]:
    """Lấy khách hàng theo số điện thoại"""
    return db.query(models.Customer).filter(models.Customer.phone_number == phone_number).first()

def get_or_create_customer(db: Session, phone_number: str) -> models.Customer:
    """Lấy hoặc tạo mới khách hàng"""
    customer = get_customer_by_phone(db, phone_number)
    if not customer:
        customer_data = schemas.CustomerCreate(phone_number=phone_number)
        customer = create_customer(db, customer_data)
    return customer

def update_customer(db: Session, customer_id: int, customer_update: schemas.CustomerUpdate) -> Optional[models.Customer]:
    """Cập nhật thông tin khách hàng"""
    db_customer = db.query(models.Customer).filter(models.Customer.id == customer_id).first()
    if not db_customer:
        return None
    
    update_data = customer_update.dict(exclude_unset=True)
    for field, value in update_data.items():
        setattr(db_customer, field, value)
    
    db_customer.updated_at = datetime.utcnow()
    db.commit()
    db.refresh(db_customer)
    return db_customer

def increment_customer_calls(db: Session, phone_number: str):
    """Tăng số lượng cuộc gọi của khách hàng"""
    customer = get_customer_by_phone(db, phone_number)
    if customer:
        customer.total_calls += 1
        customer.last_call_date = datetime.utcnow()
        db.commit()

# Analytics operations
def get_call_statistics(db: Session, days: int = 30) -> dict:
    """Lấy thống kê cuộc gọi"""
    start_date = datetime.utcnow() - timedelta(days=days)
    
    # Tổng số cuộc gọi
    total_calls = db.query(func.count(models.Call.id))\
        .filter(models.Call.created_at >= start_date)\
        .scalar()
    
    # Cuộc gọi hoàn thành
    completed_calls = db.query(func.count(models.Call.id))\
        .filter(
            and_(
                models.Call.created_at >= start_date,
                models.Call.status == models.CallStatus.COMPLETED
            )
        )\
        .scalar()
    
    # Cuộc gọi thất bại
    failed_calls = db.query(func.count(models.Call.id))\
        .filter(
            and_(
                models.Call.created_at >= start_date,
                models.Call.status.in_([models.CallStatus.FAILED, models.CallStatus.NO_ANSWER, models.CallStatus.BUSY])
            )
        )\
        .scalar()
    
    # Thời lượng trung bình
    avg_duration = db.query(func.avg(models.Call.duration))\
        .filter(
            and_(
                models.Call.created_at >= start_date,
                models.Call.status == models.CallStatus.COMPLETED
            )
        )\
        .scalar() or 0
    
    # Tổng số tin nhắn
    total_messages = db.query(func.count(models.Message.id))\
        .join(models.Call)\
        .filter(models.Call.created_at >= start_date)\
        .scalar()
    
    return {
        "total_calls": total_calls or 0,
        "completed_calls": completed_calls or 0,
        "failed_calls": failed_calls or 0,
        "average_duration": round(float(avg_duration), 2),
        "total_messages": total_messages or 0
    }
