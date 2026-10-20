import os
import threading
import unittest
from http.server import HTTPServer

from fastapi import Depends, FastAPI
from fastapi.testclient import TestClient

from app.adapters.inbound.api import auth_deps
from tests.test_authn import _UserService, token


class AuthDepsTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = HTTPServer(("127.0.0.1", 0), _UserService)
        threading.Thread(target=cls.server.serve_forever, daemon=True).start()
        os.environ["USER_SERVICE_URL"] = "http://127.0.0.1:%d" % cls.server.server_port
        os.environ["INTERNAL_API_TOKENS"] = "s3cret"
        auth_deps._authenticator = None
        app = FastAPI()

        @app.get("/admin", dependencies=[Depends(auth_deps.require_admin)])
        def admin():
            return {"ok": True}

        @app.get("/users/{user_id}")
        def own(user_id: str, p=Depends(auth_deps.require_user)):
            auth_deps.ensure_owner(p, user_id)
            return {"ok": True}

        cls.client = TestClient(app)

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()
        auth_deps._authenticator = None

    def get(self, path, bearer=None, internal=None):
        headers = {}
        if bearer:
            headers["Authorization"] = "Bearer " + bearer
        if internal:
            headers["X-Internal-Token"] = internal
        return self.client.get(path, headers=headers).status_code

    def test_routes(self):
        user = token("7", "user")
        self.assertEqual(self.get("/admin"), 401)
        self.assertEqual(self.get("/admin", bearer=user), 403)
        self.assertEqual(self.get("/admin", bearer=token("1", "admin")), 200)
        self.assertEqual(self.get("/admin", internal="s3cret"), 200)
        self.assertEqual(self.get("/admin", internal="nope"), 401)
        self.assertEqual(self.get("/users/7", bearer=user), 200)
        self.assertEqual(self.get("/users/8", bearer=user), 403)
        self.assertEqual(self.get("/users/8"), 401)


if __name__ == "__main__":
    unittest.main()
