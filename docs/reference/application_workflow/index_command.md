# Index Command

The `index` command is the main command that starts the indexer workflow. It is built using the [cobra](https://cobra.dev/) framework.

The command can be configured using CLI flags or a passed in `.toml` configuration file. See the [Configuration](../usage/configuration.md) documentation for more information on how to configure the indexer.

The command has the following workflow, implemented through the `root` parent command and the `index` child command:

1. The program is started with the first command being `index`
2. Cobra initialization functions run first
   1. The configuration file parser function is called by the cobra `OnInitialize` function
   2. The `index` command's PreRunE function is called, which calls the `setupIndex` function in the `cmd/index.go` file
      1. This function is responsible for loading configuration values, validating the configuration values and database initializing connections
3. The `index` command's Run function is called, which calls the `index` function in the `cmd/index.go` file
   1. This function is responsible for starting the indexer workflow, see the [Indexer Workflow](application_workflow.md) documentation for more information
