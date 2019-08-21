<?php

namespace App;

// A wrapper class for Environment Variables
class Environment
{
    public static function get($name, $default)
    {
        $val = getenv($name);
        if ($val === false) {
            return $default;
        }
        return $val;
    }
}
