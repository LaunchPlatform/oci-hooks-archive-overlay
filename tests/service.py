import asyncio.subprocess
import copy
import typing

from containers import Container
from containers import Mount
from containers import WindowsContainersService
from containers.services.windows import to_wsl_path

from .data_types import ImageMountWithArchive


class WindowsContainersServiceWithArchive(WindowsContainersService):
    def _filter_mount(self, mount: Mount) -> Mount:
        if not isinstance(mount, ImageMountWithArchive):
            return mount
        if mount.archive_to is None:
            return mount
        mount = copy.deepcopy(mount)
        mount.archive_to = to_wsl_path(mount.archive_to)
        if mount.archive_success is not None:
            mount.archive_success = to_wsl_path(mount.archive_success)
        return mount

    def run(
        self,
        container: Container,
        stdin: typing.Optional[int] = None,
        stdout: typing.Optional[int] = None,
        stderr: typing.Optional[int] = None,
        log_level: typing.Optional[str] = None,
    ) -> typing.AsyncContextManager[asyncio.subprocess.Process]:
        container = copy.deepcopy(container)
        container.mounts = list(map(self._filter_mount, container.mounts))
        return super().run(container, stdin, stdout, stderr, log_level)
