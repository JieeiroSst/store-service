import asyncio
import sys

import httpx
from sqlalchemy import text

from app.adapters.outbound.persistence.db import build_engine, create_schema
from app.config import settings
from app.container import Container
from app.domain import playbook
from app.domain.models import ChatTurn


async def check_database(container: Container) -> bool:
    engine = build_engine(settings.sqlalchemy_url)
    try:
        create_schema(engine)
        with engine.connect() as conn:
            return conn.execute(text("SELECT 1")).scalar() == 1
    finally:
        engine.dispose()


async def check_llm(container: Container) -> bool:
    answer = await container.llm.generate([ChatTurn("user", "Xin chào")], "Bạn là trợ lý thân thiện.")
    if answer:
        print(f"   Response: {answer[:100]}")
    return bool(answer)


async def check_tts(container: Container) -> bool:
    audio = await container.tts.synthesize("Xin chào, đây là test text to speech", settings.voice_language)
    if audio:
        print(f"   Audio size: {len(audio)} bytes")
    return bool(audio)


async def check_scenarios(container: Container) -> bool:
    for key, data in playbook.DEFAULT_SCENARIOS.items():
        print(f"   - {key}: {data['name']}")
    return True


async def check_api(container: Container) -> bool:
    async with httpx.AsyncClient() as client:
        response = await client.get(f"http://localhost:{settings.server_port}/")
    return response.status_code == 200


async def main() -> int:
    container = Container.from_settings(settings)
    checks = [
        ("Database", check_database),
        ("LLM", check_llm),
        ("TTS", check_tts),
        ("Scenarios", check_scenarios),
    ]
    if "--with-api" in sys.argv:
        checks.append(("API", check_api))

    passed = 0
    for name, check in checks:
        try:
            ok = await check(container)
        except Exception as e:
            print(f"   {name} error: {e}")
            ok = False
        passed += ok
        print(f"{name:12} {'PASSED' if ok else 'FAILED'}")
    container.shutdown()
    print(f"{passed}/{len(checks)} checks passed")
    return 0 if passed == len(checks) else 1


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
