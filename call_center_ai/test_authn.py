import base64
import json
import threading
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer

import authn


def token(user_id, role):
    payload = base64.urlsafe_b64encode(json.dumps({"sub": user_id, "role": role}).encode()).decode().rstrip("=")
    return "h." + payload + ".s"


class _UserService(BaseHTTPRequestHandler):
    def do_POST(self):
        body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
        t = body["session_token"]
        if t == "down":
            self.send_response(500)
            self.end_headers()
            return
        parts = t.split(".")
        sub = ""
        if len(parts) == 3:
            sub = json.loads(base64.urlsafe_b64decode(parts[1] + "=" * (-len(parts[1]) % 4))).get("sub", "")
        out = json.dumps({"valid": bool(sub), "user_id": sub}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(out)

    def log_message(self, *args):
        pass


class AuthnTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = HTTPServer(("127.0.0.1", 0), _UserService)
        threading.Thread(target=cls.server.serve_forever, daemon=True).start()
        cls.url = "http://127.0.0.1:%d" % cls.server.server_port

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()

    def test_rules(self):
        strict = authn.Authenticator(self.url, ["s3cret"])
        user = strict.authenticate("Bearer " + token("7", "user"), None)
        self.assertEqual(user.user_id, "7")
        self.assertFalse(user.is_admin())
        self.assertTrue(user.owns(7))
        self.assertFalse(user.owns(8))
        self.assertTrue(strict.authenticate("Bearer " + token("1", "admin"), None).is_admin())
        self.assertTrue(strict.authenticate(None, "s3cret").service)
        self.assertIsNone(strict.authenticate(None, None))
        for bad in [("Bearer nope", None), (None, "wrong")]:
            with self.assertRaises(authn.AuthError) as e:
                strict.authenticate(*bad)
            self.assertEqual(e.exception.status, 401)
        with self.assertRaises(authn.AuthError) as e:
            strict.authenticate("Bearer down", None)
        self.assertEqual(e.exception.status, 503)
        self.assertTrue(authn.Authenticator(self.url, []).authenticate(None, None).service)
        with self.assertRaises(authn.AuthError) as e:
            authn.require_owner(user, 8)
        self.assertEqual(e.exception.status, 403)
        with self.assertRaises(authn.AuthError) as e:
            authn.require_admin(user)
        self.assertEqual(e.exception.status, 403)
        with self.assertRaises(authn.AuthError) as e:
            authn.require(None)
        self.assertEqual(e.exception.status, 401)


if __name__ == "__main__":
    unittest.main()
