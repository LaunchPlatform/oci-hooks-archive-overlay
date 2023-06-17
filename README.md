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

Then you can modify the content files of the image at the mounted directory.
For many usecases, one may like to capture the changes made to the mounted from the container.
With the changes to the files in the original mounted image captured, you can then make a layer-based file system with revision history.
You can think about using it like Docker images or OCI container images but with data changes in each layers.
