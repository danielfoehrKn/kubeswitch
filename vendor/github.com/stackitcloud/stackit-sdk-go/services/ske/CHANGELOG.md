## 1.9.0
- Package `v2api`:
  - **Feature:** New model structs: `Access`, `IDPKubeconfig`
  - **Feature:** New field `Access` in `Cluster` and `CreateOrUpdateClusterPayload` model structs
  - **Feature:** New API client method: `GetIDPKubeconfig`
  - Added new possible values to description of `Code` field in the `ClusterError` model struct
- Deprecated SDK layer in root of the module:
  - **Feature:** New model structs: `Access`, `IDPKubeconfig`
  - **Feature:** New field `Access` in `Cluster` and `CreateOrUpdateClusterPayload` model structs
  - **Feature:** New API client method: `GetIDPKubeconfig`
  - Added new possible values to description of `Code` field in the `ClusterError` model struct

## v1.8.0
- **Feature:** Introduction of multi API version support for the ske SDK module. For more details please see the announcement on GitHub: https://github.com/stackitcloud/stackit-sdk-go/discussions/5062
- `v1api`: New package which can be used for communication with the ske v1 API
- `v2api`: New package which can be used for communication with the ske v2 API
- **Deprecation:** The contents in the root of this SDK module including the `wait` package are marked as deprecated and will be removed after 2026-09-30. Switch to the new packages for the available API versions instead.
- **Dependencies:** Bump STACKIT SDK core module from `v0.21.1` to `v0.22.0`

## v1.7.0
- **Feature:** new model `AccessScope`
- **Feature:** new model `V2ControlPlaneNetwork`
- **Feature:** added field `ControlPlane` of type `V2ControlPlaneNetwork` to model `Network`

## v1.6.3
- Bump STACKIT SDK core module from `v0.21.0` to `v0.21.1`

## v1.6.2
- **Dependencies**: Bump `github.com/golang-jwt/jwt/v5` from `v5.3.0` to `v5.3.1`

## v1.6.1
- **Bugfix:** Correctly handle file closing for file uploads
- Bump STACKIT SDK core module from `v0.20.1` to `v0.21.0`

## v1.6.0
- **Feature:** Add field `Identity` to model `ClusterStatus`

## v1.5.1
- Bump STACKIT SDK core module from `v0.20.0` to `v0.20.1`

## v1.5.0
- **Feature:** Add `versionState` field to ListProviderOptionsRequest struct
- **Feature:** Add new enum `GetProviderOptionsRequestVersionState`

## v1.4.1
- Bump STACKIT SDK core module from `v0.19.0` to `v0.20.0`

## v1.4.0
- **Feature:** Add new field `Kubernetes` to `Nodepool` model

## v1.3.0
- **Feature:** Add new wait handlers: `TriggerClusterHibernationWaitHandler`, `TriggerClusterMaintenanceWaitHandler`, `TriggerClusterReconciliationWaitHandler`, `TriggerClusterWakeupWaitHandler`

## v1.2.0
- **Feature:** Add new method `TriggerWakeup`

## v1.1.1
- **Dependencies:** Bump `github.com/golang-jwt/jwt/v5` from `v5.2.2` to `v5.2.3`

## v1.1.0
- **Breaking Change:** `ClusterError.Code` field is now a string type instead of an enum. The field no longer validates against predefined enum values and accepts any string value.

## v1.0.0
- **Breaking Change:** The region is no longer specified within the client configuration. Instead, the region must be passed as a parameter to any region-specific request.

## v0.27.0
- **Feature:** Add new `ClusterErrorCode` types: `CLUSTERERRORCODE_INFRA_SNA_NETWORK_NOT_FOUND`, `CLUSTERERRORCODE_FETCHING_ERRORS_NOT_POSSIBLE`

## v0.26.0
- Add `required:"true"` tags to model structs

## v0.25.0 (2025-06-10)
- **Feature:** Add new field `PodAddressRanges` to `ClusterStatus`

## v0.24.1 (2025-06-04)
- **Bugfix:** Adjusted `UnmarshalJSON` function to use enum types and added tests for enums

## v0.24.0 (2025-05-15)
- **Breaking change:** Introduce interfaces for `APIClient` and the request structs

## v0.23.0 (2025-05-14)
- **Breaking change:** Introduce typed enum constants for status attributes

## v0.22.4 (2025-05-13)
- **Feature:** Added `ClusterError`

## v0.22.3 (2025-05-09)
- **Feature:** Update user-agent header

## v0.22.2 (2025-04-29)
- **Bugfix:** Correctly handle empty payload in body

## v0.22.1 (2025-03-19)
- **Internal:** Backwards compatible change to generated code

## v0.22.0 (2025-02-21)
- **New:** Minimal go version is now Go 1.21

## v0.21.1 (2025-01-23)
- Use current SKE API

## v0.21.0 (2025-01-21)

