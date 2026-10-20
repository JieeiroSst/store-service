import os
import unittest
import xml.etree.ElementTree as ET

from fastapi.testclient import TestClient

from app.adapters.inbound.api import auth_deps
from app.adapters.inbound.api.main import create_app
from app.domain import conversation, playbook
from tests.fakes import FakeTTS, ScriptedLLM, fake_container

ADMIN = {"X-Internal-Token": "s3cret"}


class ApiTest(unittest.TestCase):
    def setUp(self):
        os.environ.pop("TWILIO_AUTH_TOKEN", None)
        os.environ["INTERNAL_API_TOKENS"] = "s3cret"
        auth_deps._authenticator = None
        self.start()

    def start(self, **kwargs):
        self.container = fake_container(**kwargs)
        app = create_app()
        app.state.container = self.container
        self.client = TestClient(app)
        self.client.__enter__()

    def tearDown(self):
        self.client.__exit__(None, None, None)
        os.environ.pop("INTERNAL_API_TOKENS", None)
        auth_deps._authenticator = None

    def twiml(self, path, **form):
        r = self.client.post(path, data=form)
        self.assertEqual(r.status_code, 200)
        self.assertEqual(r.headers["content-type"], "application/xml")
        return ET.fromstring(r.text)

    def test_voice_flow(self):
        root = self.twiml("/voice/incoming", CallSid="CA1", From="+84901", To="+84999")
        self.assertEqual([el.tag for el in root], ["Gather", "Say", "Hangup"])
        gather = root.find("Gather")
        self.assertEqual(gather.get("action"), "/voice/process-speech")
        self.assertEqual(gather.get("language"), "vi-VN")
        self.assertEqual(gather.find("Say").text, playbook.DEFAULT_SCENARIOS["customer_support"]["greeting"])

        root = self.twiml("/voice/process-speech", CallSid="CA1", SpeechResult="xin chào")
        self.assertEqual([el.tag for el in root], ["Gather", "Say", "Hangup"])
        self.assertEqual(root[1].text, conversation.TIMEOUT_FAREWELL)

        root = self.twiml("/voice/process-speech", CallSid="CA1", SpeechResult="tạm biệt")
        self.assertEqual([el.tag for el in root], ["Say", "Say", "Hangup"])
        self.assertEqual(root[1].text, conversation.FAREWELL)

        root = self.twiml("/voice/process-speech", CallSid="CA1")
        self.assertEqual([el.tag for el in root], ["Gather"])

        r = self.client.post("/voice/status", data={"CallSid": "CA1", "CallStatus": "completed", "CallDuration": "42"})
        self.assertEqual(r.json(), {"status": "ok"})

        body = self.client.get("/api/calls/1", headers=ADMIN).json()
        self.assertEqual((body["status"], body["duration"], body["outcome"]), ("completed", 42, "resolved"))
        self.assertEqual(body["scenario"], "customer_support")
        self.assertEqual(len(body["messages"]), 5)

        stats = self.client.get("/api/analytics/statistics", headers=ADMIN).json()
        self.assertEqual((stats["total_calls"], stats["completed_calls"], stats["resolved_calls"]), (1, 1, 1))
        self.assertEqual(stats["total_messages"], 5)

        self.assertEqual(self.client.get("/api/customers/+84901", headers=ADMIN).json()["total_calls"], 1)
        self.assertEqual(len(self.client.get("/api/calls?outcome=resolved", headers=ADMIN).json()), 1)
        self.assertEqual(len(self.client.get("/api/calls?outcome=transferred", headers=ADMIN).json()), 0)
        self.assertEqual(self.client.delete("/api/calls/1", headers=ADMIN).status_code, 200)
        self.assertEqual(self.client.get("/api/calls/1", headers=ADMIN).status_code, 404)

    def test_status_callback_mapping(self):
        self.twiml("/voice/incoming", CallSid="CA1", From="+84901", To="+84999")
        r = self.client.post("/voice/status", data={"CallSid": "CA1", "CallStatus": "in-progress"})
        self.assertEqual(r.json(), {"status": "ok"})
        self.assertEqual(self.client.get("/api/calls/1", headers=ADMIN).json()["status"], "in_progress")
        r = self.client.post("/voice/status", data={"CallSid": "CA1", "CallStatus": "weird"})
        self.assertEqual(r.json(), {"status": "ignored"})
        self.assertEqual(self.client.get("/api/calls/1", headers=ADMIN).json()["status"], "in_progress")

    def test_transfer_renders_dial(self):
        self.client.__exit__(None, None, None)
        self.start(agent_phone_number="+842800000")
        self.twiml("/voice/incoming", CallSid="CA1", From="+84901", To="+84999")
        root = self.twiml("/voice/process-speech", CallSid="CA1", SpeechResult="cho tôi gặp nhân viên")
        self.assertEqual([el.tag for el in root], ["Say", "Dial"])
        self.assertEqual(root[1].text, "+842800000")
        self.assertEqual(self.client.get("/api/calls/1", headers=ADMIN).json()["outcome"], "transferred")

    def test_admin_requires_auth(self):
        self.assertEqual(self.client.get("/api/calls").status_code, 401)
        self.assertEqual(self.client.get("/api/knowledge").status_code, 401)
        self.assertEqual(self.client.get("/").status_code, 200)

    def test_scenarios(self):
        r = self.client.post(
            "/api/scenarios",
            json={"name": "vip", "prompt": "p", "greeting": "Chào VIP", "phone_number": "+84888"},
            headers=ADMIN,
        )
        self.assertEqual(r.status_code, 200)
        sid = r.json()["id"]
        self.assertEqual((r.json()["greeting"], r.json()["phone_number"]), ("Chào VIP", "+84888"))
        r = self.client.post("/api/scenarios", json={"name": "vip", "prompt": "p"}, headers=ADMIN)
        self.assertEqual(r.status_code, 400)
        r = self.client.put(f"/api/scenarios/{sid}", json={"description": "d"}, headers=ADMIN)
        self.assertEqual((r.json()["description"], r.json()["prompt"]), ("d", "p"))
        root = self.twiml("/voice/incoming", CallSid="CA9", From="+84901", To="+84888")
        self.assertEqual(root.find("Gather").find("Say").text, "Chào VIP")
        self.assertEqual(self.client.delete(f"/api/scenarios/{sid}", headers=ADMIN).status_code, 200)
        self.assertEqual(self.client.get(f"/api/scenarios/{sid}", headers=ADMIN).status_code, 404)

    def test_seed_and_knowledge_crud(self):
        r = self.client.post("/api/scenarios/seed", headers=ADMIN).json()
        self.assertEqual(set(r["scenarios"]), set(playbook.DEFAULT_SCENARIOS))
        self.assertEqual(r["knowledge_entries"], len(playbook.DEFAULT_KNOWLEDGE))

        found = self.client.get("/api/knowledge/search", params={"q": "phí ship bao nhiêu"}, headers=ADMIN).json()
        self.assertEqual(found[0]["question"], "Phí giao hàng là bao nhiêu?")

        r = self.client.post(
            "/api/knowledge",
            json={"question": "Có bán hàng sỉ không?", "answer": "Có, đơn từ 50 sản phẩm.", "keywords": "sỉ, buôn"},
            headers=ADMIN,
        )
        eid = r.json()["id"]
        found = self.client.get("/api/knowledge/search", params={"q": "mua sỉ"}, headers=ADMIN).json()
        self.assertEqual(found[0]["id"], eid)
        r = self.client.put(f"/api/knowledge/{eid}", json={"is_active": 0}, headers=ADMIN)
        self.assertEqual(r.json()["is_active"], 0)
        found = self.client.get("/api/knowledge/search", params={"q": "mua sỉ"}, headers=ADMIN).json()
        self.assertEqual(found, [])
        self.assertEqual(len(self.client.get("/api/knowledge?category=giao hàng", headers=ADMIN).json()), 2)
        self.assertEqual(self.client.delete(f"/api/knowledge/{eid}", headers=ADMIN).status_code, 200)
        self.assertEqual(self.client.get(f"/api/knowledge/{eid}", headers=ADMIN).status_code, 404)

    def test_simulate_conversation(self):
        self.client.__exit__(None, None, None)
        self.start(llm=ScriptedLLM("Đơn từ 300 nghìn được miễn phí giao hàng ạ.", "Cảm ơn bạn, chào bạn! [[END]]"))
        self.client.post("/api/scenarios/seed", headers=ADMIN)

        r = self.client.post("/api/simulate", json={"scenario": "sales"}, headers=ADMIN).json()
        self.assertEqual(r["greeting"], playbook.DEFAULT_SCENARIOS["sales"]["greeting"])
        self.assertTrue(r["listening"])
        sid = r["call_sid"]

        r = self.client.post("/api/simulate", json={"call_sid": sid, "message": "phí ship bao nhiêu"}, headers=ADMIN).json()
        self.assertEqual(r["reply"], "Đơn từ 300 nghìn được miễn phí giao hàng ạ.")
        self.assertIn("Phí giao hàng là bao nhiêu?", r["knowledge"])
        self.assertFalse(r["ended"])
        self.assertIn("tư vấn bán hàng", self.container.llm.prompts[0])

        r = self.client.post("/api/simulate", json={"call_sid": sid, "message": "ok cảm ơn"}, headers=ADMIN).json()
        self.assertEqual(r["reply"], "Cảm ơn bạn, chào bạn!")
        self.assertTrue(r["ended"])

    def test_customer_update(self):
        self.twiml("/voice/incoming", CallSid="CA1", From="+84901", To="+84999")
        r = self.client.put("/api/customers/+84901", json={"name": "Anh Minh"}, headers=ADMIN)
        self.assertEqual(r.json()["name"], "Anh Minh")
        self.assertEqual(self.client.put("/api/customers/+849", json={}, headers=ADMIN).status_code, 404)

    def test_tts(self):
        r = self.client.post("/api/test/tts", params={"text": "xin chào"}, headers=ADMIN)
        self.assertEqual((r.status_code, r.content), (200, b"mp3"))
        self.container.tts = FakeTTS(None)
        r = self.client.post("/api/test/tts", params={"text": "xin chào"}, headers=ADMIN)
        self.assertEqual(r.status_code, 500)


if __name__ == "__main__":
    unittest.main()
