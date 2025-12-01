"""
Script để khởi tạo các kịch bản mẫu vào database
"""
from sqlalchemy.orm import Session
from database import SessionLocal, init_db
from ai_service import ScenarioManager
import models
import schemas
import crud

def init_default_scenarios():
    """Khởi tạo các kịch bản mặc định"""
    db = SessionLocal()
    
    try:
        print("🔧 Initializing default scenarios...")
        
        scenarios = ScenarioManager.DEFAULT_SCENARIOS
        
        for key, scenario_data in scenarios.items():
            # Kiểm tra xem scenario đã tồn tại chưa
            existing = crud.get_scenario_by_name(db, key)
            
            if not existing:
                scenario = schemas.ScenarioCreate(
                    name=key,
                    description=scenario_data["name"],
                    prompt=scenario_data["prompt"],
                    is_active=1
                )
                crud.create_scenario(db, scenario)
                print(f"✅ Created scenario: {key}")
            else:
                print(f"⏭️  Scenario already exists: {key}")
        
        print("✅ All default scenarios initialized successfully!")
        
    except Exception as e:
        print(f"❌ Error initializing scenarios: {e}")
        db.rollback()
    finally:
        db.close()

if __name__ == "__main__":
    # Khởi tạo database
    init_db()
    
    # Khởi tạo scenarios
    init_default_scenarios()
