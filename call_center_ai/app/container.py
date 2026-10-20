from __future__ import annotations

import logging
from typing import List, Optional

from sqlalchemy.engine import Engine

from app.adapters.outbound.llm.fallback import FallbackLLM
from app.adapters.outbound.persistence.db import build_engine, build_session_factory, create_schema
from app.adapters.outbound.persistence.repositories import (
    SqlCallRepository,
    SqlCustomerRepository,
    SqlKnowledgeRepository,
    SqlMessageRepository,
    SqlScenarioRepository,
)
from app.adapters.outbound.tts.fallback import FallbackTTS
from app.application.calls import CallService
from app.application.knowledge import KnowledgeService
from app.application.scenarios import ScenarioResolver, ScenarioService
from app.application.speech import SpeechService
from app.application.voice import VoicePolicy, VoiceCallService
from app.config import Settings
from app.domain.conversation import CompanyProfile
from app.domain.ports import (
    CallRepositoryPort,
    CustomerRepositoryPort,
    KnowledgeRepositoryPort,
    LLMPort,
    MessageRepositoryPort,
    ScenarioRepositoryPort,
    TTSPort,
)

log = logging.getLogger(__name__)


class Container:
    def __init__(
        self,
        settings: Settings,
        calls: CallRepositoryPort,
        messages: MessageRepositoryPort,
        customers: CustomerRepositoryPort,
        scenarios: ScenarioRepositoryPort,
        knowledge: KnowledgeRepositoryPort,
        llm: LLMPort,
        tts: TTSPort,
        engine: Optional[Engine] = None,
    ) -> None:
        self.settings = settings
        self.calls = calls
        self.messages = messages
        self.customers = customers
        self.scenarios = scenarios
        self.knowledge = knowledge
        self.llm = llm
        self.tts = tts
        self._engine = engine

    @classmethod
    def from_settings(cls, settings: Settings) -> "Container":
        engine = build_engine(settings.sqlalchemy_url)
        sessions = build_session_factory(engine)
        return cls(
            settings=settings,
            calls=SqlCallRepository(sessions),
            messages=SqlMessageRepository(sessions),
            customers=SqlCustomerRepository(sessions),
            scenarios=SqlScenarioRepository(sessions),
            knowledge=SqlKnowledgeRepository(sessions),
            llm=FallbackLLM(build_llm_providers(settings)),
            tts=FallbackTTS(build_tts_providers(settings)),
            engine=engine,
        )

    def startup(self) -> None:
        if self._engine is not None:
            create_schema(self._engine)

    def shutdown(self) -> None:
        if self._engine is not None:
            self._engine.dispose()

    def voice_policy(self) -> VoicePolicy:
        s = self.settings
        return VoicePolicy(
            company=CompanyProfile(
                name=s.company_name,
                hotline_hours=s.company_hotline_hours,
                website=s.company_website,
                extra=s.company_info,
            ),
            agent_number=s.agent_phone_number or None,
            max_turns=s.voice_max_turns,
            knowledge_limit=s.knowledge_top_k,
        )

    def scenario_resolver(self) -> ScenarioResolver:
        return ScenarioResolver(self.scenarios, self.settings.voice_scenario)

    def voice_service(self) -> VoiceCallService:
        return VoiceCallService(
            self.calls,
            self.messages,
            self.customers,
            self.knowledge,
            self.scenario_resolver(),
            self.llm,
            policy=self.voice_policy(),
        )

    def call_service(self) -> CallService:
        return CallService(self.calls, self.customers)

    def scenario_service(self) -> ScenarioService:
        return ScenarioService(self.scenarios)

    def knowledge_service(self) -> KnowledgeService:
        return KnowledgeService(self.knowledge, self.settings.voice_scenario)

    def speech_service(self) -> SpeechService:
        return SpeechService(self.tts)


def build_llm_providers(settings: Settings) -> List[LLMPort]:
    providers: List[LLMPort] = []
    if settings.anthropic_api_key:
        try:
            from app.adapters.outbound.llm.anthropic_llm import AnthropicLLM

            providers.append(
                AnthropicLLM(
                    settings.anthropic_api_key,
                    settings.anthropic_model,
                    max_tokens=settings.anthropic_max_tokens,
                    effort=settings.anthropic_effort,
                    timeout=settings.llm_timeout_seconds,
                    refusal_fallback=settings.anthropic_refusal_fallback,
                )
            )
        except Exception as e:
            log.warning("Anthropic unavailable: %s", e)
    if settings.openai_api_key:
        try:
            from app.adapters.outbound.llm.openai_llm import OpenAILLM

            providers.append(
                OpenAILLM(
                    settings.openai_api_key,
                    settings.openai_model,
                    max_tokens=settings.llm_max_tokens,
                    temperature=settings.llm_temperature,
                    timeout=settings.llm_timeout_seconds,
                )
            )
        except Exception as e:
            log.warning("OpenAI unavailable: %s", e)
    if not providers:
        log.warning("no LLM provider configured, callers will hear the fallback answer")
    return providers


def build_tts_providers(settings: Settings) -> List[TTSPort]:
    providers: List[TTSPort] = []
    if settings.google_tts_enabled:
        try:
            from app.adapters.outbound.tts.google_tts import GoogleTTS

            providers.append(GoogleTTS())
        except Exception as e:
            log.warning("Google TTS unavailable: %s", e)
    if settings.elevenlabs_api_key:
        from app.adapters.outbound.tts.elevenlabs_tts import ElevenLabsTTS

        providers.append(ElevenLabsTTS(settings.elevenlabs_api_key, settings.elevenlabs_voice_id))
    try:
        from app.adapters.outbound.tts.gtts_tts import GTTS

        providers.append(GTTS())
    except Exception as e:
        log.warning("gTTS unavailable: %s", e)
    if not providers:
        log.warning("no TTS provider configured")
    return providers
