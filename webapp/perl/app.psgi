use FindBin;
use lib "$FindBin::Bin/extlib/lib/perl5";
use lib "$FindBin::Bin/lib";
use File::Basename;
use Plack::Builder;
use Isucari::Web;

my $root_dir = File::Basename::dirname(__FILE__);

my $app = Isucari::Web->psgi($root_dir);
builder {
    enable 'ReverseProxy';
    enable 'Session::Cookie',
        session_key => 'session-isucari',
        expires     => 3600,
        secret      => 'tagomoris';
    enable 'Static',
        path => qr!^/(?:(?:static|upload|js)/|([^/]+)\.(?:js|png|ico)$|asset-manifest\.json$|manifest\.json$)!,
        root => $root_dir . '/public';
    $app;
};
