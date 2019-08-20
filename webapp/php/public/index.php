<?php

require __DIR__ . '/../vendor/autoload.php';

// for the PHP build-in http server to serve static file
if (PHP_SAPI == 'cli-server') {
    if ($_SERVER['SCRIPT_NAME'] !== '/index.php' ) {
        $_SERVER['REQUEST_URI'] = '/index.php' . $_SERVER['SCRIPT_NAME'];
        $_SERVER['SCRIPT_NAME'] = '/index.php';
    }
}

// http://www.slimframework.com/docs/v3/handlers/php-error.html
error_reporting(E_ALL);
set_error_handler(function ($severity, $message, $file, $line) {
    if (error_reporting() & $severity) {
        throw new \ErrorException($message, 0, $severity, $file, $line);
    }
});

// Instantiate the app
$settings = require __DIR__ . '/../src/settings.php';
$app = new \Slim\App($settings);

// Set up dependencies
$dependencies = require __DIR__ . '/../src/dependencies.php';
$dependencies($app);

// Register middleware
$middleware = require __DIR__ . '/../src/middleware.php';
$middleware($app);

// Register routes
$routes = require __DIR__ . '/../src/routes.php';
$routes($app);

// Run app
$app->run();
