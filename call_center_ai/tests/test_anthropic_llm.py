import asyncio
import unittest
from types import SimpleNamespace

from app.adapters.outbound.llm.anthropic_llm import CALL_CONNECTED, FALLBACK_BETA, AnthropicLLM
from app.domain.models import ChatTurn


class _Endpoint:
    def __init__(self, response):
        self.response = response
        self.requests = []

    async def create(self, **kwargs):
        self.requests.append(kwargs)
        return self.response


def _client(response):
    beta = _Endpoint(response)
    plain = _Endpoint(response)
    return SimpleNamespace(beta=SimpleNamespace(messages=beta), messages=plain), beta, plain


def _response(*blocks, stop_reason="end_turn"):
    return SimpleNamespace(content=list(blocks), stop_reason=stop_reason, stop_details=None)


HISTORY = [ChatTurn("assistant", "Xin chào!"), ChatTurn("user", "Phí ship bao nhiêu?")]


class AnthropicLLMTest(unittest.TestCase):
    def test_request_shape_with_refusal_fallback(self):
        client, beta, plain = _client(
            _response(
                SimpleNamespace(type="thinking", thinking=""),
                SimpleNamespace(type="text", text="Miễn phí từ 300 nghìn."),
            )
        )
        llm = AnthropicLLM("k", "claude-opus-5-5", client=client)
        self.assertEqual(asyncio.run(llm.generate(HISTORY, "SYS")), "Miễn phí từ 300 nghìn.")
        self.assertEqual(plain.requests, [])
        req = beta.requests[0]
        self.assertEqual(req["model"], "claude-opus-5-5")
        self.assertEqual(req["betas"], [FALLBACK_BETA])
        self.assertEqual(req["fallbacks"], "default")
        self.assertEqual(req["output_config"], {"effort": "low"})
        self.assertEqual(req["system"], "SYS")
        self.assertNotIn("temperature", req)
        self.assertEqual(
            req["messages"],
            [
                {"role": "user", "content": CALL_CONNECTED},
                {"role": "assistant", "content": "Xin chào!"},
                {"role": "user", "content": "Phí ship bao nhiêu?"},
            ],
        )

    def test_without_fallback_uses_plain_endpoint(self):
        client, beta, plain = _client(_response(SimpleNamespace(type="text", text="ok")))
        llm = AnthropicLLM("k", "claude-haiku-5-5", refusal_fallback=False, client=client)
        self.assertEqual(asyncio.run(llm.generate([ChatTurn("user", "hi")])), "ok")
        self.assertEqual(beta.requests, [])
        self.assertEqual(plain.requests[0]["messages"], [{"role": "user", "content": "hi"}])

    def test_refusal_returns_none(self):
        client, _, _ = _client(_response(stop_reason="refusal"))
        self.assertIsNone(asyncio.run(AnthropicLLM("k", "m", client=client).generate(HISTORY)))


if __name__ == "__main__":
    unittest.main()
