from datetime import datetime
import time
from typing import List, Optional, Callable, Any, TypedDict, Literal

from .exceptions import (
    JobError,
    JobFailedError,
)


class JobExecResponse(TypedDict):
    """
    Response returned when a job is created.
    """
    job_id: str
    status: Literal["queued", "running", "completed", "failed"]


class JobOutputResponse(TypedDict):
    """
    Output payload of a completed job.
    """
    console_output: Optional[str]
    errout: Optional[str]
    status_code: Optional[int]


class JobStatusResponse(TypedDict):
    """
    Full job status representation returned by the API.
    """
    job_id: str
    cmd: List[str]
    status: Literal["queued", "running", "completed", "failed"]
    created_at: datetime
    output: Optional[JobOutputResponse]


class Jobs:
    """
    High-level Jobs SDK for executing and managing commands inside a Bastion session.

    Jobs represent asynchronous command execution units that can be:
    - created (run)
    - polled (get / wait)
    - streamed (watch)
    """

    def __init__(self, api):
        """
        Initialize Jobs SDK wrapper.

        Args:
            api: Generated OpenAPI JobsApi client.
        """
        self._api = api

    def run(self, session_id: str, cmd: List[str]) -> JobExecResponse:
        """
        Execute a command inside a sandbox session.

        This creates a new job that runs asynchronously on the Bastion runtime.

        Args:
            session_id: Target session where command should run.
            cmd: Command and arguments to execute.

        Returns:
            JobExecResponse:
                - job_id: Unique job identifier
                - status: Initial job state (usually 'queued')
        """
        try:
            res = self._api.execute_command(
                id=session_id,
                job_exec_request={"cmd": cmd},
            )

            # normalize to dict (IMPORTANT: avoids object-subscript bugs)
            return {
                "job_id": res.job_id,
                "status": res.status.value,
            }

        except Exception as e:
            raise JobError(
                f"Failed to run job in session '{session_id}'"
            ) from e

    def get(self, session_id: str, job_id: str) -> JobStatusResponse:
        """
        Retrieve the current status of a job.

        Args:
            session_id: Session containing the job.
            job_id: Job identifier.

        Returns:
            JobStatusResponse:
                - job_id
                - status
                - output
        """
        try:
            res = self._api.get_job_status(
                id=session_id,
                job_id=job_id,
            )

            output = res.output

            # normalize response into dict
            return {
                "job_id": res.job_id,
                "cmd": res.cmd,
                "status": res.status.value,
                "created_at": res.created_at,
                "output": {
                    "console_output": getattr(output, "console_output", None),
                    "errout": getattr(output, "errout", None),
                    "status_code": getattr(output, "status_code", None),
                } if output else None,
            }

        except Exception as e:
            raise JobError(
                f"Failed to fetch status for job '{job_id}' in session '{session_id}'"
            ) from e

    def wait(
        self,
        session_id: str,
        job_id: str,
        interval: float = 1.0,
        timeout: Optional[float] = None,
    ) -> JobStatusResponse:
        """
        Block until a job completes or fails.

        This polls the Bastion API at a fixed interval until the job reaches a terminal state.

        Args:
            session_id: Session containing the job.
            job_id: Job to wait for.
            interval: Polling interval in seconds (default: 1.0s).
            timeout: Maximum wait time in seconds. None means no timeout.

        Returns:
            JobStatusResponse: Final job state (completed or failed).

        Raises:
            TimeoutError: If job does not finish within the specified timeout.
        """

        start = time.time()

        while True:
            try:
                status = self.get(session_id, job_id)
            except Exception as e:
                raise JobError(
                    f"Failed while polling job '{job_id}'"
                ) from e

            if status["status"] in ("completed", "failed"):

                if status["status"] == "failed":
                    raise JobFailedError(status.get("output"))

                return status

            if timeout is not None and (time.time() - start) > timeout:
                raise TimeoutError(f"Job {job_id} timed out")

            time.sleep(interval)

    def run_and_wait(
        self,
        session_id: str,
        cmd: List[str],
        interval: float = 1.0,
        timeout: Optional[float] = None,
    ) -> JobStatusResponse:
        """
        Execute a command and wait for its completion in one call.

        This is a convenience helper combining `run()` + `wait()`.

        Args:
            session_id: Target session.
            cmd: Command to execute.
            interval: Polling interval in seconds.
            timeout: Maximum execution time.

        Returns:
            JobStatusResponse: Final job result.
        """

        job = self.run(session_id, cmd)
        return self.wait(session_id, job["job_id"], interval, timeout)

    def watch(
        self,
        session_id: str,
        job_id: str,
        callback: Callable[[JobStatusResponse], None],
        interval: float = 1.0,
    ) -> JobStatusResponse:
        """
        Stream job status updates in real-time via callback.

        This continuously polls job status and triggers the callback whenever
        the status changes.

        Useful for:
        - CLI live logs
        - UI progress updates
        - debugging job execution

        Args:
            session_id: Session containing the job.
            job_id: Job to monitor.
            callback: Function called with updated status object.
            interval: Polling interval in seconds.

        Returns:
            JobStatusResponse: Final job state once completed or failed.
        """

        last_status = None

        while True:
            try:
                status = self.get(session_id, job_id)
            except Exception as e:
                raise JobError(
                    f"Failed while watching job '{job_id}'"
                ) from e

            if status["status"] != last_status:
                callback(status)
                last_status = status["status"]

            if status["status"] in ("completed", "failed"):

                if status["status"] == "failed":
                    raise JobFailedError(status.get("output"))

                return status

            time.sleep(interval)