- **Removal:** The following methods were removed after deprecation (2024-04-16) and [`serviceenablement` SDK](https://github.com/stackitcloud/stackit-sdk-go/tree/main/services/serviceenablement) must be used instead.
  - `DisableService`
  - `EnableService`
  - `GetServiceStatus`

## v0.20.1 (2024-12-17)

- **Bugfix:** Correctly handle nullable attributes in model types

## v0.20.0 (2024-10-14)

- **Feature:** Add support for nullable models

## v0.19.0 (2024-07-18)

- **Feature:** New fields for `Extension`: `Dns`

## v0.18.0 (2024-07-10)

- **Bugfix:** Fix marking of deprecated methods. Potential breaking change for users with linters that treat deprecations as errors.

## v0.17.0 (2024-07-04)

- **Feature:** Add new field `AllowSystemComponents` to the `Nodepool` model that configures whether system components are allowed to run on the node pool.

## v0.16.0 (2024-05-27)

- **Breaking change:** Renamed data types:
  - `V1Network` is now `Network`
  - `V1LoginKubeConfig` is now `LoginKubeConfig`

## v0.15.0 (2024-05-14)

- **Feature:** New operation `GetLoginKubeconfig` to get a Kubeconfig for use with the STACKIT CLI. A Kubeconfig retrieved using this endpoint does not contain any credentials and instead obtains valid credentials via the STACKIT CLI.

## v0.14.0 (2024-05-03)

- **Feature:** New fields for `MachineType`: `Architecture`, `Gpu`

## v0.13.0 (2024-04-16)

- **Deprecation:** The following methods have been deprecated and the [Service Enablement API](https://docs.api.stackit.cloud/documentation/service-enablement/version/v1) must be used instead.
  - `DisableService`
  - `EnableService`
  - `GetServiceStatus`

## v0.12.0 (2024-04-09)

- Set config.ContextHTTPRequest in Execute method
- Support WithMiddleware configuration option in the client
- Update `core` to [`v0.12.0`](../../core/CHANGELOG.md#v0120-2024-04-11)

## v0.11.1 (2024-04-09)

- Remove unused model data types.

## v0.11.0 (2024-03-28)

- **Feature**: Waiters for async operation `StartCredentialsRotationWaitHandler` and `CompleteCredentialsRotationWaitHandler`

## v0.10.1 (2024-02-28)

- Update `core` to [`v0.10.0`](../../core/CHANGELOG.md#v0100-2024-02-27)

## v0.10.0 (2024-02-06)

- **Feature:** New endpoints for credentials rotation.
  - `StartCredentialsRotation`
  - `CompleteCredentialsRotation`
  - `CreateKubeconfig`
  - These endpoints replace `GetCredentials` and `TriggerRotateCredentials`, which are **deprecated** and will not work for clusters with Kubernetes v1.27+, or if the new endpoints for kubeconfig or credentials rotation have already been used. For more information, see [How to rotate SKE credentials](https://docs.stackit.cloud/products/runtime/kubernetes-engine/how-tos/rotate-ske-credentials/).

## v0.9.3 (2024-02-02)

- **Improvement**: Fix state name in `CredentialsRotationState.Phase`
- Update `core` to `v0.7.7`. The `http.request` context is now passed in the client `Do` call.
- Increase timeout on `DeleteClusterWaitHandler` to 45 minutes

## v0.9.2 (2024-01-24)

- **Bug fix**: `NewAPIClient` now initializes a new client instead of using `http.DefaultClient` ([#236](https://github.com/stackitcloud/stackit-sdk-go/issues/236))

## v0.9.1 (2024-01-15)

- Add license and notice files

## v0.9.0 (2024-01-09)

- **Feature:** `ClusterStatus` now has a field `CredentialsRotation` with credentials' details
- **Improvement:** Add details on credentials for old clusters
- Dependency updates

## v0.8.1 (2023-12-22)

- Dependency updates

## v0.8.0 (2023-12-20)

API methods, structs and waiters were renamed to have the same look and feel across all services and according to user feedback.

- Changed methods:
  - `CreateProject` renamed to `EnableService`
  - `DeleteProject` renamed to `DisableService`
  - `GetClusters` renamed to `ListClusters`
  - `GetOptions` renamed to `ListProviderOptions`
  - `GetProject` renamed to `GetServiceStatus`
- Changed structs:
  - `ClusterResponse` renamed to `Cluster`
  - `ClustersResponse` renamed to `ListClustersResponse`
  - `CredentialsResponse` renamed to `Credentials`

## v0.7.0 (2023-12-05)

- Manage your STACKIT Kubernetes Engine resources: `Project`, `Cluster`, `Credentials`, `Options`
- Waiters for async operations: `CreateOrUpdateClusterWaitHandler`, `DeleteClusterWaitHandler`, `CreateProjectWaitHandler`, `DeleteProjectWaitHandler`, `RotateCredentialsWaitHandler`
- [Usage example](https://github.com/stackitcloud/stackit-sdk-go/tree/main/examples/ske)
