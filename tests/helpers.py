import pathlib
import tarfile


def assert_has_whiteout_file(
    path: pathlib.PurePosixPath, prefix: str, tar_file: tarfile.TarFile
):
    wh_name = "./" + str(path.with_name(f"{prefix}{path.name}"))
    try:
        tar_file.getmember(wh_name)
    except KeyError:
        assert (
            False
        ), f"Cannot find whiteout name {wh_name} among members {[info.name for info in tar_file.getmembers()]}"


def assert_has_whiteout_device(path: pathlib.PurePosixPath, tar_file: tarfile.TarFile):
    name = "./" + str(path)
    try:
        member = tar_file.getmember(name)
        assert member.type == tarfile.CHRTYPE
        assert member.devminor == 0 and member.devmajor == 0
    except KeyError:
        assert (
            False
        ), f"Cannot find whiteout device {name} among members {[info.name for info in tar_file.getmembers()]}"
