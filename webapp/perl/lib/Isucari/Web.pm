package Isucari::Web;

use strict;
use warnings;
use utf8;
use Kossy;

use JSON::XS 3.00;
use DBIx::Sunny;
use Plack::Session;
use Time::Moment;
use File::Spec;

filter 'allow_json_request' => sub {
    my $app = shift;
    return sub {
        my ($self, $c) = @_;
        $c->env->{'kossy.request.parse_json_body'} = 1;
        $app->($self, $c);
    };
};

sub dbh {
    my $self = shift;
    $self->{_dbh} ||= do {
        my $host = $ENV{DB_HOST} // '127.0.0.1';
        my $port = $ENV{DB_PORT} // 3306;
        my $database = $ENV{DB_DATABASE} // 'isucari';
        my $user = $ENV{DB_USER} // 'isucari';
        my $password = $ENV{DB_PASS} // 'isucari';
        my $dsn = "dbi:mysql:database=$database;host=$host;port=$port";
        DBIx::Sunny->connect($dsn, $user, $password, {
            mysql_enable_utf8mb4 => 1,
            mysql_auto_reconnect => 1,
            Callbacks => {
                connected => sub {
                    my $dbh = shift;
                    # XXX $dbh->do('SET SESSION sql_mode="STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"');
                    return;
                },
            },
        });
    };
}

# getTop
get '/' => sub {
    my ( $self, $c )  = @_;
    open(my $fh, File::Spec->catfile($self->root_dir,'public/index.html')) or die $!;
    my $html = do {local $/; <$fh>};
    return $html;
};

# getNewItems
get '/new_items.json' => sub {
    my ( $self, $c )  = @_;

};

# getNewCategoryItems
get '/new_items/{root_category_id}.json' => sub {
    my ($self, $c) = @_;
};

# getTransactions
get '/users/transactions.json' => sub {
    my ($self, $c) = @_;
};

# getUserItems
get '/users/{user_id}.json' => sub {
    my ($self, $c) = @_;
};

# getItem
get '/items/{item_id}.json' => sub {
    my ($self, $c) = @_;
};

# postItemEdit
post '/items/edit' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# postBuy
post '/buy' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# postSell
post '/sell' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# postShip
post '/ship' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# postShipDone
post '/ship_done' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# postComplete
post '/complete' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# getQRCode
get '/transactions/{transaction_evidence_id}.png' => sub {
    my ($self, $c) = @_;
};

# postBump
post '/bump' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# getSettings
get '/settings' => sub {
    my ($self, $c) = @_;
};

# postLogin
post '/login' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# postRegister
post '/register' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
};

# XXX mux.Handle(pat.Get("/*"), http.FileServer(http.Dir("../public")))


1;
