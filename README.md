# TKN autogenerate

A simple POC showing how to autogenerate PipelineRun automatically for the programming language of a repository with [Pipelines-as-Code](https://pipelinesascode.com/).

## Installation

```shell
go install github.com/chmouel/tkn-autogenerate@latest
```

## Usage

```shell
tkn-autogenerate org/repositoryname
```

This will query Github for the programming language on the REPOSITORY belong to
ORG and automatically generate a`PipelineRun` with the tasks added according to
the detected programming language.

## Customization

### Detect Language to Tekton hub mapping 

The file <./tknautogenerate.yaml> specify the mapping between the detected programming language and the task we want to apply into it.

For example:

```yaml
python:
  tasks:
    - name: pylint
      workspace: true
```

Will add the task `pylint` to the `PipelineRun` if one of the detected
programming language is `python`. It will add the [Pipelines as Code remote task
annotation](https://pipelinesascode.com/docs/guide/resolver/#tekton-hubhttpshubtektondev)
to have the Pipeline as Code added to the PipelineRun.

### PipelineRun template

The file [pipelinerun.yaml.go.tmpl](./pipelinerun.yaml.go.tmpl) is the actual
PipelineRun which can be customized according to the [go templating
system](https://pkg.go.dev/text/template).

## Copyright

[Apache-2.0](./LICENSE)

## Authors

### Chmouel Boudjnah

- Fediverse - <[@chmouel@chmouel.com](https://fosstodon.org/@chmouel)>
- Twitter - <[@chmouel](https://twitter.com/chmouel)>
- Blog  - <[https://blog.chmouel.com](https://blog.chmouel.com)>
