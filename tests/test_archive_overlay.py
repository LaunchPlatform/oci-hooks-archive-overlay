import pathlib

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
    )
    async with containers.run(container, log_level="debug") as proc:
        assert await proc.wait() == 0

    file_path = archive_to / "file"
    assert file_path.exists()
    assert file_path.read_text().strip() == "hello"
