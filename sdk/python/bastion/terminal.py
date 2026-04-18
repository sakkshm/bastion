import json
import threading
import time
from typing import Callable, Optional
import websocket

from .exceptions import TerminalConnectionError, TerminalSendError


class Terminal:
    """
    High-level Terminal SDK wrapper for Bastion sessions.

    Provides a persistent WebSocket-based terminal connection to a running
    sandbox session, enabling real-time command execution and streaming output.
    """

    def __init__(self, base_url: str, api_token: str):
        """
        Initialize Terminal SDK instance.

        This creates a terminal client that can connect to any session
        at runtime via `connect()`.

        Args:
            base_url: Backend WebSocket base URL
        """
        self.base_url = base_url
        self.api_token = api_token or ""
        self.ws: Optional[websocket.WebSocketApp] = None
        self.client_id: Optional[str] = None
        self.session_id: Optional[str] = None
        self._connected = False
        self._thread: Optional[threading.Thread] = None

    def connect(
        self,
        session_id: str,
        on_message: Callable[[dict], None],
        on_open: Optional[Callable[[], None]] = None,
        on_close: Optional[Callable[[], None]] = None,
        on_error: Optional[Callable[[Exception], None]] = None,
    ):
        """
        Establish a WebSocket connection to the terminal session.

        This upgrades the session into a live terminal stream and binds
        the SDK instance to the provided session_id.

        Args:
            session_id: Active sandbox session identifier
            on_message: Callback invoked for every incoming message
            on_open: Optional callback triggered when connection opens
            on_close: Optional callback triggered when connection closes
            on_error: Optional callback triggered on WebSocket errors
        """

        if self.ws and self._connected:
            raise TerminalConnectionError("Terminal already connected")

        self.session_id = session_id

        # normalize scheme
        if self.base_url.startswith("https://"):
            ws_base = self.base_url.replace("https://", "wss://", 1)
        elif self.base_url.startswith("http://"):
            ws_base = self.base_url.replace("http://", "ws://", 1)
        else:
            raise TerminalConnectionError("Invalid base_url scheme")

        url = f"{ws_base}/session/{session_id}/terminal?api_token={self.api_token}"

        def _on_message(ws, message: str):
            try:
                data = json.loads(message)

                if data.get("type") == "init":
                    client_id = data.get("client_id")
                    if not client_id:
                        raise TerminalConnectionError("Invalid init payload: missing client_id")

                    self.client_id = client_id
                    self._connected = True

                on_message(data)

            except Exception as e:
                self._connected = False
                if on_error:
                    on_error(e)

        def _on_open(ws):
            try:
                if on_open:
                    on_open()
            except Exception as e:
                if on_error:
                    on_error(e)

        def _on_close(ws, close_status_code, close_msg):
            self._connected = False
            try:
                if on_close:
                    on_close()
            except Exception as e:
                if on_error:
                    on_error(e)

        def _on_error(ws, error):
            self._connected = False
            if on_error:
                on_error(error)

        try:
            self.ws = websocket.WebSocketApp(
                url,
                on_message=_on_message,
                on_open=_on_open,
                on_close=_on_close,
                on_error=_on_error,
            )

            # run in background thread (non-blocking)
            self._thread = threading.Thread(target=self.ws.run_forever, daemon=True)
            self._thread.start()

            # wait for init handshake (client_id)
            timeout = 5
            start = time.time()

            while not self._connected:
                if time.time() - start > timeout:
                    raise TerminalConnectionError("Terminal connection timeout")
                time.sleep(0.05)

        except Exception as e:
            raise TerminalConnectionError(
                f"Failed to connect terminal: {e}"
            ) from e

    def send_input(self, input_text: str):
        """
        Send raw input to the terminal session.

        This simulates user keyboard input into the remote shell.

        Args:
            input_text: Command or input string to send (e.g. "ls -la\n")
        """

        if not self.ws or not self._connected:
            raise TerminalConnectionError("WebSocket not connected")

        if not self.client_id or not self.session_id:
            raise TerminalConnectionError(
                "Terminal not initialized (missing client_id/session_id)"
            )

        payload = {
            "type": "term_input",
            "client_id": self.client_id,
            "session_id": self.session_id,
            "payload": {
                "input": input_text
            }
        }

        try:
            self.ws.send(json.dumps(payload))
        except Exception as e:
            self._connected = False
            raise TerminalSendError(f"Failed to send input: {e}") from e

    def exec(self, cmd: str):
        """
        Execute a shell command in the terminal session.

        This sends a direct execution request to the sandbox environment.

        Args:
            cmd: Shell command to execute (e.g. "docker ps")
        """

        if not self.ws or not self._connected:
            raise TerminalConnectionError("WebSocket not connected")

        if not self.client_id or not self.session_id:
            raise TerminalConnectionError(
                "Terminal not initialized (missing client_id/session_id)"
            )

        payload = {
            "type": "term_input",  
            "client_id": self.client_id,
            "session_id": self.session_id,
            "payload": {
                "cmd": cmd
            }
        }

        try:
            self.ws.send(json.dumps(payload))
        except Exception as e:
            self._connected = False
            raise TerminalSendError(f"Failed to execute command: {e}") from e

    def close(self):
        """
        Close the terminal WebSocket connection.

        This terminates the live session stream but does not destroy the
        underlying sandbox session.
        """
        try:
            if self.ws:
                self.ws.close()
                self._connected = False
        except Exception as e:
            raise TerminalConnectionError(
                f"Failed to close terminal: {e}"
            ) from e