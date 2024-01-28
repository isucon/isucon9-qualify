lib = File.expand_path('../lib', __FILE__)
$:.unshift(lib) unless $:.include?(lib)

require 'isucari/web'

use Rack::RewindableInput::Middleware
run Isucari::Web
