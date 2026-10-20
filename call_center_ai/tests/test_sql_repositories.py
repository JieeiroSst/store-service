import os
import tempfile
import unittest
from datetime import datetime, timedelta

from sqlalchemy import inspect, text

from app.adapters.outbound.persistence.db import build_engine, build_session_factory, create_schema
from app.adapters.outbound.persistence.repositories import (
    SqlCallRepository,
    SqlCustomerRepository,
    SqlKnowledgeRepository,
    SqlMessageRepository,
    SqlScenarioRepository,
)
from app.domain.models import Call, CallOutcome, CallStatus, Customer, KnowledgeEntry, Message, MessageRole, Scenario


class SqlRepositoryTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.engine = build_engine("sqlite:///" + os.path.join(self.tmp.name, "test.db"))
        create_schema(self.engine)
        sessions = build_session_factory(self.engine)
        self.calls = SqlCallRepository(sessions)
        self.messages = SqlMessageRepository(sessions)
        self.scenarios = SqlScenarioRepository(sessions)
        self.customers = SqlCustomerRepository(sessions)
        self.knowledge = SqlKnowledgeRepository(sessions)

    def tearDown(self):
        self.engine.dispose()
        self.tmp.cleanup()

    def test_call_lifecycle(self):
        call = self.calls.add(Call(call_sid="CA1", from_number="+84901", to_number="+84999"))
        self.assertIsNotNone(call.id)
        self.assertEqual(call.status, CallStatus.INITIATED)
        self.assertIsNotNone(call.start_time)
        self.messages.add(Message(call_id=call.id, role=MessageRole.ASSISTANT, content="hi"))
        self.messages.add(Message(call_id=call.id, role=MessageRole.USER, content="yo", processing_time=0.5))
        self.assertEqual([m.content for m in self.messages.list_by_call(call.id)], ["hi", "yo"])

        call.apply_provider_status(CallStatus.COMPLETED, 30, datetime.utcnow())
        call.scenario = "sales"
        saved = self.calls.save(call)
        self.assertEqual((saved.status, saved.duration, saved.outcome), (CallStatus.COMPLETED, 30, CallOutcome.ABANDONED))
        self.assertEqual(saved.scenario, "sales")
        self.assertEqual([c.call_sid for c in self.calls.list(0, 10, outcome=CallOutcome.ABANDONED)], ["CA1"])
        self.assertEqual(self.calls.get_by_sid("CA1").status, CallStatus.COMPLETED)

        full = self.calls.get_with_messages(call.id)
        self.assertEqual([m.role for m in full.messages], [MessageRole.ASSISTANT, MessageRole.USER])

        self.calls.add(Call(call_sid="CA2", from_number="+84902", to_number="+84999", status=CallStatus.BUSY))
        self.assertEqual([c.call_sid for c in self.calls.list(0, 10, status=CallStatus.BUSY)], ["CA2"])
        self.assertEqual([c.call_sid for c in self.calls.list(0, 10, from_number="+84901")], ["CA1"])

        stats = self.calls.statistics(datetime.utcnow() - timedelta(days=1))
        self.assertEqual(
            (stats.total_calls, stats.completed_calls, stats.failed_calls, stats.average_duration, stats.total_messages),
            (2, 1, 1, 30.0, 2),
        )
        self.assertEqual((stats.abandoned_calls, stats.resolved_calls), (1, 0))

        self.assertTrue(self.calls.delete(call.id))
        self.assertFalse(self.calls.delete(call.id))
        self.assertEqual(self.messages.list_by_call(call.id), [])

    def test_scenarios_and_customers(self):
        s = self.scenarios.add(Scenario(name="vip", prompt="p"))
        self.assertEqual(s.is_active, 1)
        s.is_active = 0
        self.scenarios.save(s)
        self.assertEqual(self.scenarios.list(0, 10), [])
        self.assertEqual(len(self.scenarios.list(0, 10, active_only=False)), 1)
        self.assertEqual(self.scenarios.get_by_name("vip").id, s.id)
        self.assertTrue(self.scenarios.delete(s.id))
        self.assertIsNone(self.scenarios.get(s.id))

        routed = self.scenarios.add(Scenario(name="sales", prompt="p", phone_number="+84888", is_default=1))
        self.assertEqual(self.scenarios.get_by_phone("+84888").id, routed.id)
        self.assertEqual(self.scenarios.get_default().name, "sales")

        c = self.customers.add(Customer(phone_number="+84901"))
        self.assertEqual(c.total_calls, 0)
        c.record_call(datetime.utcnow())
        self.customers.save(c)
        self.assertEqual(self.customers.get_by_phone("+84901").total_calls, 1)

    def test_knowledge(self):
        a = self.knowledge.add(KnowledgeEntry(question="q1", answer="a1", category="giao hàng"))
        self.knowledge.add(KnowledgeEntry(question="q2", answer="a2", scenario="sales"))
        self.knowledge.add(KnowledgeEntry(question="q3", answer="a3", scenario="survey"))
        self.knowledge.add(KnowledgeEntry(question="q4", answer="a4", is_active=0))
        self.assertEqual(sorted(e.question for e in self.knowledge.list_for_scenario("sales")), ["q1", "q2"])
        self.assertEqual([e.question for e in self.knowledge.list(0, 10, category="giao hàng")], ["q1"])
        self.assertEqual(self.knowledge.find_by_question("q1").id, a.id)
        a.answer = "new"
        self.assertEqual(self.knowledge.save(a).answer, "new")
        self.assertTrue(self.knowledge.delete(a.id))
        self.assertIsNone(self.knowledge.get(a.id))


class SchemaMigrationTest(unittest.TestCase):
    def test_adds_new_columns_to_existing_tables(self):
        with tempfile.TemporaryDirectory() as tmp:
            engine = build_engine("sqlite:///" + os.path.join(tmp, "old.db"))
            with engine.begin() as conn:
                conn.execute(text(
                    "CREATE TABLE calls (id INTEGER PRIMARY KEY, call_sid VARCHAR(255) NOT NULL, "
                    "from_number VARCHAR(20) NOT NULL, to_number VARCHAR(20) NOT NULL, status VARCHAR(11), "
                    "duration INTEGER, start_time DATETIME, end_time DATETIME, recording_url VARCHAR(512), "
                    "notes TEXT, created_at DATETIME, updated_at DATETIME)"
                ))
                conn.execute(text(
                    "CREATE TABLE scenarios (id INTEGER PRIMARY KEY, name VARCHAR(255) NOT NULL, description TEXT, "
                    "prompt TEXT NOT NULL, is_active INTEGER, created_at DATETIME, updated_at DATETIME)"
                ))
                conn.execute(text(
                    "INSERT INTO scenarios (name, prompt, is_active) VALUES ('legacy', 'p', 1)"
                ))
            create_schema(engine)
            create_schema(engine)
            columns = {c["name"] for c in inspect(engine).get_columns("calls")}
            self.assertTrue({"scenario", "outcome"} <= columns)
            repo = SqlScenarioRepository(build_session_factory(engine))
            legacy = repo.get_by_name("legacy")
            self.assertEqual((legacy.is_default, legacy.greeting), (0, None))
            engine.dispose()


if __name__ == "__main__":
    unittest.main()
