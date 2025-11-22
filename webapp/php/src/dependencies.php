<?php

use DI\Container;
use Psr\Container\ContainerInterface;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Monolog\Handler\StreamHandler;
use Monolog\Logger;
use Monolog\Processor\UidProcessor;
use Slim\Views\PhpRenderer;
use SlimSession\Helper as SessionHelper;

return function (Container $container, array $settings) {
    // Settings
    $container->set('settings', $settings);

    // View renderer
    $container->set('renderer', function (ContainerInterface $c) {
        $settings = $c->get('settings')['renderer'];
        return new PhpRenderer($settings['template_path']);
    });

    // Monolog
    $container->set('logger', function (ContainerInterface $c) {
        $settings = $c->get('settings')['logger'];
        $logger = new Logger($settings['name']);
        $logger->pushProcessor(new UidProcessor());
        $logger->pushHandler(new StreamHandler($settings['path'], $settings['level']));
        return $logger;
    });

    // Database
    $container->set('dbh', function (ContainerInterface $c) {
        $settings = $c->get('settings')['database'];

        $dsn = sprintf('mysql:host=%s;port=%d;dbname=%s', $settings['host'], $settings['port'], $settings['dbname']);
        $pdo = new PDO($dsn, $settings['username'], $settings['password']);
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
        $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false);
        return $pdo;
    });

    // Session
    $container->set('session', function ($c) {
        return new SessionHelper();
    });

    // Service class
    $container->set(\App\Service::class, function(ContainerInterface $c) {
        return new \App\Service($c);
    });
};
