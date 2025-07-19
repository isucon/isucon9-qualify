<?php

use Slim\App;
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Psr\Http\Server\RequestHandlerInterface as RequestHandler;
use Slim\Middleware\Session;

return function (App $app) {
    // Session middleware
    $app->add(new Session([
        'name' => 'session-isucari',
        'autorefresh' => true,
        'lifetime' => '1 hour',
    ]));
};
