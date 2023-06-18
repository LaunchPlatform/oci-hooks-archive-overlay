# oci-hooks-archive-overlay
OCI hooks for archiving overlay mount upperdir after container is done

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

The `ARCHIVE_NAME` can be any valid annotation string without a dot in it.
The `mount-point` and `archive-to` annotations with the same archive name need to appear in pairs, otherwise it will be ignored.
For example, to archive the upperdir of mount point at `/data` to `/path/to/my-archive`, you can add annotations like this

- `com.launchplatform.oci-hooks.archive-overlay.data.mount-point=/data`
- `com.launchplatform.oci-hooks.archive-overlay.data.archive-to=/path/to/my-archive`

Please note that the `mount-point` path should be a `destination` field of the mount, i.e, it's in the container namespace.
And the `archive-to` should be a valid path in the runtime namespace.
