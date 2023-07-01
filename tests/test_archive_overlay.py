import functools
import pathlib
import tarfile
import typing

import pytest
import pytest_asyncio
from containers import Container
from containers import ContainersService
from containers import make_containers_service

from .data_types import ImageMountWithArchive
from .providers import PodmanWithArchive


@pytest.fixture
def log_level() -> typing.Optional[str]:
    return None


@pytest.fixture
def podman() -> PodmanWithArchive:
    return PodmanWithArchive()


@pytest.fixture
def containers(podman: PodmanWithArchive) -> ContainersService:
    return make_containers_service(podman)


@pytest.fixture
def image() -> str:
    return "alpine:3.18.2"


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
            image="alpine:3.18.2",
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
            image="alpine:3.18.2",
            mounts=[image_mount],
            network="none",
        )
        async with self.run(container) as proc:
            assert await proc.wait() == 0

        file_path = archive_to / "file"
        assert file_path.exists()
        assert file_path.read_text().strip() == "hello"
        assert archive_success.read_text() == ""
