class BastionError(Exception):
    """
    Base exception for all Bastion SDK errors.
    """
    def __init__(self, message: str):
        super().__init__(message)
        self.message = message


class APIError(BastionError):
    """
    Raised when the underlying API returns an error response.
    """
    def __init__(self, message: str, status_code: int | None = None):
        super().__init__(message)
        self.status_code = status_code


class SessionError(BastionError):
    """
    Raised for session-related failures (create/start/stop/delete).
    """
    pass

class SessionStateError(BastionError):
    """
    Raised for session state related failures.
    """
    pass


class JobError(BastionError):
    """
    Raised for job execution or polling failures.
    """
    pass

class JobFailedError(BastionError):
    """
    Raised for job failures.
    """
    pass


class FileError(BastionError):
    """
    Base class for file-related errors.
    """
    pass


class FileUploadError(FileError):
    """
    Raised when file upload fails.
    """
    pass


class FileListError(FileError):
    """
    Raised when listing files fails.
    """
    pass


class FileDeleteError(FileError):
    """
    Raised when file deletion fails.
    """
    pass


class TerminalError(BastionError):
    """
    Base class for terminal (WebSocket) errors.
    """
    pass


class TerminalConnectionError(TerminalError):
    """
    Raised when WebSocket connection fails or drops unexpectedly.
    """
    pass


class TerminalSendError(TerminalError):
    """
    Raised when sending data through terminal WebSocket fails.
    """
    pass