import functools
import os
import pathlib
import sys
import tarfile
import typing

import pytest
import pytest_asyncio
from containers import Container
from containers import ContainersService

from .data_types import ImageMountWithArchive
from .helpers import assert_has_whiteout_device
from .helpers import assert_has_whiteout_file
from .providers import PodmanWithArchive
from .service import WindowsContainersServiceWithArchive


# Are we using rootless whiteout?
USE_ROOTLESS_WHITEOUT = os.getenv("USE_ROOTLESS_WHITEOUT", None) == "1"


@pytest.fixture
def log_level() -> typing.Optional[str]:
    return "debug"


@pytest.fixture
def podman() -> PodmanWithArchive:
    return PodmanWithArchive()


@pytest.fixture
def containers(podman: PodmanWithArchive) -> ContainersService:
    if sys.platform == "win32":
        return WindowsContainersServiceWithArchive(podman)
    return ContainersService(podman)


@pytest.fixture
def image() -> str:
    return "alpine:3.18.2"


@pytest.fixture
def image_mount(image: str) -> ImageMountWithArchive:
    return ImageMountWithArchive(
        source=image,
        target="/data",
        read_write=True,
    )


class TestArchive:
    @pytest_asyncio.fixture(autouse=True)
    async def _request_fixtures(
        self,
        tmp_path: pathlib.Path,
        log_level: str,
        containers: ContainersService,
        image_mount: ImageMountWithArchive,
        image: str,
    ):
        self.tmp_path = tmp_path
        self.log_level = log_level
        self.containers = containers
        self.image = image
        self.image_mount = image_mount
        self.run = functools.partial(self.containers.run, log_level=log_level)
        await self.containers.load_image(self.image)

    @pytest.mark.asyncio
    async def test_archive_with_copy(self):
        archive_to = self.tmp_path / "archive"
        self.image_mount.archive_to = archive_to
        container = Container(
            command=("/bin/sh", "-c", "echo hello > /data/file"),
            image=self.image,
            mounts=[self.image_mount],
            network="none",
        )
        async with self.run(container) as proc:
            assert await proc.wait() == 0

        file_path = archive_to / "file"
        assert file_path.exists()
        assert file_path.read_text().strip() == "hello"

    @pytest.mark.asyncio
    async def test_archive_with_targz(self):
        archive_to = self.tmp_path / "archive.tar.gz"
        self.image_mount.archive_to = archive_to
        self.image_mount.archive_method = "tar.gz"
        container = Container(
            command=("/bin/sh", "-c", "echo hello > /data/file"),
            image=self.image,
            mounts=[self.image_mount],
            network="none",
        )
        async with self.run(container) as proc:
            assert await proc.wait() == 0

        assert archive_to.exists()
        with tarfile.open(archive_to, "r:gz") as tar_file:
            members = {tar_info.name: tar_info for tar_info in tar_file.getmembers()}
            file_name = "./file"
            assert file_name in members
            with tar_file.extractfile(members[file_name]) as fo:
                assert fo.read().decode("utf8").strip() == "hello"

    @pytest.mark.asyncio
    async def test_archive_with_targz_chown(self):
        archive_to = self.tmp_path / "archive.tar.gz"
        self.image_mount.archive_to = archive_to
        self.image_mount.archive_method = "tar.gz"
        self.image_mount.archive_tar_content_owner = "2000:3000"
        container = Container(
            command=("/bin/sh", "-c", "echo hello > /data/file"),
            image="alpine:3.18.2",
            mounts=[self.image_mount],
            network="none",
        )
        async with self.run(container) as proc:
            assert await proc.wait() == 0

        assert archive_to.exists()
        with tarfile.open(archive_to, "r:gz") as tar_file:
            for member in tar_file.getmembers():
                assert member.uid == 2000
                assert member.uname == ""
                assert member.gid == 3000
                assert member.gname == ""

    @pytest.mark.asyncio
    async def test_archive_success(self):
        archive_to = self.tmp_path / "archive"
        archive_success = self.tmp_path / "success"
        self.image_mount.archive_to = archive_to
        self.image_mount.archive_success = archive_success
        container = Container(
            command=("/bin/sh", "-c", "echo hello > /data/file"),
            image=self.image,
            mounts=[self.image_mount],
            network="none",
        )
        async with self.run(container) as proc:
            assert await proc.wait() == 0

        file_path = archive_to / "file"
        assert file_path.exists()
        assert file_path.read_text().strip() == "hello"
        assert archive_success.read_text() == ""

    @pytest.mark.asyncio
    async def test_archive_with_white_out_files(self):
        archive_to = self.tmp_path / "archive"
        archive_success = self.tmp_path / "success"
        self.image_mount.archive_to = archive_to
        self.image_mount.archive_success = archive_success
        self.image_mount.archive_method = "tar.gz"
        container = Container(
            command=(
                "/bin/sh",
                "-c",
                "rm -rf /data/var/empty && rm -rf /data/usr/bin/diff",
            ),
            image=self.image,
            mounts=[self.image_mount],
            network="none",
        )
        async with self.run(container) as proc:
            assert await proc.wait() == 0

        assert archive_success.read_text() == ""
        with tarfile.open(archive_to, "r:gz") as tar_file:
            if USE_ROOTLESS_WHITEOUT:
                assert_has_whiteout_file(
                    path=pathlib.PurePosixPath("./usr/bin/diff"),
                    prefix=".wh.",
                    tar_file=tar_file,
                )
                assert_has_whiteout_file(
                    path=pathlib.PurePosixPath("./var/empty"),
                    prefix=".wh.wh.opq",
                    tar_file=tar_file,
                )
            else:
                assert_has_whiteout_device(
                    path=pathlib.PurePosixPath("./usr/bin/diff"),
                    tar_file=tar_file,
                )
                assert_has_whiteout_device(
                    path=pathlib.PurePosixPath("./var/empty"),
                    tar_file=tar_file,
                )
