import asyncio
import unittest

from app.adapters.outbound.llm.fallback import FallbackLLM
from app.domain import conversation, playbook
from app.domain.errors import AlreadyExistsError, NotFoundError
from app.domain.models import CallOutcome, CallStatus, KnowledgeEntry, MessageRole, Scenario
from tests.fakes import FailingLLM, FakeLLM, ScriptedLLM, fake_container

GREETING = playbook.DEFAULT_SCENARIOS["customer_support"]["greeting"]


def run(coro):
    return asyncio.run(coro)


class VoiceCallServiceTest(unittest.TestCase):
    def setUp(self):
        self.llm = FakeLLM()
        self.c = fake_container(llm=self.llm)
        self.voice = self.c.voice_service()

    def test_start_call_registers_customer_call_and_greeting(self):
        reply = self.voice.start_call("CA1", "+84901", "+84999")
        self.assertEqual(reply.text, GREETING)
        self.assertTrue(reply.listen)
        self.assertEqual(reply.closing, conversation.NO_INPUT)
        call = self.c.calls.get_by_sid("CA1")
        self.assertEqual((call.status, call.notes, call.scenario), (CallStatus.INITIATED, "Incoming call", "customer_support"))
        msgs = self.c.messages.list_by_call(call.id)
        self.assertEqual([(m.role, m.content) for m in msgs], [(MessageRole.ASSISTANT, GREETING)])
        self.voice.start_call("CA2", "+84901", "+84999")
        self.assertEqual(len(self.c.customers.store.rows), 1)
        customer = self.c.customers.get_by_phone("+84901")
        self.assertEqual(customer.total_calls, 2)
        self.assertIsNotNone(customer.last_call_date)

    def test_speech_turn_sends_history_with_scenario_and_knowledge(self):
        self.c.knowledge.add(KnowledgeEntry(question="Phí giao hàng là bao nhiêu?", answer="Miễn phí từ 300 nghìn.", keywords="phí ship"))
        self.c.knowledge.add(KnowledgeEntry(question="Quên mật khẩu?", answer="Bấm quên mật khẩu."))
        self.voice.start_call("CA1", "+84901", "+84999")
        reply = run(self.voice.handle_speech("CA1", "Cho tôi hỏi phi ship bao nhieu vay"))
        self.assertEqual((reply.text, reply.listen, reply.closing), (self.llm.answer, True, conversation.TIMEOUT_FAREWELL))
        self.assertEqual(
            [(t.role, t.content) for t in self.llm.calls[0]],
            [("assistant", GREETING), ("user", "Cho tôi hỏi phi ship bao nhieu vay")],
        )
        prompt = self.llm.prompts[0]
        self.assertIn("nhân viên tổng đài chăm sóc khách hàng", prompt)
        self.assertIn("Miễn phí từ 300 nghìn.", prompt)
        self.assertNotIn("Bấm quên mật khẩu.", prompt)
        self.assertIn("[[TRANSFER]]", prompt)
        call = self.c.calls.get_by_sid("CA1")
        self.assertEqual(call.status, CallStatus.IN_PROGRESS)
        self.assertIsNotNone(self.c.messages.list_by_call(call.id)[-1].processing_time)

    def test_keyword_end_completes_call_as_resolved(self):
        self.voice.start_call("CA1", "+84901", "+84999")
        reply = run(self.voice.handle_speech("CA1", "Vậy thôi, tạm biệt nhé"))
        self.assertFalse(reply.listen)
        self.assertEqual(reply.closing, conversation.FAREWELL)
        call = self.c.calls.get_by_sid("CA1")
        self.assertEqual((call.status, call.outcome), (CallStatus.COMPLETED, CallOutcome.RESOLVED))
        self.assertIsNotNone(call.end_time)

    def test_thanks_mid_sentence_does_not_hang_up(self):
        self.voice.start_call("CA1", "+84901", "+84999")
        reply = run(self.voice.handle_speech("CA1", "Cảm ơn, cho tôi hỏi thêm về đổi trả"))
        self.assertTrue(reply.listen)

    def test_llm_end_directive_is_stripped_and_ends_call(self):
        c = fake_container(llm=ScriptedLLM("Cảm ơn bạn đã gọi, chúc bạn một ngày tốt lành! [[END]]"))
        voice = c.voice_service()
        voice.start_call("CA1", "+84901", "+84999")
        reply = run(voice.handle_speech("CA1", "Không cần gì nữa đâu cảm ơn em"))
        self.assertEqual(reply.text, "Cảm ơn bạn đã gọi, chúc bạn một ngày tốt lành!")
        self.assertEqual((reply.listen, reply.closing), (False, None))
        self.assertEqual(c.calls.get_by_sid("CA1").outcome, CallOutcome.RESOLVED)

    def test_llm_transfer_directive_dials_agent(self):
        c = fake_container(llm=ScriptedLLM("Rất xin lỗi bạn về sự cố này. [[TRANSFER]]"), agent_phone_number="+8428000")
        voice = c.voice_service()
        voice.start_call("CA1", "+84901", "+84999")
        reply = run(voice.handle_speech("CA1", "Tôi bị trừ tiền 2 lần"))
        self.assertEqual(reply.transfer_to, "+8428000")
        self.assertTrue(reply.text.startswith("Rất xin lỗi bạn về sự cố này. "))
        self.assertIn(conversation.TRANSFERRING, reply.text)
        call = c.calls.get_by_sid("CA1")
        self.assertEqual(call.outcome, CallOutcome.TRANSFERRED)
        self.assertNotIn("[[", c.messages.list_by_call(call.id)[-1].content)

    def test_asking_for_human_skips_llm_and_queues_callback_without_agent(self):
        self.voice.start_call("CA1", "+84901", "+84999")
        reply = run(self.voice.handle_speech("CA1", "Cho tôi gặp nhân viên"))
        self.assertEqual(self.llm.calls, [])
        self.assertEqual((reply.text, reply.listen, reply.transfer_to), (conversation.CALLBACK_PROMISE, False, None))
        call = self.c.calls.get_by_sid("CA1")
        self.assertEqual((call.status, call.outcome), (CallStatus.COMPLETED, CallOutcome.CALLBACK_REQUESTED))
        self.assertIn("gọi lại", call.notes)

    def test_too_many_turns_hands_off(self):
        c = fake_container(voice_max_turns=1)
        voice = c.voice_service()
        voice.start_call("CA1", "+84901", "+84999")
        self.assertTrue(run(voice.handle_speech("CA1", "câu một")).listen)
        reply = run(voice.handle_speech("CA1", "câu hai"))
        self.assertEqual(c.calls.get_by_sid("CA1").outcome, CallOutcome.CALLBACK_REQUESTED)
        self.assertFalse(reply.listen)

    def test_scenario_routing_by_dialed_number_and_default(self):
        self.c.scenarios.add(Scenario(name="sales", prompt="SALES PROMPT", greeting="Chào mừng tới phòng kinh doanh", phone_number="+84888"))
        self.c.scenarios.add(Scenario(name="vip", prompt="VIP PROMPT", greeting="Chào khách VIP", is_default=1))
        self.assertEqual(self.voice.start_call("CA1", "+84901", "+84888").text, "Chào mừng tới phòng kinh doanh")
        self.assertEqual(self.voice.start_call("CA2", "+84901", "+84111").text, "Chào khách VIP")
        run(self.voice.handle_speech("CA1", "alo"))
        self.assertTrue(self.llm.prompts[0].startswith("SALES PROMPT"))
        self.assertEqual(self.voice.start_call("CA3", "+84901", "+84111", scenario_name="survey").text,
                         playbook.DEFAULT_SCENARIOS["survey"]["greeting"])

    def test_customer_profile_is_used_in_prompt(self):
        self.voice.start_call("CA1", "+84901", "+84999")
        self.c.call_service().update_customer("+84901", {"name": "Chị Lan", "notes": "Khách VIP"})
        run(self.voice.handle_speech("CA1", "alo"))
        self.assertIn("Chị Lan", self.llm.prompts[0])
        self.assertIn("Khách VIP", self.llm.prompts[0])

    def test_empty_speech_and_unknown_call(self):
        reply = run(self.voice.handle_speech("CA1", ""))
        self.assertEqual((reply.text, reply.listen, reply.closing), (conversation.NOT_HEARD, True, None))
        reply = run(self.voice.handle_speech("missing", "alo"))
        self.assertEqual((reply.text, reply.listen), (conversation.SYSTEM_ERROR, False))

    def test_llm_failure_falls_back_to_apology(self):
        c = fake_container(llm=FallbackLLM([FailingLLM()]))
        voice = c.voice_service()
        voice.start_call("CA1", "+84901", "+84999")
        self.assertEqual(run(voice.handle_speech("CA1", "alo")).text, conversation.AI_UNAVAILABLE)

    def test_fallback_llm_uses_next_provider(self):
        llm = FallbackLLM([FailingLLM(), FakeLLM(None), FakeLLM("ok")])
        self.assertEqual(run(llm.generate([])), "ok")

    def test_provider_status_updates(self):
        self.voice.start_call("CA1", "+84901", "+84999")
        self.voice.update_status("CA1", CallStatus.IN_PROGRESS, 0)
        call = self.c.calls.get_by_sid("CA1")
        self.assertEqual((call.status, call.end_time), (CallStatus.IN_PROGRESS, None))
        self.voice.update_status("CA1", CallStatus.COMPLETED, 33)
        call = self.c.calls.get_by_sid("CA1")
        self.assertEqual((call.status, call.duration, call.outcome), (CallStatus.COMPLETED, 33, CallOutcome.ABANDONED))
        self.voice.update_status("CA1", CallStatus.IN_PROGRESS, 0)
        self.assertEqual(self.c.calls.get_by_sid("CA1").status, CallStatus.COMPLETED)
        self.voice.update_status("missing", CallStatus.BUSY, 1)

    def test_resolved_call_keeps_outcome_after_status_callback(self):
        self.voice.start_call("CA1", "+84901", "+84999")
        run(self.voice.handle_speech("CA1", "tạm biệt"))
        self.voice.update_status("CA1", CallStatus.COMPLETED, 20)
        call = self.c.calls.get_by_sid("CA1")
        self.assertEqual((call.outcome, call.duration), (CallOutcome.RESOLVED, 20))


