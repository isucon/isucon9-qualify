<?php

use App\Environment;

return [
    'settings' => [
        'displayErrorDetails' => true, // set to false in production
        'addContentLengthHeader' => false, // Allow the web server to send the content-length header
        'determineRouteBeforeAppMiddleware' => true,

        // Renderer settings
        'renderer' => [
            'template_path' => __DIR__ . '/../../public/',
        ],

        // Monolog settings
        'logger' => [
            'name' => 'isucari',
//            'path' => __DIR__ . '/../logs/app.log',
            'path' => 'php://stdout',
            'level' => \Monolog\Logger::INFO,
        ],

        // Database settings
        'database' => [
            'host' => Environment::get('MYSQL_HOST', '127.0.0.1'),
            'port' => Environment::get('MYSQL_PORT', '3306'),
            'username' => Environment::get('MYSQL_USER', 'isucari'),
            'password' => Environment::get('MYSQL_PASS', 'isucari'),
            'dbname' => Environment::get('MYSQL_DBNAME', 'isucari'),
        ],
        'app' => [
            'base_dir' => __DIR__ . '/../',
            'upload_path' => __DIR__ . '/../../public/upload/',
        ],
    ],
];
