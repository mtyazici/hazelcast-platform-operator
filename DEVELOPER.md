# Developing and testing Hazelcast Platform Operator

In this document, you will find the required information/steps to ease your contribution.

## Deploying the operator to Kubernetes

To deploy the operator from the source, you can run the command below:

```shell
make deploy IMG=hazelcast/hazelcast-platform-operator:latest
```

> Note: The operator is installed into the `default` namespace by default. You can pass `NAMESPACE` to `make` commands if you want to change it.

### Cleanup

To remove the operator from the Kubernetes cluster, run the following command:

```shell
make undeploy
```

## Running the operator locally

Hazelcast Platform Operator uses [hazelcast go-client](https://github.com/hazelcast/hazelcast-go-client) to connect to the cluster. For this reason, the pods need to be
exposed outside the cluster. Run the `make expose-local` command to expose Hazelcast member to `localhost:8000`.

The operator run must be built, with `build constraint` tag `localrun`:

```shell
go build -o bin/manager -tags localrun main.go
```

Or using `make` that will include the tag by default:

```shell
make install run
```

> Note: You can override the build tags in the `make` commands by setting `GO_BUILD_TAGS` env variable.

### Setting up build tags in GoLand

To run the operator from `GoLand`, execute the following steps to add build tags:

1. In GoLand Preferences, navigate to `Go | Build tags & Vendoring`
2. In the `Custom tags` field, enter `localrun`
3. Go to the `Run configuration` of the `Go build` select the `Use all custom build tags` checkbox

Now you can run the `main.go` using `GoLand`.

## Running Tests

There are different types of tests related to Hazelcast Platform Operator. 

### Running unit & integration tests

To run unit & integration tests, execute the following command.

```shell
make test
```

You can also run unit & integration tests separately by `make test-unit` and `make test-it` commands.

### Running end-to-end tests

You can run end-to-end tests by [deploying the operator to Kubernetes cluster](#deploying-the-operator-to-kubernetes).

```shell
make test-e2e NAMESPACE=<YOUR NAMESPACE>
```