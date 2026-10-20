import os
import unittest

from fastapi import Depends, FastAPI
from fastapi.testclient import TestClient
from twilio.request_validator import RequestValidator

from app.adapters.inbound.api import twilio_auth


class TwilioAuthTest(unittest.TestCase):
    def setUp(self):
        os.environ["TWILIO_AUTH_TOKEN"] = "tok"
        os.environ["PUBLIC_BASE_URL"] = "https://calls.example.com"
        app = FastAPI()

        @app.post("/voice/status", dependencies=[Depends(twilio_auth.verify_twilio)])
        def status():
            return {"ok": True}

        self.client = TestClient(app)

    def tearDown(self):
        os.environ.pop("TWILIO_AUTH_TOKEN", None)
        os.environ.pop("PUBLIC_BASE_URL", None)

    def test_signature(self):
        form = {"CallSid": "CA1", "CallStatus": "completed"}
        good = RequestValidator("tok").compute_signature("https://calls.example.com/voice/status", form)
        self.assertEqual(self.client.post("/voice/status", data=form, headers={"X-Twilio-Signature": good}).status_code, 200)
        self.assertEqual(self.client.post("/voice/status", data=form, headers={"X-Twilio-Signature": "bad"}).status_code, 403)
        self.assertEqual(self.client.post("/voice/status", data=form).status_code, 403)


if __name__ == "__main__":
    unittest.main()
