<?php

use Psr\Container\ContainerInterface;
use Slim\App;
use Slim\Http\StatusCode;

return function (App $app) {
    $container = $app->getContainer();

    // view renderer
    $container['renderer'] = function (ContainerInterface $c) {
        $settings = $c->get('settings')['renderer'];
        return new \Slim\Views\PhpRenderer($settings['template_path']);
    };

    // monolog
    $container['logger'] = function (ContainerInterface $c) {
        $settings = $c->get('settings')['logger'];
        $logger = new \Monolog\Logger($settings['name']);
        $logger->pushProcessor(new \Monolog\Processor\UidProcessor());
        $logger->pushHandler(new \Monolog\Handler\StreamHandler($settings['path'], $settings['level']));
        return $logger;
    };

    // database
    $container['dbh'] = function (ContainerInterface $c) {
        $settings = $c->get('settings')['database'];

        $dsn = sprintf('mysql:host=%s;port=%d;dbname=%s', $settings['host'], $settings['port'], $settings['dbname']);
        $options = [
            PDO::MYSQL_ATTR_INIT_COMMAND => 'SET NAMES utf8',
        ];
        $pdo = new \PDO($dsn, $settings['username'], $settings['password'], $options);
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
        $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false);
        return $pdo;
    };

    // session
    $container['session'] = function ($c) {
        return new \SlimSession\Helper;
    };

    // error logging
    $container['errorHandler'] =  function (ContainerInterface $c) {
        return function (Slim\Http\Request $request, Slim\Http\Response $response, \Exception $exception) use ($c) {
            $logger = $c['logger'];
            $logger->critical($exception->getMessage(), ['exception' => (string) $exception]);
            return $response->withStatus(StatusCode::HTTP_INTERNAL_SERVER_ERROR)
                ->withJson(['error' => $exception->getMessage()]);
        };
    };
};
