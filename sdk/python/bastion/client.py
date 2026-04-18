from generated_client import ApiClient, Configuration
from generated_client.api import (
    SessionsApi,
    JobsApi,
    FilesApi,
    SystemApi,
    TerminalApi,
)

from .sessions import Sessions
from .jobs import Jobs
from .files import Files
from .terminal import Terminal

from .exceptions import (
    APIError,
)


class Bastion:
    """
    Bastion SDK client.
    Wraps generated OpenAPI client with a stable interface.
    """

    def __init__(
        self,
        base_url: str,
        api_key: str | None = None,
        timeout: int = 30,
        debug: bool = False,
    ):
        try:
            self._config = Configuration(
                host=base_url,
                access_token=api_key,
            )

            self._config.timeout = timeout
            self._config.debug = debug

            self._client = ApiClient(self._config)

            # internal APIs
            self._sessions_api = SessionsApi(self._client)
            self._jobs_api = JobsApi(self._client)
            self._files_api = FilesApi(self._client)
            self._system_api = SystemApi(self._client)
            self._terminal_api = TerminalApi(self._client)

            # public SDK modules
            self.sessions = Sessions(self._sessions_api)
            self.jobs = Jobs(self._jobs_api)
            self.files = Files(self._files_api)
            self.terminal = Terminal(base_url=base_url, api_token=api_key)

        except Exception as e:
            raise APIError(
                "Failed to initialize Bastion SDK client",
                payload=str(e),
            )

    def close(self) -> None:
        """Explicit cleanup of underlying HTTP client."""
        try:
            if hasattr(self._client, "close"):
                self._client.close()
        except Exception as e:
            raise APIError(
                "Failed to close Bastion client",
                payload=str(e),
            )

    def __enter__(self) -> "Bastion":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        self.close()