from __future__ import annotations

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    app_name: str = "call-center-ai"
    log_level: str = "INFO"
    server_host: str = "0.0.0.0"
    server_port: int = 8000

    database_url: str = ""
    db_host: str = "localhost"
    db_port: int = 3306
    db_user: str = "root"
    db_password: str = ""
    db_name: str = "call_center_db"

    anthropic_api_key: str = ""
    anthropic_model: str = "claude-opus-5-5"
    anthropic_effort: str = "low"
    anthropic_max_tokens: int = 2048
    anthropic_refusal_fallback: bool = True
    openai_api_key: str = ""
    openai_model: str = "gpt-4-turbo-preview"
    llm_max_tokens: int = 500
    llm_temperature: float = 0.7
    llm_timeout_seconds: float = 10.0

    google_tts_enabled: bool = True
    elevenlabs_api_key: str = ""
    elevenlabs_voice_id: str = "21m00Tcm4TlvDq8ikWAM"

    voice_language: str = "vi-VN"
    voice_gather_timeout: int = 5
    voice_scenario: str = "customer_support"
    voice_max_turns: int = 20
    knowledge_top_k: int = 3

    agent_phone_number: str = ""
    company_name: str = "công ty"
    company_hotline_hours: str = "8:00 - 21:00 tất cả các ngày trong tuần"
    company_website: str = ""
    company_info: str = ""

    @property
    def sqlalchemy_url(self) -> str:
        if self.database_url:
            return self.database_url
        return (
            f"mysql+pymysql://{self.db_user}:{self.db_password}"
            f"@{self.db_host}:{self.db_port}/{self.db_name}?charset=utf8mb4"
        )


settings = Settings()
