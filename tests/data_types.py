import dataclasses
import typing

from containers import ImageMount
from containers.data_types import PathType


@dataclasses.dataclass
class ImageMountWithArchive(ImageMount):
    archive_to: typing.Optional[PathType] = None
    archive_success: typing.Optional[PathType] = None
    archive_method: typing.Optional[str] = None
    archive_tar_content_owner: typing.Optional[str] = None
