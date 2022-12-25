
import argparse
from dataclasses import dataclass
import asyncio
import signal

@dataclass
class AgentConfig:
    server_addr: str
    heart_beat_dur: int

class ModelBoxAgent:
    def __init__(self, config) -> None:
        super().__init__()
        self._config = config

    async def heartbeat(self):
        while True:
            print("heart beat")
            await asyncio.sleep(self._config.heart_beat_dur)

    async def agent_runner(self):
        try:
            await self.heartbeat()
        except asyncio.CancelledError:
            print("exiting")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(prog="modelbox-agent", description="modelbox agent")
    parser.add_argument('--server_addr', default="localhost:8081", help="address of the admin api")
    parser.add_argument('--heart_beat_dur', default=5, help="heart beat duration")
    args = parser.parse_args()
    agent = ModelBoxAgent(config=AgentConfig(args.server_addr, args.heart_beat_dur))
    loop = asyncio.get_event_loop()
    main_task = asyncio.ensure_future(agent.agent_runner())
    for sig in [signal.SIGTERM, signal.SIGINT]:
        loop.add_signal_handler(sig, main_task.cancel)

    try:
        loop.run_until_complete(main_task)
    finally:
        loop.close()

