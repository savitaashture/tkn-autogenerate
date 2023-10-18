# tkn autogenerate - Generate PipelineRun for [Pipelines-as-Code](https://pipelinesascode.com) automagically 🪄

## Description

`tkn-autogenerate` will inspect a repository and try to guess which tasks to
add and generate a pipelinerun suitable for
[Pipelines-as-Code](https://pipelinesascode.com).

It uses GitHub API to get the languages associated to a repository and some
files heuristics pattern for other rules detections.

This tools should be suitable to an automated system or to be plugged with soon
to be released Pipelines-as-Code [pluggable tekton directory resolver](https://docs.google.com/document/d/1_PfB-OyODXniQXdJ64E-XMiFge3ogPhE6T4cU8MZztA/edit)
to get a fully automated system.

## Installation

```shell
go install github.com/chmouel/tkn-autogenerate@latest
```

## Usage

```shell
tkn-autogenerate org/repositoryname
```

This will query GitHub for the programming language on the REPOSITORY belong to
ORG and automatically generate a`PipelineRun` with the tasks added according to
the detected programming language.

You can specify a GitHub token with the flag `--token` (or `GITHUB_TOKEN`
environment variable) for private repos or don't get rate limited.

## Customization

### Detect Language to Tekton hub mapping

The file [tknautogenerate.yaml](./pkg/tknautogenerate/templates/tknautogenerate.yaml) specify the mapping
between the detected programming language and the task we want to apply into
it.

For example:

```yaml
python:
  tasks:
    - name: pylint
```

Will add the task `pylint` to the `PipelineRun` if one of the detected
programming language is `python`. It will add the [Pipelines as Code remote task
annotation](https://pipelinesascode.com/docs/guide/resolver/#tekton-hubhttpshubtektondev)
to have the Pipeline as Code added to the PipelineRun.

You can add a `name` parameter to the task to customize the name of the task
instead of using the detected language.

```yaml
python:
  name: cobra
  tasks:
    - name: pylint
```

### Passing Parameters

You can add `params` to the task to add parameters to be passed to the task

```yaml
python:
  tasks:
    - name: pylint
      params:
        - name: path
          value: ./package
```

### Passing Workspace

A workspace is automatically to the task unless you don't want it added and
then you can add the `workspace.disabled = true` to the task

```yaml
python:
  tasks:
    - name: pylint
      params:
        - name: path
          value: ./package
          workspace:
            disabled: true
```

If the task is expecting another name than source (the default name we use) you
can specify a name for this:

```yaml
python:
  tasks:
    - name: pylint
      params:
        - name: path
          value: ./package
          workspace:
            name: repo
```

### Task dependencies

You can add a optional `runAfter` parameter to the task to chain dependencies
between tasks which will be passed to the generated PipelineRun.

```yaml
python:
  tasks:
    - name: pylint
      runAfter: [fetch-repository]
```

### Add task matching using patterns to match file repositories

You can also add tasks according to file patterns, for example:

```yaml
containerbuild:
  pattern: "(Docker|Container)file$"
  tasks:
    - name: buildah
      workspace: true
      params:
        - name: IMAGE
          value: "image-registry.openshift-image-registry.svc:5000/$(context.pipelineRun.namespace)/$(context.pipelineRun.name)"
```

If you have a configuration containing the `pattern` keyword it will try to
match it against all the files in the repository, it will be queried using the API
on the `default_branch` of the repository unless you pass the flag
`--target-ref` to another reference or SHA.

### PipelineRun default template

The file [pipelinerun.yaml.go.tmpl](./pkg/tknautogenerate/templates/pipelinerun.yaml.go.tmpl) is the actual
PipelineRun which can be customized according to the [go templating
system](https://pkg.go.dev/text/template).

### Using a specific PipelineRun for a detected language or file pattern

You can specify a specific PipelineRun for a detected language or file pattern, for example:

```yaml
java:
    pipelinerun: "java"
```

will use the file in
[./pkg/tknautogenerate/templates/languages/java.yaml.go.tmpl](./pkg/tknautogenerate/templates/languages/java.yaml.go.tmpl)
to generate the pipelinerun.

If multiple language have multiple pipelinerun then the first one will win.

You can combine this with file pattern too:

```yaml
java:
    pattern: "^pom.xml$"
    pipelinerun: "java"
```

and you can add extra tasks to add if you want to (altho it probably better to
have this in your specific template) directly:

```yaml
java:
    pattern: "^pom.xml$"
    pipelinerun: "java"
    tasks:
    - name: sbom
```

no other language detection will be passed to the templates, it's really a
static pipelinerun generated out of file pattern or programming language
detection.

(TODO: make the language that has most amount of percentage in a repo wins)

## Copyright

[Apache-2.0](./LICENSE)

## Authors

### Chmouel Boudjnah

- Fediverse - <[@chmouel@chmouel.com](https://fosstodon.org/@chmouel)>
- Twitter - <[@chmouel](https://twitter.com/chmouel)>
- Blog - <[https://blog.chmouel.com](https://blog.chmouel.com)>
