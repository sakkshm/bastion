from typing import BinaryIO, Optional, Union, Any, TypedDict, List
from datetime import datetime

from .exceptions import (
    FileUploadError,
    FileListError,
    FileDeleteError,
)


class UploadResponse(TypedDict):
    status: str
    path: str


class FileEntry(TypedDict):
    name: str
    is_dir: bool
    size: int
    mode: str
    modified_at: datetime


class ListResponse(TypedDict):
    page: int
    limit: int
    total: int
    total_pages: int
    files: List[FileEntry]


class DeleteResponse(TypedDict):
    status: str


class Files:
    """
    File operations for a Bastion session.
    """

    def __init__(self, api: Any):
        self._api = api

    def upload(
        self,
        session_id: str,
        file: Union[BinaryIO, bytes],
        path: str = "",
    ) -> UploadResponse:
        """
        Upload a file into a running Bastion session.

        Args:
            session_id: Active session identifier
            file: File object or raw bytes to upload
            path: Path of uploaded file in session

        Returns:
            UploadResponse containing status and uploaded file path
        """
        try:
            res = self._api.upload_file(
                id=session_id,
                file=file,
                metadata={"path": path},
            )

            return {
                "status": res.status,
                "path": res.path,
            }

        except Exception as e:
            raise FileUploadError(
                f"Upload failed in session '{session_id}' at path '{path}'"
            ) from e

    def list(
        self,
        session_id: str,
        path: str,
        *,
        page: Optional[int] = None,
        limit: Optional[int] = None,
    ) -> ListResponse:
        """
        List the contents of a directory in a running Bastion session.

        Args:
            session_id: Active session identifier
            path: Path of directory to list
            page: Page index (for pagination)
            limit: Limit of responses per page (for pagination)

        Returns:
            ListResponse containing paginated file entries
        """
        try:
            res = self._api.list_files(
                id=session_id,
                path=path,
                page=page,
                limit=limit,
            )

            return {
                "page": res.page,
                "limit": res.limit,
                "total": res.total,
                "total_pages": res.total_pages,
                "files": [
                    {
                        "name": f.name,
                        "is_dir": f.is_dir,
                        "size": f.size,
                        "mode": f.mode,
                        "modified_at": f.mod_time,
                    }
                    for f in res.files
                ],
            }

        except Exception as e:
            raise FileListError(
                f"Failed to list files at '{path}' in session '{session_id}'"
            ) from e

    def delete(
        self,
        session_id: str,
        path: str,
    ) -> DeleteResponse:
        """
        Delete a file from a session.

        Args:
            session_id: Active session identifier
            path: File path to delete

        Returns:
            DeleteResponse containing operation status
        """
        try:
            res = self._api.delete_file(
                id=session_id,
                delete_request={"path": path},
            )

            return {
                "status": res.status,
            }

        except Exception as e:
            raise FileDeleteError(
                f"Failed to delete file at '{path}' in session '{session_id}'"
            ) from e