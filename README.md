# Git LFS S3 caching adapter
[![pipeline status](https://gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/badges/main/pipeline.svg)](https://gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/-/commits/main) 
[![Latest Release](https://gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/-/badges/release.svg)](https://gitlab.heliumnet.nl/toolbox/git-lfs-s3-caching-adapter/-/releases) 

The Git LFS S3 caching adapter is an implementation of a [Git LFS custom transfer agent](https://github.com/git-lfs/git-lfs/blob/main/docs/custom-transfers.md) that has as a goal to add S3 bucket caching to existing LFS implementations. This could prove useful if bandwidth to the main LFS storage is expensive (because of metering or fysical distance to the data).

As a custom transfer agent, it has to be set-up for a repository, but could then work completely transparent afterwards.

## Installation
Download the Git LFS S3 caching adapter binary for your operating system and put it on your `PATH`. Then, run `git-lfs-s3-caching-adapter install` to insert the adapter as a custom transfer adapter inside your global `.gitconfig` file. Optionally, you can also install it in the system config file, your worktree or even locally. See `git-lfs-s3-caching-adapter install --help` for more details.

## Configuration
The Git LFS S3 caching adapter will use the Git LFS configuration to determine the upstream LFS API and storage settings. However, it needs to be configured with the credentials of a S3 bucket to actually start caching any objects.

Configuration is read from multiple sources, with preference from high to less preferred:
 - The `.lfscaching.json` configuration file.
 - The repository's `.git/config` Git configuration `lfscache` scope.
 - The users global `.gitconfig` Git configuration `lfscache` scope, e.g. `~/.gitconfig`.
 - The systems global `.gitconfig` Git configuration `lfscache` scope, e.g. `/etc/gitconfig`.

All configuration keys can be set in every config. The following keys are available:
 - `bucket` (`string`): The name of the bucket to store the cached objects in/read the cached objects from
 - `configurationFiles` (`array` of `string`): The paths to the AWS S3 style configuration files to use when configuring the S3 connection. See [this page](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html#cli-configure-files-format) for more information.
   - In Git configuration style, use `configFile`, and provide only a single file.
 - `credentialsFiles` (`array` of `string`): The paths to the AWS S3 style credential files to use when configuring the S3 connection. See [this page](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html#cli-configure-files-format) for more information.
   - In Git configuration style, use `credentialFile`, and provide only a single file.
 - `endpoint` (`string`): The S3 endpoint to connect to when connecting to the bucket.
 - `prefix` (`string`): The prefix to use for every stored object in the bucket/when reading an object from the bucket.
 - `profile` (`string`): The AWS profile to use from the specified configuration/credential files.
 - `region` (`string`): The region in which the bucket resides.
 - `scope`: (`string`): A scope to read global configuration settings from. See [Scopes](#scopes).
 - `usePathStyle` (`boolean`): When `true`, use path style endpoints to connect to the bucket. Useful for custom S3 implementations such as Minio and Ceph Object Gateway.

An example of these keys in a `.lfscaching.json` file:
```
{
    "bucket": "my-lfs-cache-bucket",
    "configurationFiles": [
        "/etc/lfs/caching_config",
    ],
    "credentialsFiles": [
        "/etc/lfs/caching_credentials",
    ],
    "endpoint": "s3.eu-central-1.amazonaws.com",
    "prefix": "my-repo-name",
    "profile": "lfs",
    "region": "eu-central-1",
    "scope": "test",
    "usePathStyle": false
}
```

An example of these keys in any Git configuration file:
```
[lfscache]
    bucket = my-lfs-cache-bucket
    configFile = /etc/lfs/caching_config
    credentialsFile = /etc/lfs/caching_credentials
    endpoint = s3.eu-central-1.amazonaws.com
    prefix = my-repo-name
    profile = lfs
    region = eu-central-1
    usePathStyle = false
```

### Scopes
It might prove difficult to provide a global configuration in your `.gitconfig` that suits every repository you want to use the Git LFS S3 caching adapter in. Therefore, it is also possible to provide every configuration key in a scoped form (except for the `scope` key itself). For example, if my scope is named `test`, I could set:
```
[lfscache "test"]
    bucket = my-lfs-cache-bucket
    configFile = /etc/lfs/caching_config
    credentialsFile = /etc/lfs/caching_credentials
    endpoint = s3.eu-central-1.amazonaws.com
    prefix = my-repo-name
    profile = lfs
    region = eu-central-1
    usePathStyle = false
```
To use these configuration keys for a specific repository, set the `scope` key inside the `.lfscaching.json` file or the `.git/config` file to `test`. Then the configuration will be read from this scope first, with a fallback to the unscoped global configuration.

## Activation
To actually use the Git LFS S3 caching adapter for a repository (or multiple repositories), the `lfs.url` option must be set to `caching::`. To enable it for a repository, this could be set in the `.lfsconfig` file. For example:
```
[lfs]
    url = caching::
```
You can also use different values for different remotes, to enable caching on specific remotes. For more information, read [this Git LFS custom configuration section](https://github.com/git-lfs/git-lfs/blob/main/docs/api/server-discovery.md#custom-configuration).

## Usage
After configuration, you can use `git lfs` and `git` commands as you normally would. Git LFS will invoke the Git LFS S3 caching adapter when needed. The Git LFS S3 caching adapter will then perform the download and upload tasks.
 - When downloading files, the Git LFS S3 caching adapter will check the S3 bucket for the object first and download it from the bucket when available. In that case, the upstream Git LFS storage is not even invoked. If the file is not available in the bucket, it is downloaded from the upstream Git LFS storage first, and then added to the bucket for future downloads.
 - When uploading files, the Git LFS S3 caching adapter will actually perform 2 upload operations per file. First, the file is uploaded to the upstream Git LFS storage. When successful, the file is uploaded a second time to the bucket. This way, the cache for this file can be used immediatly on its first download.

### Statistics
Because the Git LFS S3 caching adapter works as transparently as possible, it might be difficult to measure how much bandwidth is being saved by using it. Therefore, the Git LFS S3 caching adapter keeps statistics on cache usage per repository. This can be requested by navigating to the Git repository and running:
```
git-lfs-s3-caching-adapter stats
```

Statistics files are saved per session of the `git-lfs-s3-caching-adapter` in `.git/lfs/cache_stats`. You might accumulate a lot of little files in that directory. To clean up, you might want to 'compact' the statistics objects:
```
git-lfs-s3-caching-adapter stats compact
```

To see fresh numbers, it is possible to reset the statistics. Old statistics will no be removed, but will no count against the output in `git-lfs-s3-caching-adapter stats`. To reset the statistics, run:
```
git-lfs-s3-caching-adapter stats reset
```
If the user wants to include historic statistics (pre-reset) for a repository, the user can run:
```
git-lfs-s3-caching-adapter stats --total
```

It is also possible to completely remove all statistics files from a repository. To delete all statistics files, run:
```
git-lfs-s3-caching-adapter stats reset --purge
```

More options might be available, please use flag `--help` for more information on a command, for example:
```
git-lfs-s3-caching-adapter stats --help
```

### Debugging
When running any Git of Git LFS commands, prefix the following environment variables to see debugging output:
```
GIT_TRACE=1 GIT_CURL_VERBOSE=1 GIT_TRANSFER_TRACE=1
```
For example:
```
GIT_TRACE=1 GIT_CURL_VERBOSE=1 GIT_TRANSFER_TRACE=1 git lfs pull
```
This way, Git LFS will print tracing output to the terminal when performing the command. Invokes of the `git-lfs-s3-caching-adapter` will be visible in this output.

## Uninstallation
Ensure to remove any relevant configuration flags from your Git configuration or `.lfsconfig` file. Then run `git-lfs-s3-caching-adapter uninstall` to remove the adapter from your global `.gitconfig` file. Optionally, you can also uninstall it in the system config file, your worktree or even locally. See `git-lfs-s3-caching-adapter uninstall --help` for more details.

After uninstalling the configuration, you can simply remove the binary from your `PATH` or your system.