class ScenarioServiceTest(unittest.TestCase):
    def setUp(self):
        self.c = fake_container()
        self.service = self.c.scenario_service()

    def test_crud(self):
        s = self.service.create_scenario(Scenario(name="vip", prompt="p"))
        with self.assertRaises(AlreadyExistsError):
            self.service.create_scenario(Scenario(name="vip", prompt="p2"))
        updated = self.service.update_scenario(s.id, {"prompt": "new", "id": 999})
        self.assertEqual((updated.id, updated.prompt, updated.name), (s.id, "new", "vip"))
        self.service.update_scenario(s.id, {"is_active": 0})
        self.assertEqual(self.service.list_scenarios(), [])
        self.service.delete_scenario(s.id)
        with self.assertRaises(NotFoundError):
            self.service.get_scenario(s.id)
        with self.assertRaises(NotFoundError):
            self.service.delete_scenario(s.id)

    def test_only_one_default(self):
        a = self.service.create_scenario(Scenario(name="a", prompt="p", is_default=1))
        b = self.service.create_scenario(Scenario(name="b", prompt="p", is_default=1))
        self.assertEqual(self.service.get_scenario(a.id).is_default, 0)
        self.service.update_scenario(a.id, {"is_default": 1})
        self.assertEqual(self.service.get_scenario(b.id).is_default, 0)

    def test_seed_is_idempotent(self):
        self.assertEqual(set(self.service.seed_defaults()), set(playbook.DEFAULT_SCENARIOS))
        self.assertEqual(self.service.seed_defaults(), [])
        self.assertEqual(self.c.scenarios.get_default().name, "customer_support")


