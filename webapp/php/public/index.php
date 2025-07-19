<?php

use DI\Container;
use Slim\Factory\AppFactory;

require __DIR__ . '/../vendor/autoload.php';

// for the PHP build-in http server to serve static file
if (PHP_SAPI == 'cli-server') {
    if ($_SERVER['SCRIPT_NAME'] !== '/index.php' ) {
        $_SERVER['REQUEST_URI'] = '/index.php' . $_SERVER['SCRIPT_NAME'];
        $_SERVER['SCRIPT_NAME'] = '/index.php';
    }
}

error_reporting(E_ALL);
set_error_handler(function ($severity, $message, $file, $line) {
    if (error_reporting() & $severity) {
        throw new \ErrorException($message, 0, $severity, $file, $line);
    }
});

// Load settings
$settings = require __DIR__ . '/../src/settings.php';

// Create Container
$container = new Container();

// Set up dependencies
$dependencies = require __DIR__ . '/../src/dependencies.php';
$dependencies($container, $settings['settings']);

// Create App
AppFactory::setContainer($container);
$app = AppFactory::create();

// Add error middleware
$app->addErrorMiddleware(true, true, true);

// Register middleware
$middleware = require __DIR__ . '/../src/middleware.php';
$middleware($app);

// Register routes
$routes = require __DIR__ . '/../src/routes.php';
$routes($app);

// Run app
$app->run();
