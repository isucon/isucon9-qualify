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

    public static function want($name)
    {
        $val = getenv($name);
        if ($val === false) {
            throw new \RuntimeException(sprintf("required env variable %s is missing", $name));
        }
        return $val;
    }
}
