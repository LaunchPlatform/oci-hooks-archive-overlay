import asyncio.subprocess
import pathlib

import pytest

from containers import BindMount
from containers import Container
from containers import ContainersService
from containers import ImageMount
from containers import make_containers_service
from .providers import PodmanWithArchive
from .data_types import ImageMountWithArchive


@pytest.fixture
def podman() -> PodmanWithArchive:
    return make_containers_service()

@pytest.fixture
def containers(podman: PodmanWithArchive) -> ContainersService:
    return make_containers_service(podman)


@pytest.mark.asyncio
async def test_archive_with_copy(tmp_path: pathlib.Path, containers: ContainersService):
    data_image = "alpine:3.18.2"
    await containers.load_image(data_image)

    archive_to = tmp_path / "archive"

    image_mount = ImageMount(source=data_image, target="/data")
    container = Container(
        command=("ls", "/data"), image="alpine:3.18.2", mounts=[image_mount]
    )
    async with containers.run(container, stdout=asyncio.subprocess.PIPE) as proc:
        stdout = await proc.stdout.read()
        folders = frozenset(
            filter(lambda item: bool(item.strip()), stdout.decode("utf8").split("\n"))
        )
        assert folders == frozenset(
            {
                "bin",
                "dev",
                "etc",
                "home",
                "lib",
                "media",
                "mnt",
                "opt",
                "proc",
                "root",
                "run",
                "sbin",
                "srv",
                "sys",
                "tmp",
                "usr",
                "var",
            }
        )
        assert await proc.wait() == 0


@pytest.mark.asyncio
async def test_run_with_bind_mount(
        tmp_path: pathlib.Path, containers: ContainersService
):
    data_image = "alpine:3.18.2"
    await containers.load_image(data_image)

    mount_dir = tmp_path / "data"
    mount_dir.mkdir()
    file0 = mount_dir / "my-file0.txt"
    file0.write_text("hello")
    file1 = mount_dir / "nested" / "my-file1.txt"
    file1.parent.mkdir(parents=True, exist_ok=True)
    file1.write_text("there")

    bind_mount = BindMount(source=mount_dir, target="/data", readonly=False)
    container = Container(
        command=("/bin/sh", "-c", "touch /data/new && ls /data"),
        image="alpine:3.18.2",
        mounts=[bind_mount],
    )
    async with containers.run(container, stdout=asyncio.subprocess.PIPE) as proc:
        stdout = await proc.stdout.read()
        folders = frozenset(
            filter(lambda item: bool(item.strip()), stdout.decode("utf8").split("\n"))
        )
        assert folders == frozenset(
            {
                "my-file0.txt",
                "nested",
                "new",
            }
        )
        assert await proc.wait() == 0
