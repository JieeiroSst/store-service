"""
Script test các thành phần của hệ thống
"""
import asyncio
import sys
from dotenv import load_dotenv

load_dotenv()

async def test_database():
    """Test kết nối database"""
    print("\n🔍 Testing Database Connection...")
    try:
        from database import engine, init_db
        from sqlalchemy import text
        
        init_db()
        
        with engine.connect() as conn:
            result = conn.execute(text("SELECT 1"))
            assert result.fetchone()[0] == 1
        
        print("✅ Database connection: OK")
        return True
    except Exception as e:
        print(f"❌ Database connection failed: {e}")
        return False

async def test_ai_service():
    """Test AI service"""
    print("\n🔍 Testing AI Service...")
    try:
        from ai_service import get_ai_manager
        
        ai_manager = get_ai_manager()
        
        test_messages = [
            {"role": "user", "content": "Xin chào"}
        ]
        
        response, time_taken = await ai_manager.generate_response(
            test_messages,
            "Bạn là trợ lý thân thiện."
        )
        
        if response:
            print(f"✅ AI Service: OK")
            print(f"   Response: {response[:100]}...")
            print(f"   Time: {time_taken:.2f}s")
            return True
        else:
            print("❌ AI Service: No response")
            return False
            
    except Exception as e:
        print(f"❌ AI Service failed: {e}")
        return False

async def test_tts_service():
    """Test TTS service"""
    print("\n🔍 Testing TTS Service...")
    try:
        from tts_service import get_tts_manager
        
        tts_manager = get_tts_manager()
        
        audio_content = await tts_manager.synthesize(
            "Xin chào, đây là test text to speech",
            "vi-VN"
        )
        
        if audio_content:
            print(f"✅ TTS Service: OK")
            print(f"   Audio size: {len(audio_content)} bytes")
            
            # Lưu file test
            filename = "test_audio.mp3"
            saved = tts_manager.save_audio(audio_content, filename)
            if saved:
                print(f"   Saved to: {filename}")
            
            return True
        else:
            print("❌ TTS Service: No audio generated")
            return False
            
    except Exception as e:
        print(f"❌ TTS Service failed: {e}")
        return False

async def test_scenarios():
    """Test scenario management"""
    print("\n🔍 Testing Scenario Management...")
    try:
        from ai_service import ScenarioManager
        
        scenarios = ScenarioManager.list_scenarios()
        
        print(f"✅ Scenarios: OK")
        print(f"   Available scenarios: {len(scenarios)}")
        for key, name in scenarios.items():
            print(f"   - {key}: {name}")
        
        return True
            
    except Exception as e:
        print(f"❌ Scenarios failed: {e}")
        return False

async def test_api_endpoints():
    """Test API endpoints"""
    print("\n🔍 Testing API Endpoints...")
    try:
        import httpx
        
        async with httpx.AsyncClient() as client:
            # Test root endpoint
            response = await client.get("http://localhost:8000/")
            
            if response.status_code == 200:
                data = response.json()
                print(f"✅ API Endpoints: OK")
                print(f"   Status: {data.get('status')}")
                print(f"   Service: {data.get('service')}")
                print(f"   Version: {data.get('version')}")
                return True
            else:
                print(f"❌ API returned status: {response.status_code}")
                return False
                
    except Exception as e:
        print(f"❌ API Endpoints failed: {e}")
        print("   Note: Make sure the server is running (python main.py)")
        return False

async def run_all_tests():
    """Chạy tất cả tests"""
    print("=" * 60)
    print("🚀 CALL CENTER AI SYSTEM - TESTING")
    print("=" * 60)
    
    results = []
    
    # Test database
    results.append(("Database", await test_database()))
    
    # Test AI service
    results.append(("AI Service", await test_ai_service()))
    
    # Test TTS service
    results.append(("TTS Service", await test_tts_service()))
    
    # Test scenarios
    results.append(("Scenarios", await test_scenarios()))
    
    # Test API (optional - chỉ chạy nếu server đang running)
    if len(sys.argv) > 1 and sys.argv[1] == "--with-api":
        results.append(("API Endpoints", await test_api_endpoints()))
    
    # Summary
    print("\n" + "=" * 60)
    print("📊 TEST SUMMARY")
    print("=" * 60)
    
    passed = sum(1 for _, result in results if result)
    total = len(results)
    
    for name, result in results:
        status = "✅ PASSED" if result else "❌ FAILED"
        print(f"{name:20} {status}")
    
    print("-" * 60)
    print(f"Total: {passed}/{total} tests passed")
    
    if passed == total:
        print("\n🎉 All tests passed! System is ready.")
        return 0
    else:
        print(f"\n⚠️  {total - passed} test(s) failed. Please check configuration.")
        return 1

if __name__ == "__main__":
    exit_code = asyncio.run(run_all_tests())
    sys.exit(exit_code)
