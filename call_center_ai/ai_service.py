import os
from typing import List, Dict, Optional
from abc import ABC, abstractmethod
from dotenv import load_dotenv
import time

load_dotenv()

class AIService(ABC):
    """Abstract base class cho AI service"""
    
    @abstractmethod
    async def generate_response(self, messages: List[Dict[str, str]], system_prompt: str = None) -> str:
        """Tạo response từ AI"""
        pass

class OpenAIService(AIService):
    """OpenAI GPT service"""
    
    def __init__(self):
        self.api_key = os.getenv("OPENAI_API_KEY")
        if self.api_key:
            try:
                from openai import AsyncOpenAI
                self.client = AsyncOpenAI(api_key=self.api_key)
                print("OpenAI service initialized successfully")
            except Exception as e:
                print(f"Failed to initialize OpenAI: {e}")
                self.client = None
        else:
            print("OpenAI API key not found")
            self.client = None
    
    async def generate_response(self, messages: List[Dict[str, str]], system_prompt: str = None) -> Optional[str]:
        """Tạo response từ OpenAI GPT"""
        if not self.client:
            return None
        
        try:
            # Thêm system prompt nếu có
            full_messages = []
            if system_prompt:
                full_messages.append({"role": "system", "content": system_prompt})
            full_messages.extend(messages)
            
            # Gọi API
            response = await self.client.chat.completions.create(
                model="gpt-4-turbo-preview",
                messages=full_messages,
                temperature=0.7,
                max_tokens=500
            )
            
            return response.choices[0].message.content
            
        except Exception as e:
            print(f"OpenAI error: {e}")
            return None

class AnthropicService(AIService):
    """Anthropic Claude service"""
    
    def __init__(self):
        self.api_key = os.getenv("ANTHROPIC_API_KEY")
        if self.api_key:
            try:
                from anthropic import AsyncAnthropic
                self.client = AsyncAnthropic(api_key=self.api_key)
                print("Anthropic service initialized successfully")
            except Exception as e:
                print(f"Failed to initialize Anthropic: {e}")
                self.client = None
        else:
            print("Anthropic API key not found")
            self.client = None
    
    async def generate_response(self, messages: List[Dict[str, str]], system_prompt: str = None) -> Optional[str]:
        """Tạo response từ Anthropic Claude"""
        if not self.client:
            return None
        
        try:
            # Claude API yêu cầu format khác một chút
            claude_messages = []
            for msg in messages:
                if msg["role"] != "system":  # Claude xử lý system prompt riêng
                    claude_messages.append({
                        "role": msg["role"],
                        "content": msg["content"]
                    })
            
            # Gọi API
            response = await self.client.messages.create(
                model="claude-3-sonnet-20240229",
                max_tokens=500,
                temperature=0.7,
                system=system_prompt or "Bạn là trợ lý AI thông minh và hữu ích.",
                messages=claude_messages
            )
            
            return response.content[0].text
            
        except Exception as e:
            print(f"Anthropic error: {e}")
            return None

class AIManager:
    """Manager để quản lý AI services"""
    
    def __init__(self):
        # Khởi tạo AI services theo thứ tự ưu tiên
        self.services = []
        
        # Thử Claude trước (tốt hơn cho tiếng Việt)
        anthropic_service = AnthropicService()
        if anthropic_service.client:
            self.services.append(anthropic_service)
        
        # Sau đó OpenAI
        openai_service = OpenAIService()
        if openai_service.client:
            self.services.append(openai_service)
        
        if not self.services:
            raise Exception("No AI service available! Please configure at least one AI service.")
        
        print(f"AI Manager initialized with {len(self.services)} service(s)")
    
    async def generate_response(
        self, 
        messages: List[Dict[str, str]], 
        system_prompt: str = None
    ) -> tuple[Optional[str], float]:
        """
        Tạo response từ AI service
        Returns: (response_text, processing_time)
        """
        start_time = time.time()
        
        for service in self.services:
            try:
                response = await service.generate_response(messages, system_prompt)
                if response:
                    processing_time = time.time() - start_time
                    return response, processing_time
            except Exception as e:
                print(f"AI service failed: {e}, trying next service...")
                continue
        
        print("All AI services failed!")
        processing_time = time.time() - start_time
        return None, processing_time

class ScenarioManager:
    """Quản lý các kịch bản hội thoại"""
    
    DEFAULT_SCENARIOS = {
        "customer_support": {
            "name": "Hỗ trợ khách hàng",
            "prompt": """Bạn là nhân viên hỗ trợ khách hàng chuyên nghiệp và thân thiện.
Nhiệm vụ của bạn:
- Lắng nghe và hiểu vấn đề của khách hàng
- Cung cấp giải pháp rõ ràng và hữu ích
- Giữ thái độ lịch sự và nhiệt tình
- Trả lời ngắn gọn, súc tích (1-2 câu)
- Nếu không có thông tin, hãy hỏi thêm hoặc hứa sẽ kiểm tra

Luôn kết thúc câu trả lời với câu hỏi để tiếp tục hội thoại."""
        },
        "sales": {
            "name": "Tư vấn bán hàng",
            "prompt": """Bạn là nhân viên tư vấn bán hàng chuyên nghiệp.
Nhiệm vụ của bạn:
- Giới thiệu sản phẩm/dịch vụ một cách hấp dẫn
- Tìm hiểu nhu cầu của khách hàng
- Đưa ra đề xuất phù hợp
- Xử lý khéo léo các từ chối
- Trả lời ngắn gọn và thuyết phục (1-2 câu)

Hãy tập trung vào lợi ích khách hàng nhận được."""
        },
        "appointment": {
            "name": "Đặt lịch hẹn",
            "prompt": """Bạn là trợ lý đặt lịch hẹn tự động.
Nhiệm vụ của bạn:
- Thu thập thông tin: họ tên, số điện thoại, thời gian mong muốn
- Xác nhận lại thông tin với khách hàng
- Thông báo lịch hẹn đã được ghi nhận
- Trả lời ngắn gọn, rõ ràng (1-2 câu)

Hãy lịch sự và chính xác trong từng bước."""
        },
        "survey": {
            "name": "Khảo sát ý kiến",
            "prompt": """Bạn là nhân viên khảo sát ý kiến khách hàng.
Nhiệm vụ của bạn:
- Đặt câu hỏi khảo sát một cách tự nhiên
- Ghi nhận ý kiến của khách hàng
- Cảm ơn khách hàng vì đã dành thời gian
- Giữ cuộc trò chuyện ngắn gọn (1-2 câu mỗi lượt)

Hãy tôn trọng thời gian của khách hàng."""
        }
    }
    
    @classmethod
    def get_scenario_prompt(cls, scenario_name: str) -> str:
        """Lấy prompt của scenario"""
        scenario = cls.DEFAULT_SCENARIOS.get(scenario_name)
        if scenario:
            return scenario["prompt"]
        return cls.DEFAULT_SCENARIOS["customer_support"]["prompt"]  # Default
    
    @classmethod
    def list_scenarios(cls) -> Dict[str, str]:
        """Liệt kê các scenario có sẵn"""
        return {name: data["name"] for name, data in cls.DEFAULT_SCENARIOS.items()}

# Global AI manager instance
ai_manager = None

def get_ai_manager() -> AIManager:
    """Get hoặc tạo AI manager instance"""
    global ai_manager
    if ai_manager is None:
        ai_manager = AIManager()
    return ai_manager
