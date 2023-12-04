# starton

### What is it?

Barely anything at all... Just a tiny node script that expects
git repos to conform to a convention. It does almost nothing but
shell out to Github to clone a repo, and try to run scripts it
expects to be in that repo.

### Why do this?

`¯\_(ツ)\_/¯` Partially to use `commander`... partially because every job I've
ever been at has had an issue where every project is a little different to
get started and READMEs are sub-par or poorly maintained... I've had pretty
good success with sticking steps to get projects up and running in bash scripts.
They're easy to use without knowing much about the project and serve
as documentation for how setting up a project works.

## Installation

`npm install -g starton`

## Usage

starton makes assumptions about your project.

1. That it's hosted on Github
2. It has the following scripts in the root directory:
  1. _configure_ - this should do things like copy configuration file templates
    replacing dummy values with values that will make the thing run
  2. _build_ - this should do things like install dependencies and anything else
    required to run the project
  3. _run_ - should run the project
  4. _test_ - should run tests

If any of those scripts don't make sense for your project just leave an empty
file by that name as a no-op.

If you have those things taken care of you can use starton as follows:

### Starting a brand new project

Command: _start_

This command will:
1. Clone down the github project into your current working directory
2. Run the `configure` script
3. Run the `build` script

Example:

```bash
starton start niftyn8/starton
```

### Configure a project

Command: _configure_

This should get run in the root directory of your project.

Example:

```bash
starton configure
```

### Build a project

Command: _build_

This should get run in the root directory of your project.

Example:

```bash
starton build
```

### Run a project

Command: _run_

This should get run in the root directory of your project.

Example:

```bash
starton run
```

### Test a project

Command _test_

This should get run in the root directory of your project.

```bash
starton test
```

### Get help

You can always run `starton --help` to get some helpful output.

## Contributing

1. Fork it ( https://github.com/niftyn8/starton/fork )
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
4. Push to the branch (git push origin my-new-feature)
5. Create a new Pull Request