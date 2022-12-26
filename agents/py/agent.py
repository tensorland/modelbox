
import argparse
from dataclasses import dataclass
import asyncio
import signal
import logging

from client import AdminClient
from modelbox import admin_pb2
import platform

@dataclass
class AgentConfig:
    server_addr: str
    heartbeat_dur: int
    name: str

@dataclass
class Node:
    host: str

    def get_id(self) -> str:
        #TODO Probe the node for more metadata
        #and hash them into a node
        return self.host

logger = logging.getLogger(__name__)

class ModelBoxAgent:
    def __init__(self, config: AgentConfig) -> None:
        super().__init__()
        self._config: AgentConfig = config
        self._client: AdminClient = AdminClient(self._config.server_addr)
        self._node: Node = Node(host=platform.node())
        self._server_node_id: str = None

    async def register_node(self):
        logger.info(f"registering node")
        node_info = admin_pb2.NodeInfo(host_name=platform.node(), ip_addr="", arch="x86")
        while True:
            try:
                resp = self._client.register_agent(node_info=node_info, name=self._config.name)
                self._server_node_id =  resp.node_id
                logger.info(f"registered node, server node id:{self._server_node_id}")
                break
            except Exception as ex:
                logger.error(f"unable to register agent with server {ex}. Trying again in {self._config.heartbeat_dur}")
                await asyncio.sleep(self._config.heartbeat_dur)
                continue
        

    async def heartbeat(self):
        logger.info(f"starting to heartbeat ever {self._config.heartbeat_dur}s")
        while True:
            try:
                response = self._client.heartbeat(node_id=self._server_node_id)
            except Exception as ex:
                logger.error(f"couldn't register heartbeat {ex}")
            await asyncio.sleep(self._config.heartbeat_dur)

    async def poll_for_work(self):
        while True:
            try:
                respone = self._client.get_runnable_actions()
            except Exception as ex:
                logger.error(f"unable to get work {ex}")
            await asyncio.sleep(self._config.heartbeat_dur)
        pass

    async def agent_runner(self):
        try:
            # Register Node
            await self.register_node()
            
            # Start the heartbeat
            self.heartbeat()

            # Keep waiting for available work
            await self.poll_for_work()
        except asyncio.CancelledError:
            logger.info("exiting agent")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(prog="modelbox-agent", description="modelbox agent")
    parser.add_argument('--server_addr', default="localhost:8081", help="address of the admin api")
    parser.add_argument('--heartbeat_dur', default=5, help="heart beat duration")
    parser.add_argument("--workers", nargs="+", help="list of workers(separated by space)")
    parser.add_argument("--name", default="default-agent", help="agent name")
    args = parser.parse_args()

    agent = ModelBoxAgent(config=AgentConfig(args.server_addr, args.heartbeat_dur, name=args.name))
    loop = asyncio.get_event_loop()
    main_task = asyncio.ensure_future(agent.agent_runner())
    for sig in [signal.SIGTERM, signal.SIGINT]:
        loop.add_signal_handler(sig, main_task.cancel)

    try:
        loop.run_until_complete(main_task)
    finally:
        loop.close()

