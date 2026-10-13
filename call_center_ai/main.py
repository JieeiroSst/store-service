from fastapi import FastAPI, Request, Depends, HTTPException, BackgroundTasks
from fastapi.responses import Response, JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy.orm import Session
from twilio.twiml.voice_response import VoiceResponse, Gather
from typing import Optional, List
import os
from dotenv import load_dotenv
from contextlib import asynccontextmanager

# Import local modules
import models
import schemas
import crud
from database import engine, get_db, init_db
from ai_service import get_ai_manager, ScenarioManager
from auth_deps import require_admin
from twilio_auth import verify_twilio
from tts_service import get_tts_manager

load_dotenv()

# Khởi tạo database
models.Base.metadata.create_all(bind=engine)

# Context manager cho lifecycle
@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifecycle events"""
    # Startup
    print("🚀 Starting Call Center AI System...")
    init_db()
    
    # Khởi tạo AI và TTS managers
    try:
        get_ai_manager()
        get_tts_manager()
        print("✅ All services initialized successfully!")
    except Exception as e:
        print(f"⚠️  Warning: Some services failed to initialize: {e}")
    
    yield
    
    # Shutdown
    print("👋 Shutting down Call Center AI System...")

# Tạo FastAPI app
app = FastAPI(
    title="Call Center AI API",
    description="Hệ thống call center với AI tự động trả lời",
    version="1.0.0",
    lifespan=lifespan
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Biến toàn cục lưu conversation history
conversation_history = {}

# ==================== TWILIO VOICE ROUTES ====================

@app.post("/voice/incoming", dependencies=[Depends(verify_twilio)])
async def handle_incoming_call(request: Request, db: Session = Depends(get_db)):
    """Xử lý cuộc gọi đến"""
    form_data = await request.form()
    
    call_sid = form_data.get("CallSid")
    from_number = form_data.get("From")
    to_number = form_data.get("To")
    
    print(f"📞 Incoming call from {from_number} (CallSid: {call_sid})")
    
    # Tạo hoặc lấy khách hàng
    crud.get_or_create_customer(db, from_number)
    
    # Tạo call record
    call_data = schemas.CallCreate(
        call_sid=call_sid,
        from_number=from_number,
        to_number=to_number,
        notes="Incoming call"
    )
    db_call = crud.create_call(db, call_data)
    
    # Khởi tạo conversation history
    conversation_history[call_sid] = []
    
    # Tạo TwiML response
    response = VoiceResponse()
    
    # Lời chào
    greeting = "Xin chào! Tôi là trợ lý ảo của công ty. Tôi có thể giúp gì cho bạn?"
    
    # Lưu message
    message_data = schemas.MessageCreate(
        call_id=db_call.id,
        role=models.MessageRole.ASSISTANT,
        content=greeting
    )
    crud.create_message(db, message_data)
    
    # Thêm vào conversation history
    conversation_history[call_sid].append({
        "role": "assistant",
        "content": greeting
    })
    
    # Gather input từ người dùng
    gather = Gather(
        input="speech",
        action="/voice/process-speech",
        language="vi-VN",
        timeout=5,
        speech_timeout="auto"
    )
    gather.say(greeting, language="vi-VN")
    response.append(gather)
    
    # Nếu không có input
    response.say("Tôi không nghe thấy bạn nói gì. Vui lòng thử lại sau.", language="vi-VN")
    response.hangup()
    
    return Response(content=str(response), media_type="application/xml")

@app.post("/voice/process-speech", dependencies=[Depends(verify_twilio)])
async def process_speech(request: Request, db: Session = Depends(get_db)):
    """Xử lý speech input từ người dùng"""
    form_data = await request.form()
    
    call_sid = form_data.get("CallSid")
    speech_result = form_data.get("SpeechResult", "")
    
    print(f"🎤 User said: {speech_result}")
    
    if not speech_result:
        response = VoiceResponse()
        response.say("Tôi không nghe rõ. Bạn có thể nói lại không?", language="vi-VN")
        
        gather = Gather(
            input="speech",
            action="/voice/process-speech",
            language="vi-VN",
            timeout=5,
            speech_timeout="auto"
        )
        response.append(gather)
        return Response(content=str(response), media_type="application/xml")
    
    # Lấy call record
    db_call = crud.get_call_by_sid(db, call_sid)
    if not db_call:
        response = VoiceResponse()
        response.say("Xin lỗi, có lỗi xảy ra. Vui lòng gọi lại sau.", language="vi-VN")
        response.hangup()
        return Response(content=str(response), media_type="application/xml")
    
    # Cập nhật status
    crud.update_call(db, db_call.id, schemas.CallUpdate(status=models.CallStatus.IN_PROGRESS))
    
    # Lưu user message
    user_message = schemas.MessageCreate(
        call_id=db_call.id,
        role=models.MessageRole.USER,
        content=speech_result
    )
    crud.create_message(db, user_message)
    
    # Thêm vào conversation history
    if call_sid not in conversation_history:
        conversation_history[call_sid] = []
    
    conversation_history[call_sid].append({
        "role": "user",
        "content": speech_result
    })
    
    # Tạo AI response
    ai_manager = get_ai_manager()
    scenario_prompt = ScenarioManager.get_scenario_prompt("customer_support")
    
    ai_response, processing_time = await ai_manager.generate_response(
        conversation_history[call_sid],
        scenario_prompt
    )
    
    if not ai_response:
        ai_response = "Xin lỗi, tôi đang gặp sự cố kỹ thuật. Bạn có thể thử lại sau được không?"
    
    print(f"🤖 AI response: {ai_response}")
    
    # Lưu assistant message
    assistant_message = schemas.MessageCreate(
        call_id=db_call.id,
        role=models.MessageRole.ASSISTANT,
        content=ai_response,
        processing_time=processing_time
    )
    crud.create_message(db, assistant_message)
    
    # Thêm vào conversation history
    conversation_history[call_sid].append({
        "role": "assistant",
        "content": ai_response
    })
    
    # Kiểm tra xem có phải câu kết thúc không
    end_phrases = ["tạm biệt", "cảm ơn", "kết thúc", "bye", "goodbye"]
    should_end = any(phrase in speech_result.lower() for phrase in end_phrases)
    
    # Tạo TwiML response
    response = VoiceResponse()
    
    if should_end:
        response.say(ai_response, language="vi-VN")
        response.say("Cảm ơn bạn đã gọi. Chúc bạn một ngày tốt lành!", language="vi-VN")
        response.hangup()
        
        # Cập nhật call status
        crud.update_call(db, db_call.id, schemas.CallUpdate(
            status=models.CallStatus.COMPLETED,
            end_time=models.datetime.utcnow()
        ))
        
        # Tăng số lượng cuộc gọi của customer
        crud.increment_customer_calls(db, db_call.from_number)
        
    else:
        # Tiếp tục hội thoại
        gather = Gather(
            input="speech",
            action="/voice/process-speech",
            language="vi-VN",
            timeout=5,
            speech_timeout="auto"
        )
        gather.say(ai_response, language="vi-VN")
        response.append(gather)
        
        # Nếu timeout
        response.say("Cảm ơn bạn đã gọi. Hẹn gặp lại!", language="vi-VN")
        response.hangup()
    
    return Response(content=str(response), media_type="application/xml")

@app.post("/voice/status", dependencies=[Depends(verify_twilio)])
async def call_status_callback(request: Request, db: Session = Depends(get_db)):
    """Callback khi trạng thái cuộc gọi thay đổi"""
    form_data = await request.form()
    
    call_sid = form_data.get("CallSid")
    call_status = form_data.get("CallStatus")
    call_duration = form_data.get("CallDuration", 0)
    
    print(f"📊 Call {call_sid} status: {call_status}, duration: {call_duration}s")
    
    db_call = crud.get_call_by_sid(db, call_sid)
    if db_call:
        status_map = {
            "completed": models.CallStatus.COMPLETED,
            "busy": models.CallStatus.BUSY,
            "no-answer": models.CallStatus.NO_ANSWER,
            "failed": models.CallStatus.FAILED,
            "canceled": models.CallStatus.FAILED
        }
        
        update_data = schemas.CallUpdate(
            status=status_map.get(call_status, models.CallStatus.FAILED),
            duration=int(call_duration),
            end_time=models.datetime.utcnow()
        )
        crud.update_call(db, db_call.id, update_data)
        
        # Cleanup conversation history
        if call_sid in conversation_history:
            del conversation_history[call_sid]
    
    return {"status": "ok"}

# ==================== REST API ROUTES ====================

@app.get("/")
async def root():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "Call Center AI",
        "version": "1.0.0"
    }

# Call management endpoints
@app.get("/api/calls", response_model=List[schemas.CallResponse], dependencies=[Depends(require_admin)])
async def list_calls(
    skip: int = 0,
    limit: int = 100,
    status: Optional[models.CallStatus] = None,
    from_number: Optional[str] = None,
    db: Session = Depends(get_db)
):
    """Lấy danh sách cuộc gọi"""
    calls = crud.get_calls(db, skip=skip, limit=limit, status=status, from_number=from_number)
    return calls

@app.get("/api/calls/{call_id}", response_model=schemas.CallHistoryResponse, dependencies=[Depends(require_admin)])
async def get_call_detail(call_id: int, db: Session = Depends(get_db)):
    """Lấy chi tiết cuộc gọi kèm lịch sử hội thoại"""
    call = crud.get_call_with_messages(db, call_id)
    if not call:
        raise HTTPException(status_code=404, detail="Call not found")
    return call

@app.delete("/api/calls/{call_id}", dependencies=[Depends(require_admin)])
async def delete_call(call_id: int, db: Session = Depends(get_db)):
    """Xóa cuộc gọi"""
    success = crud.delete_call(db, call_id)
    if not success:
        raise HTTPException(status_code=404, detail="Call not found")
    return {"message": "Call deleted successfully"}

# Scenario management endpoints
@app.get("/api/scenarios", response_model=List[schemas.ScenarioResponse], dependencies=[Depends(require_admin)])
async def list_scenarios(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    """Lấy danh sách kịch bản"""
    scenarios = crud.get_scenarios(db, skip=skip, limit=limit)
    return scenarios

@app.post("/api/scenarios", response_model=schemas.ScenarioResponse, dependencies=[Depends(require_admin)])
async def create_scenario(scenario: schemas.ScenarioCreate, db: Session = Depends(get_db)):
    """Tạo kịch bản mới"""
    # Kiểm tra trùng tên
    existing = crud.get_scenario_by_name(db, scenario.name)
    if existing:
        raise HTTPException(status_code=400, detail="Scenario with this name already exists")
    
    return crud.create_scenario(db, scenario)

@app.get("/api/scenarios/{scenario_id}", response_model=schemas.ScenarioResponse, dependencies=[Depends(require_admin)])
async def get_scenario(scenario_id: int, db: Session = Depends(get_db)):
    """Lấy chi tiết kịch bản"""
    scenario = crud.get_scenario(db, scenario_id)
    if not scenario:
        raise HTTPException(status_code=404, detail="Scenario not found")
    return scenario

@app.put("/api/scenarios/{scenario_id}", response_model=schemas.ScenarioResponse, dependencies=[Depends(require_admin)])
async def update_scenario(
    scenario_id: int,
    scenario_update: schemas.ScenarioUpdate,
    db: Session = Depends(get_db)
):
    """Cập nhật kịch bản"""
    scenario = crud.update_scenario(db, scenario_id, scenario_update)
    if not scenario:
        raise HTTPException(status_code=404, detail="Scenario not found")
    return scenario

@app.delete("/api/scenarios/{scenario_id}", dependencies=[Depends(require_admin)])
async def delete_scenario(scenario_id: int, db: Session = Depends(get_db)):
    """Xóa kịch bản"""
    success = crud.delete_scenario(db, scenario_id)
    if not success:
        raise HTTPException(status_code=404, detail="Scenario not found")
    return {"message": "Scenario deleted successfully"}

# Customer management endpoints
@app.get("/api/customers/{phone_number}", response_model=schemas.CustomerResponse, dependencies=[Depends(require_admin)])
async def get_customer(phone_number: str, db: Session = Depends(get_db)):
    """Lấy thông tin khách hàng"""
    customer = crud.get_customer_by_phone(db, phone_number)
    if not customer:
        raise HTTPException(status_code=404, detail="Customer not found")
    return customer

# Analytics endpoints
@app.get("/api/analytics/statistics", response_model=schemas.CallStatistics, dependencies=[Depends(require_admin)])
async def get_statistics(days: int = 30, db: Session = Depends(get_db)):
    """Lấy thống kê cuộc gọi"""
    stats = crud.get_call_statistics(db, days=days)
    return stats

# Test TTS endpoint
@app.post("/api/test/tts", dependencies=[Depends(require_admin)])
async def test_tts(text: str, language: str = "vi-VN"):
    """Test text-to-speech"""
    tts_manager = get_tts_manager()
    audio_content = await tts_manager.synthesize(text, language)
    
    if audio_content:
        return Response(content=audio_content, media_type="audio/mpeg")
    else:
        raise HTTPException(status_code=500, detail="TTS failed")

if __name__ == "__main__":
    import uvicorn
    
    host = os.getenv("SERVER_HOST", "0.0.0.0")
    port = int(os.getenv("SERVER_PORT", "8000"))
    
    uvicorn.run(
        "main:app",
        host=host,
        port=port,
        reload=True,
        log_level="info"
    )
