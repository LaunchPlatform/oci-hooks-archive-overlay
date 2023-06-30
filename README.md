# oci-hooks-archive-overlay
An OCI hook for archiving overlay mount upperdir after container is done

# Why

For some OCI tools like [podman](https://podman.io), it allows you to mount a container image in either readonly or read-write mode.
Like this:

```bash
podman run --mount type=image,source=my-data-image,destination=/data,rw=true -it alpine
```

A read-write mount in the OCI spec config may look like this

```json
{
  "destination": "/data",
  "type": "overlay",
  "source": "/home/user/.local/share/containers/storage/overlay-containers/f6e6a7a7eaeb695bb433da1e057d92d9c2e376fb9920792c809d2c1af49e5709/userdata/overlay/3190055391/merge",
  "options": [
    "lowerdir=/home/user/.local/share/containers/storage/overlay/7877ad4aca46f49c306c8044f0d2a1528b642db9aed165f8b022f2b59fc9c237/merged",
    "upperdir=/home/user/.local/share/containers/storage/overlay-containers/f6e6a7a7eaeb695bb433da1e057d92d9c2e376fb9920792c809d2c1af49e5709/userdata/overlay/3190055391/upper",
    "workdir=/home/user/.local/share/containers/storage/overlay-containers/f6e6a7a7eaeb695bb433da1e057d92d9c2e376fb9920792c809d2c1af49e5709/userdata/overlay/3190055391/work",
    "private",
    "userxattr"
  ]
}
```

As you can see it's an [overlayfs](https://docs.kernel.org/filesystems/overlayfs.html) mount.
Then you can modify the content files of the image at the mounted directory in the container.
For many usecases, one could be to capture the changes made to the mounted from the container.
Given the example shown above, the changes made to the `merge` will result in the `upper` directory:

```
/home/user/.local/share/containers/storage/overlay-containers/f6e6a7a7eaeb695bb433da1e057d92d9c2e376fb9920792c809d2c1af49e5709/userdata/overlay/3190055391/upper
```

With the changes to the files in the original mounted image captured, you can then make a layer-based file system with revision history.
You can think about using it like Docker images or OCI container images but with just data changes in each layers.
The job of this hook is to archive the `upperdir` after the container stops.

# How

To use this hook, you need to add annotations to your container to tell it which upperdir of the desired overlay mount point and a given destination path to copy to after the container stops.
There are two types of annotation you can add, one for the mount point and another for the destination of archive.

- com.launchplatform.oci-hooks.archive-overlay.**<ARCHIVE_NAME>**.mount-point
- com.launchplatform.oci-hooks.archive-overlay.**<ARCHIVE_NAME>**.archive-to
- com.launchplatform.oci-hooks.archive-overlay.**<ARCHIVE_NAME>**.success (optional)
- com.launchplatform.oci-hooks.archive-overlay.**<ARCHIVE_NAME>**.method (optional)
- com.launchplatform.oci-hooks.archive-overlay.**<ARCHIVE_NAME>**.tar-content-owner (optional)

The `ARCHIVE_NAME` can be any valid annotation string without a dot in it.
The `mount-point` and `archive-to` annotations with the same archive name need to appear in pairs, otherwise it will be ignored.
For example, to archive the upperdir of mount point at `/data` to `/path/to/my-archive`, you can add annotations like this

- `com.launchplatform.oci-hooks.archive-overlay.data.mount-point=/data`
- `com.launchplatform.oci-hooks.archive-overlay.data.archive-to=/path/to/my-archive`

Please note that the `mount-point` path should be a `destination` field of the mount, i.e, it's in the container namespace.
And the `archive-to` should be a valid path in the runtime namespace.

Here's an example command with podman:

```bash
podman run \
    --annotation=com.launchplatform.oci-hooks.archive-overlay.data.mount-point=/data \
    --annotation=com.launchplatform.oci-hooks.archive-overlay.data.archive-to=/tmp/my-archive \
    --mount type=image,source=my-data-image,destination=/data,rw=true \
    -it alpine
# Change /data folder in the container then exit
ls /tmp/my-archive
```

The `success` is a path to the empty file to be created as an indicator of a successful archive.
The `method` option by default is `copy`, if you want to archive the upperdir as a tar.gz file, you can set it to `tar.gz` instead.
If you want to change the content file owner of the tar file, you can set `tar-content-owner` value, such as `2000` or `2000:3000`.
Please note that only integer uid and gid supported, username won't work.

## Add poststop hook directly in the OCI spec

There are different ways of running a container, if you are generating OCI spec yourself and running OCI runtimes such as [crun](https://github.com/containers/crun) yourself, you can add the `poststop` hook directly into the spec file like this:

```json
{
  "//": "... other OCI spec content ...",
  "hooks": {
    "poststop": [
      {
        "path": "/usr/bin/archive_overlay"
      }
    ]
  }
}
```

For more information about the OCI spec schema, please see the [document here](https://github.com/opencontainers/runtime-spec/blob/48415de180cf7d5168ca53a5aa27b6fcec8e4d81/config.md#posix-platform-hooks).

## Add OCI hook config

Another way to add the OCI hook is to create a OCI hook config file.
Here's an example:

```json
{
  "version": "1.0.0",
  "hook": {
    "path": "/usr/bin/archive_overlay"
  },
  "when": {
    "annotations": {
        "com\\.launchplatform\\.oci-hooks\\.archive-overlay\\.([^.]+)\\.mount-point": "(.+)",
        "com\\.launchplatform\\.oci-hooks\\.archive-overlay\\.([^.]+)\\.archive-to": "(.+)"
    }
  },
  "stages": ["poststop"]
}
```

For more information about the OCI hooks schema, please see the [document here](https://github.com/containers/podman/blob/v3.4.7/pkg/hooks/docs/oci-hooks.5.md).

# Debug

To debug the hook, you can add `--log-level=debug` (or `trace` if you need more details) argument for the `archive_overlay` executable, it will print debug information.
With OCI runtimes like [crun](https://github.com/containers/crun), you can also add an annotation like this:

```
run.oci.hooks.stderr=/path/to/stderr
```

to make the runtime redirect the stderr from the hook executable to specific file.
Please note that podman invokes poststop hook instead of delegating it to crun, so the annotation won't work for podman.

## Syslog

You can also pass `--syslog` option to make the hook omits log messages to syslog.
However, please ensure that you have syslog daemon running on your system otherwise the hook still runs but no log messages will be sent.