class KnowledgeServiceTest(unittest.TestCase):
    def setUp(self):
        self.service = fake_container().knowledge_service()
        self.service.seed_defaults()

    def test_seed_is_idempotent(self):
        self.assertEqual(self.service.seed_defaults(), [])

    def test_search_answers_common_questions(self):
        cases = {
            "phí ship bao nhiêu": "Phí giao hàng là bao nhiêu?",
            "bao lau thi toi nhan duoc hang": "Bao lâu thì nhận được hàng?",
            "tôi muốn trả hàng vì không vừa": "Chính sách đổi trả hàng như thế nào?",
            "khi nào tôi được hoàn tiền": "Bao lâu thì được hoàn tiền?",
            "có trả góp không": "Có những phương thức thanh toán nào?",
            "tôi quên mật khẩu": "Tôi quên mật khẩu tài khoản thì làm sao?",
            "xuất hóa đơn vat": "Tôi muốn xuất hóa đơn VAT?",
            "tổng đài làm việc mấy giờ": "Giờ làm việc của tổng đài và cửa hàng là khi nào?",
            "đơn hàng của tôi đang ở đâu": "Làm sao để kiểm tra tình trạng đơn hàng?",
            "tôi bị trừ tiền rồi mà đơn vẫn chưa thanh toán": "Tôi đã bị trừ tiền nhưng đơn hàng báo chưa thanh toán thì sao?",
        }
        for query, expected in cases.items():
            with self.subTest(query=query):
                found = self.service.search(query)
                self.assertTrue(found, query)
                self.assertEqual(found[0].question, expected)

    def test_scenario_specific_entries(self):
        general = self.service.search("đặt lịch hẹn khung giờ nào", scenario="customer_support")
        self.assertFalse([e for e in general if e.scenario == "appointment"])
        found = self.service.search("đặt lịch hẹn khung giờ nào", scenario="appointment")
        self.assertEqual(found[0].scenario, "appointment")

    def test_irrelevant_query_returns_nothing(self):
        self.assertEqual(self.service.search("alo alo"), [])


class ConversationRulesTest(unittest.TestCase):
    def test_directives(self):
        self.assertEqual(conversation.parse_directive("Dạ vâng."), ("Dạ vâng.", conversation.Directive.NONE))
        self.assertEqual(conversation.parse_directive("Tạm biệt [[END]]"), ("Tạm biệt", conversation.Directive.END))
        self.assertEqual(conversation.parse_directive("[[ transfer ]]"), ("", conversation.Directive.TRANSFER))

    def test_intents(self):
        self.assertTrue(conversation.wants_human("cho toi gap nhan vien that"))
        self.assertTrue(conversation.wants_human("Chuyển máy cho tổng đài viên giúp"))
        self.assertFalse(conversation.wants_human("nhân viên giao hàng chưa tới"))
        self.assertTrue(conversation.wants_to_end("Ok tạm biệt"))
        self.assertFalse(conversation.wants_to_end("cảm ơn"))


if __name__ == "__main__":
    unittest.main()
