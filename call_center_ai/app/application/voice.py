from __future__ import annotations

import logging
import time
from dataclasses import dataclass, field
from datetime import datetime
from typing import Callable, Optional

from app.application.scenarios import ScenarioResolver
from app.domain import conversation, knowledge
from app.domain.conversation import CompanyProfile, Directive
from app.domain.models import Call, CallOutcome, CallStatus, Customer, Message, MessageRole, VoiceReply
from app.domain.ports import (
    CallRepositoryPort,
    CustomerRepositoryPort,
    KnowledgeRepositoryPort,
    LLMPort,
    MessageRepositoryPort,
)

log = logging.getLogger(__name__)


@dataclass(frozen=True)
class VoicePolicy:
    company: CompanyProfile = field(default_factory=CompanyProfile)
    agent_number: Optional[str] = None
    max_turns: int = 20
    knowledge_limit: int = 3


class VoiceCallService:
    def __init__(
        self,
        calls: CallRepositoryPort,
        messages: MessageRepositoryPort,
        customers: CustomerRepositoryPort,
        knowledge_base: KnowledgeRepositoryPort,
        resolver: ScenarioResolver,
        llm: LLMPort,
        policy: VoicePolicy = VoicePolicy(),
        clock: Callable[[], datetime] = datetime.utcnow,
    ) -> None:
        self._calls = calls
        self._messages = messages
        self._customers = customers
        self._knowledge = knowledge_base
        self._resolver = resolver
        self._llm = llm
        self._policy = policy
        self._clock = clock

    def start_call(
        self,
        call_sid: str,
        from_number: str,
        to_number: str,
        scenario_name: Optional[str] = None,
        notes: str = "Incoming call",
    ) -> VoiceReply:
        log.info("incoming call from %s to %s (CallSid: %s)", from_number, to_number, call_sid)
        customer = self._customers.get_by_phone(from_number) or self._customers.add(Customer(phone_number=from_number))
        customer.record_call(self._clock())
        self._customers.save(customer)

        if scenario_name:
            scenario = self._resolver.by_name(scenario_name)
        else:
            scenario = self._resolver.for_number(to_number)
        greeting = scenario.greeting or conversation.GREETING
        call = self._calls.add(
            Call(
                call_sid=call_sid,
                from_number=from_number,
                to_number=to_number,
                notes=notes,
                scenario=scenario.name,
            )
        )
        self._say(call, greeting)
        return VoiceReply(text=greeting, listen=True, closing=conversation.NO_INPUT)

    async def handle_speech(self, call_sid: str, speech: str) -> VoiceReply:
        if not speech:
            return VoiceReply(text=conversation.NOT_HEARD, listen=True)

        call = self._calls.get_by_sid(call_sid)
        if call is None:
            return VoiceReply(text=conversation.SYSTEM_ERROR, listen=False)

        call.mark_in_progress()
        self._calls.save(call)
        self._messages.add(Message(call_id=call.id, role=MessageRole.USER, content=speech))

        if conversation.wants_human(speech):
            return self._handoff(call)

        history = conversation.to_chat_history(self._messages.list_by_call(call.id))
        if conversation.user_turns(history) > self._policy.max_turns:
            return self._handoff(call)

        scenario = self._resolver.by_name(call.scenario)
        facts = knowledge.search(
            conversation.recent_user_text(history),
            self._knowledge.list_for_scenario(scenario.name),
            limit=self._policy.knowledge_limit,
        )
        system_prompt = conversation.build_system_prompt(
            scenario.prompt,
            self._policy.company,
            facts,
            self._customers.get_by_phone(call.from_number),
        )

        started = time.monotonic()
        answer = await self._llm.generate(history, system_prompt)
        processing_time = time.monotonic() - started

        text, directive = conversation.parse_directive(answer or "")
        if not answer:
            text, directive = conversation.AI_UNAVAILABLE, Directive.NONE

        if directive == Directive.TRANSFER:
            return self._handoff(call, lead=text, processing_time=processing_time)

        if directive == Directive.END:
            text = text or conversation.FAREWELL
        self._say(call, text, processing_time)

        if directive == Directive.END or conversation.wants_to_end(speech):
            call.complete(CallOutcome.RESOLVED, self._clock())
            self._calls.save(call)
            closing = None if directive == Directive.END else conversation.FAREWELL
            return VoiceReply(text=text, listen=False, closing=closing)

        return VoiceReply(text=text, listen=True, closing=conversation.TIMEOUT_FAREWELL)

    def update_status(self, call_sid: str, status: CallStatus, duration: int) -> None:
        log.info("call %s status: %s, duration: %ss", call_sid, status.value, duration)
        call = self._calls.get_by_sid(call_sid)
        if call is None:
            return
        call.apply_provider_status(status, duration, self._clock())
        self._calls.save(call)

    def _handoff(self, call: Call, lead: str = "", processing_time: Optional[float] = None) -> VoiceReply:
        agent = self._policy.agent_number
        if agent:
            text = f"{lead} {conversation.TRANSFERRING}".strip()
            call.transfer()
            self._calls.save(call)
            self._say(call, text, processing_time)
            log.info("call %s transferred to agent", call.call_sid)
            return VoiceReply(text=text, listen=False, transfer_to=agent)

        text = f"{lead} {conversation.CALLBACK_PROMISE}".strip()
        call.notes = f"{call.notes or ''}\nKhách yêu cầu nhân viên gọi lại".strip()
        call.complete(CallOutcome.CALLBACK_REQUESTED, self._clock())
        self._calls.save(call)
        self._say(call, text, processing_time)
        log.info("call %s queued for callback", call.call_sid)
        return VoiceReply(text=text, listen=False, closing=conversation.FAREWELL)

    def _say(self, call: Call, text: str, processing_time: Optional[float] = None) -> None:
        self._messages.add(
            Message(call_id=call.id, role=MessageRole.ASSISTANT, content=text, processing_time=processing_time)
        )
