# Introduction

This tool is created for the purpose of debugging issues with [Elastic Filebeat](https://www.elastic.co/beats/filebeat).

The tool is able to analyse the [Filebeat registry logs](https://www.elastic.co/guide/en/beats/filebeat/current/how-filebeat-works.html#_how_does_filebeat_keep_the_state_of_files) (Filestream and Log input) and find suspicious records like:

* a changed [inode](https://en.wikipedia.org/wiki/Inode) for a file with the same filename. This might happen because of certain log rotation strategies or if a target file is on a network share.
* ... (something might come later)

# Installation

Just download a binary from the latest release (according to your OS and CPU architecture).

To build the binary from the repository use `make` in the root of the project. You'll need to have [Go](https://go.dev/learn/) installed on your machine.

# Usage

The Filebeat registry log file can be usually found in `/usr/share/filebeat/data/registry/filebeat/log.json`.

If you have multiple files (for example, from multiple Filebeat containers) you can run the tool with multiple filenames. Then the logs will be concatenated and treated as one.

## Example

```sh
./bin/regan log-*.json
2022/09/29 13:33:51 Given 2 files: log-8.3.3.json, log-8.4.1.json
2022/09/29 13:33:51 Starting analysis with 2 workers...
2022/09/29 13:33:51 Reading from log-8.3.3.json...
2022/09/29 13:33:51 Reading from log-8.4.1.json...
2022/09/29 13:33:51 Reading from log-8.4.1.json finished.
2022/09/29 13:33:51 Reading from log-8.3.3.json finished.
2022/09/29 13:33:51 Found 12116 records in 2 files
2022/09/29 13:33:51 Found 64 unique files in the log
2022/09/29 13:33:51 Analysing...
2022/09/29 13:33:51 File /var/log/containers/7df78df4b5-af1a2c38e09f817c665e58bb1bcc4fa6330ff9496fb83fdb1d56e976162f73f7.log has multiple keys in the registry:
        filestream::filestream-kubernetes.container_logs-af1a2c38e09f817c665e58bb1bcc4fa6330ff9496fb83fdb1d56e976162f73f7::native::247584613-64768
        filestream::filestream-kubernetes.container_logs-af1a2c38e09f817c665e58bb1bcc4fa6330ff9496fb83fdb1d56e976162f73f7::native::247548407-64768
2022/09/29 13:33:51 Analysis is complete, 1 fact(s) reported
```
