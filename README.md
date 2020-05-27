# Elasticsearch Document Resource

Tracks Elasticsearch documents/events.

## Source Configuration

* `addresses`: *Required.* The list of URIs the client should connect to.
Must include protocol (http/s), ip/host, and port.

* `index`: *Required.* The index to track.

* `sort_fields`: *Required.* The ordered fields to sort on when querying for new records.

  Sort fields are used to align with index sort configuration and optimize queries.
  These fields dictate search semantics and therefore concourse resource version ordering and shouldn't be changed once set.
  Prefer using a new index/resource entirely and backfilling with the same IDs in the original index if changes to these fields are required.
  Backfilling with the original IDs can help preserve history as long as the final ordering of the versions hasn't changed.

* `username`: *Optional.* The username to use when authenticating.

* `password`: *Optional.* The password to use when authenticating.

## Behavior

### `check`: Check for new documents.

The latest untracked IDs are fetched from the given index, ordered in descending order according to the `sort_fields`.

### `in`: Fetch the document from the index.

Fetches the document using a tracked ID.

The following files will be placed in the destination:

* `/$VERSION`: The fetched document, named according to its version as reported by concourse which is identical to the ES document ID.

#### Parameters

* `document`: *Optional.* File name of the document.
If not set, will use the document's ID.

### `out`: Upload a document to the index.

Upload a document to the source's index.
The ID is generated if one isn't provided and is used as the version.

If the index already exists, it is used as-is.
No attempt is made to reconcile the index in any way; not the sort order (should be immutable anyway) nor the field mapping.
For maximum tuning and semantic accuracy, the index should be created separately.

If the index doesn't exist, it is created using the `field_map` field in conjunction with the source `sort_fields` field.

#### Parameters

* `document`: *Required.* Path to the document to be uploaded.

* `field_map`: *Optional.* A map of fields to elasticsearch types.

  This is a primitive way to create the index if it doesn't already exist.
  These fields provide type requirements to ES up front rather than allowing ES to infer the types at runtime.
  This is used in conjunction with `sort_fields` to configure the index's field mappings and sort configuration.
  To maximize performance or fields semantic accuracy, one should create the index prior to use here.
  At which point, this `field_map` need not be set.
  
  Note: Validation of types used in this field is performed by the underlying elastic go client.

## Example

```yaml
resource_types:
- name: es
  type: docker-image
  source:
    repository: quay.io/dmarkwat/concourse-elasticsearch
    tag: latest

resources:
- name: my-events
  type: es
  source:
    # elasticsearch deployed on kubernetes using Helm
    addresses: ['http://elasticsearch-master.elasticsearch.svc.cluster.local:9200']
    index: my-events
    sort_fields: ['timestamp']

jobs:
- name: watch-events
  plan:
  - get: my-events
    trigger: true
  - get: debian
  - task: print status
    image: debian
    config:
      platform: linux
      inputs:
      - name: my-events
      run:
        path: /bin/ls
        args: [my-events]
```

## Development

### Prerequisites

* golang is *required* - version 1.14.x is tested; support for modules required.
* docker is *required* - version 19+ is tested; earlier versions may also work.

### Building

Everything is in the Makefile:
```bash
make docker
# or without docker:
# make build
```

### Running the tests

As ES doesn't offer easy ways to unit test with their client, integration tests are largely used throughout.
This isn't ideal, but it was the sanest way to get this tested at the time.
Docker is used to spin up a single-node elasticsearch cluster for the tests to run against.

Same as building, this is contained in the Makefile:
```bash
# docker must be running for these to work
make test
```
Most modern systems shouldn't experience issues, but no startup readiness checking is used at this time, so if ES is slow to start, tests will likely fail.

ES doesn't always obey the `docker start` directive in the test target, so sometimes cleanup needs to happen.
This is OK as every test run should use wholly different indices from previous runs and indices are ideally cleaned up every time, allowing these tests to be hammered frequently and with minimal risk of test result collision.

But that said, when in doubt, `make clean` it out.

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass locally.
