import typing

from containers import ImageMount
from containers import Podman
from containers.providers.helpers import make_annotation_args

from tests.data_types import ImageMountWithArchive

OCI_HOOK_PREFIX = "com.launchplatform.oci-hooks.archive-overlay."


class PodmanWithArchive(Podman):
    def make_overlay_archive_annotations(
        self, image_mount: ImageMountWithArchive, name: typing.Optional[str] = None
    ) -> typing.Tuple[str, ...]:
        if image_mount.archive_to is None:
            return tuple()
        if name is None:
            name = self._make_unique_mount_name()
        args = {
            f"{OCI_HOOK_PREFIX}{name}.mount-point": image_mount.target,
            # TODO: special treatment for windows
            f"{OCI_HOOK_PREFIX}{name}.archive-to": str(
                image_mount.archive_to
            ),
        }
        if image_mount.archive_success is not None:
            args[f"{OCI_HOOK_PREFIX}{name}.success"] = str(
                image_mount.archive_success
            )
        if image_mount.archive_method is not None:
            args[f"{OCI_HOOK_PREFIX}{name}.method"] = str(
                image_mount.archive_method
            )
        if image_mount.archive_tar_content_owner is not None:
            args[
                f"{OCI_HOOK_PREFIX}{name}.tar-content-owner"
            ] = str(image_mount.archive_tar_content_owner)
        return make_annotation_args(args)

    def make_image_mount(self, mount: ImageMount, name: typing.Optional[str] = None):
        args = super().make_image_mount(mount, name)
        if isinstance(mount, ImageMountWithArchive):
            return args + self.make_overlay_archive_annotations(image_mount=mount, name=name)
        return args
