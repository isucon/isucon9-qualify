package Isucari::Web;

use strict;
use warnings;
use utf8;
use Kossy;

use JSON::XS 3.00;
use JSON::Types;
use DBIx::Sunny;
use Plack::Session;
use Time::Moment;
use File::Spec;
use HTTP::Date qw//;

# XXX sessionName = "session_isucari"

our $ITEM_MIN_PRICE    = 100;
our $TEM_MAX_PRICE    = 1000000;
our $ITEM_PRICE_ERRMSG = "商品価格は100円以上、1,000,000円以下にしてください";

our $ITEM_STATUS_ON_SALE  = "on_sale";
our $ITEM_STATUS_TRADING = "trading";
our $ITEM_STATUS_SOLD_OUT = "sold_out";
our $ITEM_STATUS_STOP    = "stop";
our $ITEM_STATUS_CANCEL  = "cancel";

our $PAYMENT_SERVICE_ISUCARI_APIKEY = "a15400e46c83635eb181-946abb51ff26a868317c";
our $PAYMENT_SERVICE_ISUCARI_SHOPID = "11";

our $TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING = "wait_shipping";
our $TRANSACTION_EVIDENCE_STATUS_WAIT_DONE     = "wait_done";
our $TRANSACTION_EVIDENCE_STATUS_DONE          = "done";

our $SHIPPINGS_STATUS_INITIAL    = "initial";
our $SHIPPINGS_STATUS_WAIT_PICKUP = "wait_pickup";
our $SHIPPINGS_STATUS_SHIPPING   = "shipping";
our $SHIPPINGS_STATUS_DONE       = "done";

our $BUMP_CHARGE_SECONDS = 3;

our $ITEMS_PER_PAGE = 48;
our $TRANSACTIONS_PER_PAGE = 10;


filter 'allow_json_request' => sub {
    my $app = shift;
    return sub {
        my ($self, $c) = @_;
        $c->env->{'kossy.request.parse_json_body'} = 1;
        $app->($self, $c);
    };
};

sub unix_from_mysql_datetime {
    my $str = shift;
    return HTTP::Date::str2time($str)
}

sub mysql_datetime_from_unix {
    my $time = shift;
    my @lt = localtime($time);
    sprintf("%04d-%02d-%02d %02d:%02d:%02d", $lt[5]+1900,$lt[4]+1,$lt[3],$lt[2],$lt[1],$lt[0]);
}

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

sub getUserSimpleByID {
    my ($self, $user_id) = @_;
    my $user = $self->dbh->select_row('SELECT * FROM `users` WHERE `id` = ?',$user_id);
    return unless $user;
    return +{
        id => number $user->{id},
        account_name => $user->{account_name},
        num_sell_items => number $user->{num_sell_items},
    };
}

sub getCategoryByID {
    my ($self, $id) = @_;
    my $category = $self->dbh->select_row('SELECT * FROM `categories` WHERE `id` = ?',$id);
    return unless $category;
    if ($category->{parent_id} != 0) {
        my $parent_category = $self->getCategoryByID($category->{parent_id});
        return unless $parent_category;
        $category->{parent_category_name} = $parent_category->{name};
    }
    return +{
        id => number $category->{id},
        parent_id => number $category->{parent_id},
        category_name => $category->{category_name},
        parent_category_name => $category->{parent_category_name}
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
    my $item_id = $c->req->parameters->get('item_id');
    my $created_at = $c->req->parameters->get('created_at');

    my $items = [];
    if ($item_id && $created_at) {
        # paging
        $items = $self->dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE `status` IN (?,?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT %d', $ITEMS_PER_PAGE+1),
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_SOLD_OUT,
            mysql_datetime_from_unix($created_at),
            $item_id,
        );
    }
    else {
        # 1st page
        $items = $self->dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE `status` IN (?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT %d',$ITEMS_PER_PAGE+1),
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_SOLD_OUT,
        );
    }

    my @item_simples = ();
    for my $item (@$items) {
        my $seller = $self->getUserSimpleByID($item->{seller_id});
        if (!$seller) {
            $c->halt(404,"seller not found"); #XXX
        }
        my $category = $self->getCategoryByID($item->{category_id});
		if (!$category) {
            $c->halt(404,"category not found"); #XXX
		}
        push @item_simples, +{
            id => number $item->{id},
            seller_id => number $item->{seller_id},
            seller => $seller,
            status => $item->{status},
            name => $item->{name},
            price => number $item->{price},
            category_id => number $item->{category_id},
            category => $category,
            created_at => number unix_from_mysql_datetime($item->{created_at}),
        }
    }

    my $has_next = 0;
	if (@item_simples > $ITEMS_PER_PAGE) {
		$has_next = 1;
        pop @item_simples;
	}

    $c->render_json({
        items => \@item_simples,
        hax_next => bool $has_next
    });
};

# getNewCategoryItems
get '/new_items/{root_category_id}.json' => sub {
    my ($self, $c) = @_;
    my $root_category_id = $c->args->{root_category_id};
    my $item_id = $c->req->parameters->get('item_id');
    my $created_at = $c->req->parameters->get('created_at');

    my $root_category = $self->getCategoryByID($root_category_id);
    if (!$root_category) {
        $c->halt(404,"root category not found"); #XXX
    }

    my $categories = $self->dbh->select_all('SELECT id FROM `categories` WHERE parent_id=?', $root_category_id);
    my @category_ids = map {$_->{id}} @$categories;

    my $items = [];
    if ($item_id && $created_at) {
        # paging
        $items = $self->dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (?) AND `created_at` <= ? AND `id` < ? ORDER BY `created_at` DESC, `id` DESC LIMIT %d', $ITEMS_PER_PAGE+1),
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_SOLD_OUT,
            \@category_ids,
            mysql_datetime_from_unix($created_at),
            $item_id,
        );
    }
    else {
        # 1st page
        $items = $self->dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (?) ORDER BY `created_at` DESC, `id` DESC LIMIT %d',$ITEMS_PER_PAGE+1),
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_SOLD_OUT,
            \@category_ids,
        );
    }

    my @item_simples = ();
    for my $item (@$items) {
        my $seller = $self->getUserSimpleByID($item->{seller_id});
        if (!$seller) {
            $c->halt(404,"seller not found"); #XXX
        }
        my $category = $self->getCategoryByID($item->{category_id});
		if (!$category) {
            $c->halt(404,"category not found"); #XXX
		}
        push @item_simples, +{
            id => number $item->{id},
            seller_id => number $item->{seller_id},
            seller => $seller,
            status => $item->{status},
            name => $item->{name},
            price => number $item->{price},
            category_id => number $item->{category_id},
            category => $category,
            created_at => number unix_from_mysql_datetime($item->{created_at}),
        }
    }

    my $has_next = 0;
	if (@item_simples > $ITEMS_PER_PAGE) {
		$has_next = 1;
        pop @item_simples;
	}

    $c->render_json({
        root_category_id => number $root_category->{id},
        root_category_name => $root_category->{name},
        items => \@item_simples,
        hax_next => bool $has_next
    });
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
