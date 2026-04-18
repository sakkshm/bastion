from datetime import datetime
from typing import Any, TypedDict, Literal

from .exceptions import (
    SessionError,
    SessionStateError,
)


class CreateSessionResponse(TypedDict):
    session_id: str
    status: Literal["created", "starting", "running", "stopped", "deleted"]
    created_at: datetime


class DeleteSessionResponse(TypedDict):
    session_id: str
    status: Literal["deleted", "stopped", "failed"]


class SessionStatusResponse(TypedDict):
    id: str
    container_id: str
    created_at: datetime
    last_used_at: str
    status: Literal["created", "starting", "running", "stopped", "deleted"]


class Sessions:
    """
    High-level Sessions SDK wrapper for Bastion Sandbox API.

    Provides lifecycle management for sandbox sessions including:
    creation, startup, shutdown, status tracking, and deletion.
    """

    def __init__(self, api):
        """
        Initialize Sessions SDK wrapper.

        Args:
            api: Generated OpenAPI SessionsApi client.
        """
        self.api = api

    def create(self) -> CreateSessionResponse:
        """
        Create a new sandbox session.

        This initializes a fresh isolated environment that can be used to
        execute jobs, run commands, and manage files.

        Returns:
            CreateSessionResponse:
                - session_id: Unique identifier of the created session
                - status: Current session state (usually 'created')
                - created_at: Timestamp of session creation
        """
        try:
            res = self.api.create_session()
            return {
                "session_id": res.session_id,
                "status": res.status.value,
                "created_at": res.created_at,
            }

        except Exception as e:
            raise SessionError(
                "Failed to create session"
            ) from e

    def start(self, session_id: str) -> SessionStatusResponse:
        """
        Start a previously created sandbox session.

        This transitions the session into a running state and provisions
        the underlying container environment.

        Args:
            session_id: Target session to start.

        Returns:
            SessionStatusResponse:
                - id: Session identifier
                - container_id: Underlying runtime container ID
                - created_at: Session creation timestamp
                - last_used_at: Last activity timestamp
                - status: Current session state
        """
        try:
            res = self.api.start_session(id=session_id)
            return {
                "id": res.id,
                "container_id": res.container_id,
                "created_at": res.created_at,
                "last_used_at": res.last_used_at,
                "status": res.status.value,
            }

        except Exception as e:
            raise SessionStateError(
                f"Failed to start session '{session_id}'"
            ) from e

    def stop(self, session_id: str) -> SessionStatusResponse:
        """
        Stop a running sandbox session.

        This halts execution inside the session while preserving its state
        for potential restart or inspection.

        Args:
            session_id: Target session to stop.

        Returns:
            SessionStatusResponse:
                - id: Session identifier
                - container_id: Underlying runtime container ID
                - created_at: Session creation timestamp
                - last_used_at: Last activity timestamp
                - status: Updated session state
        """
        try:
            res = self.api.stop_session(id=session_id)
            return {
                "id": res.id,
                "container_id": res.container_id,
                "created_at": res.created_at,
                "last_used_at": res.last_used_at,
                "status": res.status.value,
            }

        except Exception as e:
            raise SessionStateError(
                f"Failed to stop session '{session_id}'"
            ) from e

    def status(self, session_id: str) -> SessionStatusResponse:
        """
        Retrieve the current status of a sandbox session.

        This provides real-time information about the session lifecycle,
        including runtime state and metadata.

        Args:
            session_id: Target session to query.

        Returns:
            SessionStatusResponse:
                - id: Session identifier
                - container_id: Underlying runtime container ID
                - created_at: Session creation timestamp
                - last_used_at: Last activity timestamp
                - status: Current session state
        """
        try:
            res = self.api.get_session_status(id=session_id)
            return {
                "id": res.id,
                "container_id": res.container_id,
                "created_at": res.created_at,
                "last_used_at": res.last_used_at,
                "status": res.status.value,
            }

        except Exception as e:
            raise SessionError(
                f"Failed to fetch status for session '{session_id}'"
            ) from e

    def delete(self, session_id: str) -> DeleteSessionResponse:
        """
        Delete a sandbox session and release all associated resources.

        This permanently removes the session and its underlying runtime
        environment.

        Args:
            session_id: Target session to delete.

        Returns:
            DeleteSessionResponse:
                - session_id: Identifier of deleted session
                - status: Final deletion state
        """
        try:
            res = self.api.delete_session(id=session_id)
            return {
                "session_id": res.session_id,
                "status": res.status.value,
            }

        except Exception as e:
            raise SessionError(
                f"Failed to delete session '{session_id}'"
            ) from e