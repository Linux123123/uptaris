<?php

namespace Deployer;

require 'recipe/common.php';

set('application', 'uptaris-api');
set('repository', 'git@github.com:Linux123123/uptaris.git');
set('default_timeout', 600);
set('update_code_strategy', 'clone');
set('keep_releases', 5);
set('go_binary', getenv('GO_BINARY') ?: 'go');
set('service_name', 'uptaris-api');

add('shared_files', ['.env']);

$hostname = getenv('DEPLOY_HOST');
if ($hostname === false || trim($hostname) === '') {
    throw new \RuntimeException('Missing required deployment secret: DEPLOY_HOST');
}

host('uptaris-api')
    ->set('hostname', trim($hostname))
    ->set('port', 22)
    ->set('remote_user', 'uptaris')
    ->set('deploy_path', '/home/uptaris/uptaris-api')
    ->set('forward_agent', true)
    ->setLabels(['type' => 'production']);

desc('Check that Go is installed on deployment host');
task('deploy:check_go', function () {
    $go = get('go_binary');
    $version = run("command -v {$go} >/dev/null && {$go} version");
    writeln("<info>{$version}</info>");
});

desc('Download Go dependencies');
task('deploy:go_mod_download', function () {
    $go = get('go_binary');
    within('{{release_path}}/backend', function () use ($go) {
        run("{$go} mod download");
    });
});

desc('Build API and migration binaries');
task('deploy:go_build', function () {
    $go = get('go_binary');
    within('{{release_path}}/backend', function () use ($go) {
        run('mkdir -p ../bin');
        run("CGO_ENABLED=0 {$go} build -trimpath -ldflags='-s -w' -o ../bin/uptaris-api ./cmd/api");
        run("CGO_ENABLED=0 {$go} build -trimpath -ldflags='-s -w' -o ../bin/uptaris-migrate ./cmd/migrate");
        run('chmod +x ../bin/uptaris-api ../bin/uptaris-migrate');
    });
});

desc('Apply database migrations');
task('deploy:migrate', function () {
    within('{{release_path}}/backend', function () {
        run('../bin/uptaris-migrate');
    });
});

desc('Restart Uptaris API service');
task('deploy:restart_service', function () {
    $service = get('service_name');
    run("systemctl --user restart {$service}.service");
});

task('deploy', [
    'deploy:prepare',
    'deploy:check_go',
    'deploy:go_mod_download',
    'deploy:go_build',
    'deploy:migrate',
    'deploy:publish',
    'deploy:restart_service',
]);

after('deploy:failed', 'deploy:unlock');
after('rollback', 'deploy:restart_service');
