import pathlib
import tarfile

import pytest
from containers import Container
from containers import ContainersService
from containers import make_containers_service

from .data_types import ImageMountWithArchive
from .providers import PodmanWithArchive


@pytest.fixture
def podman() -> PodmanWithArchive:
    return PodmanWithArchive()


@pytest.fixture
def containers(podman: PodmanWithArchive) -> ContainersService:
    return make_containers_service(podman)


@pytest.mark.asyncio
async def test_archive_with_copy(tmp_path: pathlib.Path, containers: ContainersService):
    data_image = "alpine:3.18.2"
    await containers.load_image(data_image)

    archive_to = tmp_path / "archive"

    image_mount = ImageMountWithArchive(
        source=data_image,
        target="/data",
        read_write=True,
        archive_to=archive_to,
    )
    container = Container(
        command=("/bin/sh", "-c", "echo hello > /data/file"),
        image="alpine:3.18.2",
        mounts=[image_mount],
        network="none",
    )
    async with containers.run(container, log_level="debug") as proc:
        assert await proc.wait() == 0

    file_path = archive_to / "file"
    assert file_path.exists()
    assert file_path.read_text().strip() == "hello"


@pytest.mark.asyncio
async def test_archive_with_targz(
    tmp_path: pathlib.Path, containers: ContainersService
):
    data_image = "alpine:3.18.2"
    await containers.load_image(data_image)

    archive_to = tmp_path / "archive.tar.gz"

    image_mount = ImageMountWithArchive(
        source=data_image,
        target="/data",
        read_write=True,
        archive_to=archive_to,
        archive_method="tar.gz",
    )
    container = Container(
        command=("/bin/sh", "-c", "echo hello > /data/file"),
        image="alpine:3.18.2",
        mounts=[image_mount],
        network="none",
    )
    async with containers.run(container, log_level="debug") as proc:
        assert await proc.wait() == 0

    assert archive_to.exists()
    with tarfile.open(archive_to, "r:gz") as tar_file:
        members = {tar_info.name: tar_info for tar_info in tar_file.getmembers()}
        file_name = "./file"
        assert file_name in members
        with tar_file.extractfile(members[file_name]) as fo:
            assert fo.read().decode("utf8").strip() == "hello"


@pytest.mark.asyncio
async def test_archive_with_targz_chown(
    tmp_path: pathlib.Path, containers: ContainersService
):
    data_image = "alpine:3.18.2"
    await containers.load_image(data_image)

    archive_to = tmp_path / "archive.tar.gz"

    image_mount = ImageMountWithArchive(
        source=data_image,
        target="/data",
        read_write=True,
        archive_to=archive_to,
        archive_method="tar.gz",
        archive_tar_content_owner="2000:3000",
    )
    container = Container(
        command=("/bin/sh", "-c", "echo hello > /data/file"),
        image="alpine:3.18.2",
        mounts=[image_mount],
        network="none",
    )
    async with containers.run(container, log_level="debug") as proc:
        assert await proc.wait() == 0

    assert archive_to.exists()
    with tarfile.open(archive_to, "r:gz") as tar_file:
        members = {tar_info.name: tar_info for tar_info in tar_file.getmembers()}
        for member in members:
            assert member.uid == 2000
            assert member.uname == ""
            assert member.gid == 3000
            assert member.gname == ""


@pytest.mark.asyncio
async def test_archive_success(tmp_path: pathlib.Path, containers: ContainersService):
    data_image = "alpine:3.18.2"
    await containers.load_image(data_image)

    archive_to = tmp_path / "archive"
    archive_success = tmp_path / "success"

    image_mount = ImageMountWithArchive(
        source=data_image,
        target="/data",
        read_write=True,
        archive_to=archive_to,
        archive_success=archive_success,
    )
    container = Container(
        command=("/bin/sh", "-c", "echo hello > /data/file"),
        image="alpine:3.18.2",
        mounts=[image_mount],
        network="none",
    )
    async with containers.run(container, log_level="debug") as proc:
        assert await proc.wait() == 0

    file_path = archive_to / "file"
    assert file_path.exists()
    assert file_path.read_text().strip() == "hello"
    assert archive_success.read_text() == ""
