import os
from abc import ABC, abstractmethod
from typing import Optional
import base64
import tempfile
from dotenv import load_dotenv

load_dotenv()

class TTSService(ABC):
    """Abstract base class cho Text-to-Speech service"""
    
    @abstractmethod
    async def synthesize(self, text: str, language: str = "vi-VN") -> bytes:
        """Chuyển đổi text thành audio bytes"""
        pass

class GoogleTTS(TTSService):
    """Google Cloud Text-to-Speech service"""
    
    def __init__(self):
        try:
            from google.cloud import texttospeech
            self.client = texttospeech.TextToSpeechClient()
            print("Google TTS initialized successfully")
        except Exception as e:
            print(f"Failed to initialize Google TTS: {e}")
            self.client = None
    
    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        """Chuyển text sang speech bằng Google Cloud TTS"""
        if not self.client:
            return None
            
        try:
            from google.cloud import texttospeech
            
            # Cấu hình voice
            voice_map = {
                "vi-VN": ("vi-VN-Standard-A", texttospeech.SsmlVoiceGender.FEMALE),
                "en-US": ("en-US-Standard-C", texttospeech.SsmlVoiceGender.FEMALE),
            }
            
            voice_name, gender = voice_map.get(language, voice_map["vi-VN"])
            
            # Tạo synthesis input
            synthesis_input = texttospeech.SynthesisInput(text=text)
            
            # Cấu hình voice
            voice = texttospeech.VoiceSelectionParams(
                language_code=language,
                name=voice_name,
                ssml_gender=gender
            )
            
            # Cấu hình audio
            audio_config = texttospeech.AudioConfig(
                audio_encoding=texttospeech.AudioEncoding.MP3,
                speaking_rate=1.0,
                pitch=0.0
            )
            
            # Thực hiện synthesis
            response = self.client.synthesize_speech(
                input=synthesis_input,
                voice=voice,
                audio_config=audio_config
            )
            
            return response.audio_content
            
        except Exception as e:
            print(f"Google TTS error: {e}")
            return None

class ElevenLabsTTS(TTSService):
    """ElevenLabs Text-to-Speech service"""
    
    def __init__(self):
        self.api_key = os.getenv("ELEVENLABS_API_KEY")
        self.voice_id = os.getenv("ELEVENLABS_VOICE_ID", "21m00Tcm4TlvDq8ikWAM")  # Default voice
        
        if self.api_key:
            print("ElevenLabs TTS initialized successfully")
        else:
            print("ElevenLabs API key not found")
    
    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        """Chuyển text sang speech bằng ElevenLabs"""
        if not self.api_key:
            return None
            
        try:
            import httpx
            
            url = f"https://api.elevenlabs.io/v1/text-to-speech/{self.voice_id}"
            
            headers = {
                "Accept": "audio/mpeg",
                "Content-Type": "application/json",
                "xi-api-key": self.api_key
            }
            
            data = {
                "text": text,
                "model_id": "eleven_multilingual_v2",
                "voice_settings": {
                    "stability": 0.5,
                    "similarity_boost": 0.5
                }
            }
            
            async with httpx.AsyncClient() as client:
                response = await client.post(url, json=data, headers=headers, timeout=30.0)
                
                if response.status_code == 200:
                    return response.content
                else:
                    print(f"ElevenLabs TTS error: {response.status_code} - {response.text}")
                    return None
                    
        except Exception as e:
            print(f"ElevenLabs TTS error: {e}")
            return None

class SimpleTTS(TTSService):
    """Simple TTS fallback using gTTS"""
    
    def __init__(self):
        try:
            from gtts import gTTS
            self.gtts = gTTS
            print("Simple TTS (gTTS) initialized successfully")
        except ImportError:
            print("gTTS not installed. Install with: pip install gtts")
            self.gtts = None
    
    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        """Chuyển text sang speech bằng gTTS"""
        if not self.gtts:
            return None
            
        try:
            # Map language code
            lang_code = language.split("-")[0]  # vi-VN -> vi
            
            # Tạo TTS object
            tts = self.gtts(text=text, lang=lang_code, slow=False)
            
            # Lưu vào temporary file
            with tempfile.NamedTemporaryFile(delete=False, suffix=".mp3") as fp:
                temp_file = fp.name
                tts.save(temp_file)
                
                # Đọc audio content
                with open(temp_file, "rb") as audio_file:
                    audio_content = audio_file.read()
                
                # Xóa file tạm
                os.unlink(temp_file)
                
                return audio_content
                
        except Exception as e:
            print(f"Simple TTS error: {e}")
            return None

class TTSManager:
    """Manager để quản lý nhiều TTS service"""
    
    def __init__(self):
        # Khởi tạo các TTS services theo thứ tự ưu tiên
        self.services = []
        
        # Thử Google TTS trước
        google_tts = GoogleTTS()
        if google_tts.client:
            self.services.append(google_tts)
        
        # Sau đó ElevenLabs
        elevenlabs_tts = ElevenLabsTTS()
        if elevenlabs_tts.api_key:
            self.services.append(elevenlabs_tts)
        
        # Cuối cùng là Simple TTS
        simple_tts = SimpleTTS()
        if simple_tts.gtts:
            self.services.append(simple_tts)
        
        if not self.services:
            raise Exception("No TTS service available! Please configure at least one TTS service.")
        
        print(f"TTS Manager initialized with {len(self.services)} service(s)")
    
    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        """Thử synthesize với các service theo thứ tự ưu tiên"""
        for service in self.services:
            try:
                audio_content = await service.synthesize(text, language)
                if audio_content:
                    return audio_content
            except Exception as e:
                print(f"TTS service failed: {e}, trying next service...")
                continue
        
        print("All TTS services failed!")
        return None
    
    def save_audio(self, audio_content: bytes, filename: str) -> str:
        """Lưu audio content ra file"""
        try:
            with open(filename, "wb") as f:
                f.write(audio_content)
            return filename
        except Exception as e:
            print(f"Error saving audio: {e}")
            return None

# Global TTS manager instance
tts_manager = None

def get_tts_manager() -> TTSManager:
    """Get hoặc tạo TTS manager instance"""
    global tts_manager
    if tts_manager is None:
        tts_manager = TTSManager()
    return tts_manager
