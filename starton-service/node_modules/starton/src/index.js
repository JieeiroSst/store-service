#!/usr/bin/env node

var packageInfo = require('../package.json');
var program = require('commander');
var shell = require('shelljs');

function gitClone(gitUri, targetPath) {
    shell.exec('git clone ' + gitUri + ' ' + targetPath);
};

function configure() { shell.exec('./configure') };

function build() { shell.exec('./build'); };

function test() { shell.exec('./test'); };

function run() { shell.exec('./run'); };

function startGithub(githubPath) {
    var projectName = githubPath.split('/')[1];

    if (!shell.ls(projectName).length) {
        gitClone('https://github.com/' + githubPath, projectName);
    }
    shell.cd(projectName);
    configure();
    build();
};

program
    .version(packageInfo.version)

program
    .command('start <githubProject>')
    .description('clones down a project and sets it up')
    .action(startGithub);

program
    .command('configure')
    .description('configures the project')
    .action(configure);

program
    .command('build')
    .description('sets up the project')
    .action(build);

program
    .command('test')
    .description('runs tests for the project')
    .action(test);

program
    .command('run')
    .description('runs the software')
    .action(run);

program.parse(process.argv);