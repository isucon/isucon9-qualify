package Isucari::Web;

use strict;
use warnings;
use utf8;
use Kossy;

use JSON::XS 3.00;
use JSON::Types;
use DBIx::Sunny;
use Plack::Session;
use File::Spec;
use HTTP::Date qw//;
use HTTP::Status qw/:constants/;
use Crypt::Eksblowfish::Bcrypt qw/bcrypt/;
use Crypt::OpenSSL::Random;
use Digest::SHA;
use File::Basename;
use File::Copy;

use Isucari::API;

our $DEFAULT_PAYMENT_SERVICE_URL  = "http://localhost:5555";
our $DEFAULT_SHIPMENT_SERVICE_URL = "http://localhost:7000";

our $ITEM_MIN_PRICE    = 100;
our $ITEM_MAX_PRICE    = 1000000;
our $ITEM_PRICE_ERRMSG = "商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下にしてください";

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

our $BCRYPT_COST = 10;

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

sub encrypt_password {
    my $password = shift;
    my $salt = shift || Crypt::Eksblowfish::Bcrypt::en_base64(Crypt::OpenSSL::Random::random_bytes(16));
    my $settings = '$2a$'.$BCRYPT_COST.'$'.$salt;
    return Crypt::Eksblowfish::Bcrypt::bcrypt($password, $settings);
}

sub check_password {
    my ($plain_password, $hashed_password) = @_;
    if ($hashed_password =~ m!^\$2a\$\d{2}\$(.+)$!) {
        return encrypt_password($plain_password, $1) eq $hashed_password;
    }
    die "crypt_error";
}

sub secure_random_str {
    my $length = shift || 16;
    unpack("H*",Crypt::OpenSSL::Random::random_bytes($length))
}

sub get_image_url {
    my $image_name = shift;
    sprintf("/upload/%s", $image_name)
}

sub error_with_msg {
    my ($self, $c, $status, $msg) = @_;
    $c->res->code($status);
    $c->res->content_type('application/json;charset=utf-8');
    $c->res->body(JSON::encode_json({error => $msg}));
    $c->res;
}

sub dbh {
    my $self = shift;
    $self->{_dbh} ||= do {
        my $host = $ENV{MYSQL_HOST} // '127.0.0.1';
        my $port = $ENV{MYSQL_PORT} // 3306;
        my $database = $ENV{MYSQL_DBNAME} // 'isucari';
        my $user = $ENV{MYSQL_USER} // 'isucari';
        my $password = $ENV{MYSQL_PASS} // 'isucari';
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

sub api_client {
    my $self = shift;
    $self->{_api_client} ||= do {
        Isucari::API->new();
    };
}

sub getCSRFToken {
    my ($self, $c) = @_;
    my $session = Plack::Session->new($c->env);
    return $session->get('csrf_token') // "";
}

sub getUser {
    my ($self, $c) = @_;
    my $session = Plack::Session->new($c->env);
    my $user_id = $session->get('user_id');
    return unless $user_id;
    return $self->dbh->select_row('SELECT * FROM users WHERE id = ?', $user_id);
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
        $category->{parent_category_name} = $parent_category->{category_name};
    }
    return +{
        id => number $category->{id},
        parent_id => number $category->{parent_id},
        category_name => $category->{category_name},
        parent_category_name => $category->{parent_category_name}
    };
}

sub getConfigByName {
    my ($self, $name) = @_;
    my $row = $self->dbh->select_row('SELECT * FROM `configs` WHERE `name` = ?', $name);
    return unless $row;
    $row->{val};
}
sub getPaymentServiceURL {
    my $self = shift;
    $self->getConfigByName('payment_service_url') || $DEFAULT_PAYMENT_SERVICE_URL;

}
sub getShipmentServiceURL {
    my $self = shift;
    $self->getConfigByName('shipment_service_url') || $DEFAULT_SHIPMENT_SERVICE_URL;
}

# Frontend: getIndex
my $get_index = sub {
    my ( $self, $c )  = @_;
    open(my $fh, File::Spec->catfile($self->root_dir,'public/index.html')) or die $!;
    my $html = do {local $/; <$fh>};
    return $html;
};
get '/' => $get_index;
get '/login' => $get_index;
get '/register' => $get_index;
get '/timeline' => $get_index;
get '/categories/{category_id:\d+}/items' => $get_index;
get '/sell' => $get_index;
get '/items/{item_id:\d+}' => $get_index;
get '/items/{item_id:\d+}/edit' => $get_index;
get '/items/{item_id:\d+}/buy' => $get_index;
get '/buy/complete' => $get_index;
get '/transactions/{transaction_id:\d+}' => $get_index;
get '/users/{user_id:\d+}' => $get_index;
get '/users/setting' => $get_index;

# postInitialize
post '/initialize' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;

    # TODO initialize data
    my $ret = system+File::Spec->catfile($self->root_dir, "../sql/init.sh");
    if ( $ret != 0 ) {
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "exec init.sh error");
    }


    for my $name (qw/payment_service_url shipment_service_url/) {
        $self->dbh->query(
            'INSERT INTO `configs` (name, val) VALUES (?,?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)',
            $name,
            $c->req->body_parameters->get($name) // ""
        );
    }

    $c->render_json({
        # キャンペーン実施時には還元率の設定を返す。詳しくはマニュアルを参照のこと。
        campaign => 0,
        # 実装言語を返す
        language => "perl"
    });
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
            sprintf('SELECT * FROM `items` WHERE `status` IN (?,?) AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT %d', $ITEMS_PER_PAGE+1),
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_SOLD_OUT,
            mysql_datetime_from_unix($created_at),
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
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "seller not found");
        }
        my $category = $self->getCategoryByID($item->{category_id});
		if (!$category) {
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "category not found");
		}
        push @item_simples, +{
            id => number $item->{id},
            seller_id => number $item->{seller_id},
            seller => $seller,
            status => $item->{status},
            name => $item->{name},
            price => number $item->{price},
            image_url => get_image_url($item->{image_name}),
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
        has_next => bool $has_next
    });
};

# getNewCategoryItems
get '/new_items/{root_category_id:\d+}.json' => sub {
    my ($self, $c) = @_;
    my $root_category_id = $c->args->{root_category_id};
    my $item_id = $c->req->parameters->get('item_id');
    my $created_at = $c->req->parameters->get('created_at');

    my $root_category = $self->getCategoryByID($root_category_id);
    if (!$root_category) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "root category not found");
    }

    my $categories = $self->dbh->select_all('SELECT id FROM `categories` WHERE parent_id=?', $root_category_id);
    my @category_ids = map {$_->{id}} @$categories;

    my $items = [];
    if ($item_id && $created_at) {
        # paging
        $items = $self->dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE `status` IN (?,?) AND category_id IN (?) AND (`created_at` < ?  OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT %d', $ITEMS_PER_PAGE+1),
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_SOLD_OUT,
            \@category_ids,
            mysql_datetime_from_unix($created_at),
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
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "seller not found");
        }
        my $category = $self->getCategoryByID($item->{category_id});
		if (!$category) {
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "category not found");
		}
        push @item_simples, +{
            id => number $item->{id},
            seller_id => number $item->{seller_id},
            seller => $seller,
            status => $item->{status},
            name => $item->{name},
            price => number $item->{price},
            image_url => get_image_url($item->{image_name}),
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
        root_category_name => $root_category->{category_name},
        items => \@item_simples,
        has_next => bool $has_next
    });
};

# getUserItems
get '/users/{user_id:\d+}.json' => sub {
    my ($self, $c) = @_;
    my $user_id = $c->args->{user_id};
    my $item_id = $c->req->parameters->get('item_id');
    my $created_at = $c->req->parameters->get('created_at');

    my $user_simple = $self->getUserSimpleByID($user_id);
    if (!$user_simple) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $items = [];
    if ($item_id && $created_at) {
        # paging
        $items = $self->dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE `status` IN (?,?,?) AND seller_id = ? AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT %d', $ITEMS_PER_PAGE+1),
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_TRADING,
            $ITEM_STATUS_SOLD_OUT,
            $user_simple->{id},
            mysql_datetime_from_unix($created_at),
            mysql_datetime_from_unix($created_at),
            $item_id,
        );
    }
    else {
        # 1st page
        $items = $self->dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE `status` IN (?,?,?) AND seller_id = ? ORDER BY `created_at` DESC, `id` DESC LIMIT %d',$ITEMS_PER_PAGE+1),
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_TRADING,
            $ITEM_STATUS_SOLD_OUT,
            $user_simple->{id},
        );
    }

    my @item_simples = ();
    for my $item (@$items) {
        my $seller = $self->getUserSimpleByID($item->{seller_id});
        if (!$seller) {
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "seller not found");
        }
        my $category = $self->getCategoryByID($item->{category_id});
		if (!$category) {
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "category not found");
		}
        push @item_simples, +{
            id => number $item->{id},
            seller_id => number $item->{seller_id},
            seller => $seller,
            status => $item->{status},
            name => $item->{name},
            price => number $item->{price},
            image_url => get_image_url($item->{image_name}),
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
        user => $user_simple,
        items => \@item_simples,
        has_next => bool $has_next
    });
};

# getTransactions
get '/users/transactions.json' => sub {
    my ($self, $c) = @_;
    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }
    my $item_id = $c->req->parameters->get('item_id');
    my $created_at = $c->req->parameters->get('created_at');

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    my $items = [];
    if ($item_id && $created_at) {
        # paging
        $items = $dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE (`seller_id` = ? OR `buyer_id` = ?) AND `status` IN (?,?,?,?,?) AND (`created_at` < ? OR (`created_at` <= ? AND `id` < ?)) ORDER BY `created_at` DESC, `id` DESC LIMIT %d', $TRANSACTIONS_PER_PAGE+1),
            $user->{id},
            $user->{id},
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_TRADING,
            $ITEM_STATUS_SOLD_OUT,
            $ITEM_STATUS_CANCEL,
            $ITEM_STATUS_STOP,
            mysql_datetime_from_unix($created_at),
            mysql_datetime_from_unix($created_at),
            $item_id,
        );
    }
    else {
        # 1st page
        $items = $dbh->select_all(
            sprintf('SELECT * FROM `items` WHERE (`seller_id` = ? OR `buyer_id` = ?) AND `status` IN (?,?,?,?,?) ORDER BY `created_at` DESC, `id` DESC LIMIT %d',$TRANSACTIONS_PER_PAGE+1),
            $user->{id},
            $user->{id},
            $ITEM_STATUS_ON_SALE,
            $ITEM_STATUS_TRADING,
            $ITEM_STATUS_SOLD_OUT,
            $ITEM_STATUS_CANCEL,
            $ITEM_STATUS_STOP,
        );
    }

    my @item_details = ();
    for my $item (@$items) {
        my $seller = $self->getUserSimpleByID($item->{seller_id});
        if (!$seller) {
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "seller not found");
        }
        my $category = $self->getCategoryByID($item->{category_id});
        if (!$category) {
            return $self->error_with_msg($c, HTTP_NOT_FOUND, "category not found");
        }

        my $item_detail = +{
            id => number $item->{id},
            seller_id => number $item->{seller_id},
            seller => $seller,
            # buyer_id
            # buyer
            status => $item->{status},
            name => $item->{name},
            price => number $item->{price},
            description => $item->{description},
            image_url => get_image_url($item->{image_name}),
            category_id => number $item->{category_id},
            # transaction_evidence_id
            # transaction_evidence_status
            # shipping_status
            category => $category,
            created_at => number unix_from_mysql_datetime($item->{created_at}),
        };

        if ($item->{buyer_id} != 0) {
            my $buyer = $self->getUserSimpleByID($item->{buyer_id});
            if (!$buyer) {
                return $self->error_with_msg($c, HTTP_NOT_FOUND, 'buyer not found');
            }
            $item_detail->{buyer_id} = number $item->{buyer_id};
            $item_detail->{buyer} = $buyer;
        }

        my $transaction_evidence = $dbh->select_row(
            'SELECT * FROM `transaction_evidences` WHERE `item_id` = ?',
            $item->{id}
        );

        if ($transaction_evidence) {
            my $shipping = $dbh->select_row(
                'SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?',
                $transaction_evidence->{id}
            );
            if (!$shipping) {
                return $self->error_with_msg($c, HTTP_NOT_FOUND, 'shipping not found');
            }

            my $ssr = eval {
                $self->api_client->shipment_status(
                    $self->getShipmentServiceURL(),
                    {reserve_id => string $shipping->{reserve_id}}
                );
            };
            if ($@) {
                warn $@;
                return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "failed to request to shipment service");
            }

            $item_detail->{transaction_evidence_id} = number $transaction_evidence->{id};
            $item_detail->{transaction_evidence_status} = $transaction_evidence->{status};
            $item_detail->{shipping_status} = $ssr->{status};
        }

        push @item_details, $item_detail;
    }

    $txn->commit();

    my $has_next = 0;
    if (@item_details > $TRANSACTIONS_PER_PAGE) {
        $has_next = 1;
        pop @item_details;
    }

    $c->render_json({
        items => \@item_details,
        has_next => bool $has_next
    });
};

# getItem
get '/items/{item_id:\d+}.json' => sub {
    my ($self, $c) = @_;
    my $item_id = $c->args->{item_id};

    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $item = $self->dbh->select_row('SELECT * FROM `items` WHERE `id` = ?', $item_id);
    if (!$item) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'item not found');
    }

    my $seller = $self->getUserSimpleByID($item->{seller_id});
    if (!$seller) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND,"seller not found");
    }
    my $category = $self->getCategoryByID($item->{category_id});
    if (!$category) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "category not found");
    }

    my $item_detail = +{
        id => number $item->{id},
        seller_id => number $item->{seller_id},
        seller => $seller,
        # buyer_id
        # buyer
        status => $item->{status},
        name => $item->{name},
        price => number $item->{price},
        description => $item->{description},
        image_url => get_image_url($item->{image_name}),
        category_id => number $item->{category_id},
        # transaction_evidence_id
        # transaction_evidence_status
        # shipping_status
        category => $category,
        created_at => number unix_from_mysql_datetime($item->{created_at}),
    };

    # if (user.ID == item.SellerID || user.ID == item.BuyerID) && item.BuyerID != 0 {
    if (($user->{id} == $item->{seller_id} || $user->{id} == $item->{buyer_id}) && $item->{buyer_id} != 0) {
        my $buyer = $self->getUserSimpleByID($item->{buyer_id});
        if (!$buyer) {
            return $self->error_with_msg($c, HTTP_NOT_FOUND, 'buyer not found');
        }
        $item_detail->{buyer_id} = $item->{buyer_id};
        $item_detail->{buyer} = $buyer;

        my $transaction_evidence = $self->dbh->select_row(
            'SELECT * FROM `transaction_evidences` WHERE `item_id` = ?',
            $item->{id}
        );

        if ($transaction_evidence) {
            my $shipping = $self->dbh->select_row(
                'SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?',
                $transaction_evidence->{id}
            );
            if (!$shipping) {
                return $self->error_with_msg($c, HTTP_NOT_FOUND,'shipping not found');
            }

            $item_detail->{transaction_evidence_id} = number $transaction_evidence->{id};
            $item_detail->{transaction_evidence_status} = $transaction_evidence->{status};
            $item_detail->{shipping_status} = $shipping->{status};
        }

    }

    $c->render_json($item_detail);
};

# postItemEdit
post '/items/edit' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;

    my $csrf_token = $c->req->body_parameters->get('csrf_token') // "";
	my $item_id    = $c->req->body_parameters->get('item_id') // "";
	my $price      = $c->req->body_parameters->get('item_price') // 0;

    if ($csrf_token ne $self->getCSRFToken($c)) {
        return $self->error_with_msg($c, HTTP_UNPROCESSABLE_ENTITY, 'csrf token error');
    }

    if ($price < $ITEM_MIN_PRICE || $price > $ITEM_MAX_PRICE) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, $ITEM_PRICE_ERRMSG);
	}

    my $seller = $self->getUser($c);
    if (!$seller) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $target_item = $self->dbh->select_row('SELECT * FROM `items` WHERE `id` = ?', $item_id);
    if (!$target_item) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'item not found');
    }
    if ($target_item->{seller_id} != $seller->{id}) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "自分の商品以外は編集できません");
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();
    $target_item = $dbh->select_row('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE', $item_id);

    if ($target_item->{status} ne $ITEM_STATUS_ON_SALE) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "販売中の商品以外編集できません");
	}

    $dbh->query(
        'UPDATE `items` SET `price` = ?, `updated_at` = ? WHERE `id` = ?',
        $price,
        mysql_datetime_from_unix(time),
        $item_id
    );

    $target_item = $dbh->select_row('SELECT * FROM `items` WHERE `id` = ?', $item_id);

    $txn->commit();

    $c->render_json({
        item_id => number $target_item->{id},
        item_price => number $target_item->{price},
        item_created_at => number unix_from_mysql_datetime($target_item->{created_at}),
        item_updated_at => number unix_from_mysql_datetime($target_item->{updated_at}),
    });
};

# postBuy
post '/buy' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
    my $csrf_token = $c->req->body_parameters->get('csrf_token') // "";
    my $item_id    = $c->req->body_parameters->get('item_id') // "";
    my $token    = $c->req->body_parameters->get('token') // "";

    if ($csrf_token ne $self->getCSRFToken($c)) {
        return $self->error_with_msg($c, HTTP_UNPROCESSABLE_ENTITY, 'csrf token error');
    }

    my $buyer = $self->getUser($c);
    if (!$buyer) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    my $target_item = $dbh->select_row('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE', $item_id);
    if (!$target_item) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'item not found');
    }
    if ($target_item->{status} ne $ITEM_STATUS_ON_SALE) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "item is not for sale");
	}
    if ($target_item->{seller_id} == $buyer->{id}) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "自分の商品は買えません")
    }

    my $seller = $dbh->select_row('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE', $target_item->{seller_id});
    if (!$seller) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'seller not found');
    }

    my $category = $self->getCategoryByID($target_item->{category_id});
    if (!$category) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'category id error');
    }

    $dbh->query(
        'INSERT INTO `transaction_evidences` (`seller_id`, `buyer_id`, `status`, `item_id`, `item_name`, `item_price`, `item_description`,`item_category_id`,`item_root_category_id`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
        $target_item->{seller_id},
        $buyer->{id},
        $TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING,
        $target_item->{id},
        $target_item->{name},
        $target_item->{price},
        $target_item->{description},
        $category->{id},
        $category->{parent_id}
    );

    my $transaction_evidence_id = $dbh->last_insert_id();

    $dbh->query(
        'UPDATE `items` SET `buyer_id` = ?, `status` = ?, `updated_at` = ? WHERE `id` = ?',
        $buyer->{id},
        $ITEM_STATUS_TRADING,
        mysql_datetime_from_unix(time),
        $target_item->{id},
    );

    my $scr = eval{
        $self->api_client->shipment_create($self->getShipmentServiceURL(), {
            to_address   => $buyer->{address},
            to_name      => $buyer->{account_name},
            from_address => $seller->{address},
            from_name    => $seller->{account_name}
        });
    };
    if ($@) {
        warn $@;
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "failed to request to shipment service");
    }

    my $pstr = eval {
        $self->api_client->payment_token($self->getPaymentServiceURL(), {
            shop_id => $PAYMENT_SERVICE_ISUCARI_SHOPID,
            token   => $token,
            api_key => $PAYMENT_SERVICE_ISUCARI_APIKEY,
            price   => number $target_item->{price},
        })
    };
    if ($@) {
        warn $@;
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "payment service is failed");
    }

    if ($pstr->{status} eq "invalid") {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "カード情報に誤りがあります");
	}

	if ($pstr->{status} eq "fail") {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "カードの残高が足りません");
	}

	if ($pstr->{status} ne "ok") {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "想定外のエラー");
	}

    $dbh->query(
        'INSERT INTO `shippings` (`transaction_evidence_id`, `status`, `item_name`, `item_id`, `reserve_id`, `reserve_time`, `to_address`, `to_name`, `from_address`, `from_name`, `img_binary`) VALUES (?,?,?,?,?,?,?,?,?,?,?)',
		$transaction_evidence_id,
		$SHIPPINGS_STATUS_INITIAL,
        $target_item->{name},
        $target_item->{id},
        $scr->{reserve_id},
        $scr->{reserve_time},
		$buyer->{address},
		$buyer->{account_name},
		$seller->{address},
		$seller->{account_name},
		"", # default img_binary
	);
    $txn->commit();

    $c->render_json({
        transaction_evidence_id => number $transaction_evidence_id
    });
};

# postSell
post '/sell' => sub {
    my ($self, $c) = @_;
    my $csrf_token = $c->req->body_parameters->get('csrf_token') // "";
    my $name    = $c->req->body_parameters->get('name') // "";
    my $price    = $c->req->body_parameters->get('price') // 0;
    my $description    = $c->req->body_parameters->get('description') // "";
    my $category_id    = $c->req->body_parameters->get('category_id') // 0;

    if ($csrf_token ne $self->getCSRFToken($c)) {
        return $self->error_with_msg($c, HTTP_UNPROCESSABLE_ENTITY, 'csrf token error');
    }

    if ($name eq "" || $description eq "" || $price == 0 || $category_id == 0) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "all parameters are required");
	}

    if ($price < $ITEM_MIN_PRICE || $price > $ITEM_MAX_PRICE) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, $ITEM_PRICE_ERRMSG);
	}

    my $category = $self->getCategoryByID($category_id);
    if (!$category) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "incorrect category id");
    }

    my $upload = $c->req->uploads->{'image'};
    my ($filename, $dirs, $ext) = fileparse($upload->basename,qr/\.[^.]*/);
    if ($ext ne ".jpg" && $ext ne ".jpeg" && $ext ne ".png" && $ext ne ".gif") {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "unsupported image format error");
    }
    if ($ext eq ".jpeg") {
        $ext = ".jpg"
    }

    my $image_name = sprintf("%s%s", secure_random_str(16), $ext);
    if (!copy($upload->path, File::Spec->catfile($self->root_dir,'public/upload',$image_name))) {
        warn $!;
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "Saving image failed");
    }

    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    my $seller = $dbh->select_row('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE', $user->{id});
    if (!$seller) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'seller not found');
    }

    $dbh->query(
        'INSERT INTO `items` (`seller_id`, `status`, `name`, `price`, `description`,`image_name`,`category_id`) VALUES (?, ?, ?, ?, ?, ?, ?)',
        $seller->{id},
        $ITEM_STATUS_ON_SALE,
        $name,
        $price,
        $description,
        $image_name,
        $category->{id},
    );
    my $item_id = $dbh->last_insert_id();

    $dbh->query(
        'UPDATE `users` SET `num_sell_items`=?, `last_bump`=? WHERE `id`=?',
        $seller->{num_sell_items} + 1,
        mysql_datetime_from_unix(time),
        $seller->{id},
    );

    $txn->commit();

    $c->render_json({id => $item_id});
};

# postShip
post '/ship' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
    my $csrf_token = $c->req->body_parameters->get('csrf_token') // "";
    my $item_id    = $c->req->body_parameters->get('item_id') // "";

    if ($csrf_token ne $self->getCSRFToken($c)) {
        return $self->error_with_msg($c, HTTP_UNPROCESSABLE_ENTITY, 'csrf token error');
    }

    my $seller = $self->getUser($c);
    if (!$seller) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $transaction_evidence = $self->dbh->select_row(
        'SELECT * FROM `transaction_evidences` WHERE `item_id` = ?',
        $item_id,
    );
    if (!$transaction_evidence) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "transaction_evidences not found");
    }

    if ($transaction_evidence->{seller_id} != $seller->{id}) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, '権限がありません');
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    my $item = $dbh->select_row('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE', $item_id);
    if (!$item_id) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'item not found');
    }
    if ($item->{status} ne $ITEM_STATUS_TRADING) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "商品が取引中ではありません");
    }

    my $target_transaction_evidence = $dbh->select_row(
        'SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE',
        $transaction_evidence->{id}
    );
    if (!$target_transaction_evidence) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'transaction_evidences not found');
    }
    if ($target_transaction_evidence->{status} ne $TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "準備ができていません");
    }

    my $shipping = $dbh->select_row(
        'SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE',
        $target_transaction_evidence->{id},
    );
    if (!$shipping) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'shipping not found');
    }

    my $img_binary = eval {
        $self->api_client->shipment_request(
            $self->getShipmentServiceURL(),
            {reserve_id => string $shipping->{reserve_id}}
            );
    };
    if ($@) {
        warn $@;
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, 'failed to request to shipment service');
    }

    $dbh->query(
        'UPDATE `shippings` SET `status` = ?, `img_binary` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?',
        $SHIPPINGS_STATUS_WAIT_PICKUP,
        $img_binary,
        mysql_datetime_from_unix(time),
        $target_transaction_evidence->{id},
    );

    $txn->commit();

    $c->render_json({
        reserve_id => $shipping->{reserve_id},
        path => sprintf('/transactions/%d.png', $target_transaction_evidence->{id})
    })
};

# postShipDone
post '/ship_done' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
    my $csrf_token = $c->req->body_parameters->get('csrf_token') // "";
    my $item_id    = $c->req->body_parameters->get('item_id') // "";

    if ($csrf_token ne $self->getCSRFToken($c)) {
        return $self->error_with_msg($c, HTTP_UNPROCESSABLE_ENTITY, 'csrf token error');
    }

    my $seller = $self->getUser($c);
    if (!$seller) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $transaction_evidence = $self->dbh->select_row(
        'SELECT * FROM `transaction_evidences` WHERE `item_id` = ?',
        $item_id
    );
    if (!$transaction_evidence) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'transaction_evidences not found');
    }
    if ($transaction_evidence->{seller_id} != $seller->{id}) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "権限がありません");
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    my $item = $dbh->select_row('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE', $item_id);
    if (!$item_id) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'item not found');
    }
    if ($item->{status} ne $ITEM_STATUS_TRADING) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "商品が取引中ではありません");
    }

    my $target_transaction_evidence = $dbh->select_row(
        'SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE',
        $transaction_evidence->{id}
    );
    if (!$target_transaction_evidence) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'transaction_evidences not found');
    }
    if ($target_transaction_evidence->{status} ne $TRANSACTION_EVIDENCE_STATUS_WAIT_SHIPPING) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "準備ができていません");
    }

    my $shipping = $dbh->select_row(
        'SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE',
        $target_transaction_evidence->{id},
    );
    if (!$shipping) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'shipping not found');
    }

    my $ssr = eval {
        $self->api_client->shipment_status(
            $self->getShipmentServiceURL(),
            {reserve_id => string $shipping->{reserve_id}},
        )
    };
    if ($@) {
        warn $@;
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, "failed to request to shipment service");
    }
    if (!($ssr->{status} eq $SHIPPINGS_STATUS_SHIPPING || $ssr->{status} eq $SHIPPINGS_STATUS_DONE)) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, 'shipment service側で配送中か配送完了になっていません');
    }

    my $now = time;
    $dbh->query(
        'UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?',
        $ssr->{status},
        mysql_datetime_from_unix($now),
        $target_transaction_evidence->{id},
    );

    $dbh->query(
        'UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?',
        $TRANSACTION_EVIDENCE_STATUS_WAIT_DONE,
        mysql_datetime_from_unix($now),
        $target_transaction_evidence->{id},
    );

    $txn->commit();
    $c->render_json({transaction_evidence_id => $target_transaction_evidence->{id}})
};

# postComplete
post '/complete' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
    my $csrf_token = $c->req->body_parameters->get('csrf_token') // "";
    my $item_id    = $c->req->body_parameters->get('item_id') // "";

    if ($csrf_token ne $self->getCSRFToken($c)) {
        return $self->error_with_msg($c, HTTP_UNPROCESSABLE_ENTITY, 'csrf token error');
    }

    my $buyer = $self->getUser($c);
    if (!$buyer) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $transaction_evidence = $self->dbh->select_row(
        'SELECT * FROM `transaction_evidences` WHERE `item_id` = ?',
        $item_id,
    );
    if (!$transaction_evidence) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "transaction_evidence not found");
    }
    if ($transaction_evidence->{buyer_id} != $buyer->{id}) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, '権限がありません');
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    my $item = $dbh->select_row('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE', $item_id);
    if (!$item_id) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'item not found');
    }
    if ($item->{status} ne $ITEM_STATUS_TRADING) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "商品が取引中ではありません");
    }

    my $target_transaction_evidence = $dbh->select_row(
        'SELECT * FROM `transaction_evidences` WHERE `id` = ? FOR UPDATE',
        $transaction_evidence->{id}
    );
    if (!$target_transaction_evidence) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'transaction_evidences not found');
    }
    if ($target_transaction_evidence->{status} ne $TRANSACTION_EVIDENCE_STATUS_WAIT_DONE) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, "準備ができていません");
    }

    my $shipping = $dbh->select_row(
        'SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE',
        $target_transaction_evidence->{id},
    );
    if (!$shipping) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'shipping not found');
    }

    my $ssr = eval {
        $self->api_client->shipment_status(
            $self->getShipmentServiceURL(),
            {reserve_id => string $shipping->{reserve_id}},
        );
    };
    if ($@) {
        warn $@;
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, 'failed to request to shipment service');
    }
    if (!($ssr->{status} eq $SHIPPINGS_STATUS_DONE)) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, "shipment service側で配送完了になっていません");
    }

    my $now = time;
    $dbh->query(
        'UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?',
        $SHIPPINGS_STATUS_DONE,
        mysql_datetime_from_unix($now),
        $target_transaction_evidence->{id},
    );
    $dbh->query(
        'UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?',
        $TRANSACTION_EVIDENCE_STATUS_DONE,
        mysql_datetime_from_unix($now),
        $target_transaction_evidence->{id},
    );
    $dbh->query(
        'UPDATE `items` SET `status` = ?, `updated_at` = ? WHERE `id` = ?',
        $ITEM_STATUS_SOLD_OUT,
        mysql_datetime_from_unix($now),
        $item->{id}
    );

    $txn->commit();
    $c->render_json({transaction_evidence_id => $target_transaction_evidence->{id}});
};

# getQRCode
get '/transactions/{transaction_evidence_id:\d+}.png' => sub {
    my ($self, $c) = @_;
    my $transaction_evidence_id = $c->args->{transaction_evidence_id};

    my $seller = $self->getUser($c);
    if (!$seller) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $transaction_evidence = $self->dbh->select_row(
        'SELECT * FROM `transaction_evidences` WHERE `id` = ?',
        $transaction_evidence_id,
    );
    if (!$transaction_evidence) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, "transaction_evidence not found");
    }
    if ($transaction_evidence->{seller_id} != $seller->{id}) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, '権限がありません');
    }

    my $shipping = $self->dbh->select_row(
        'SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ?',
        $transaction_evidence->{id},
    );
    if (!$shipping) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'shippings not found');
    }

    if ($shipping->{status} ne $SHIPPINGS_STATUS_WAIT_PICKUP && $shipping->{status} ne $SHIPPINGS_STATUS_SHIPPING) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, 'qrcode not available');
    }

    if (length($shipping->{img_binary}) == 0) {
        return $self->error_with_msg($c, HTTP_INTERNAL_SERVER_ERROR, 'empty qrcode image');
    }

    $c->res->content_type('image/png');
    return $shipping->{img_binary};
};

# postBump
post '/bump' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
    my $csrf_token = $c->req->body_parameters->get('csrf_token') // "";
    my $item_id    = $c->req->body_parameters->get('item_id') // "";

    if ($csrf_token ne $self->getCSRFToken($c)) {
        return $self->error_with_msg($c, HTTP_UNPROCESSABLE_ENTITY, 'csrf token error');
    }

    my $user = $self->getUser($c);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $dbh = $self->dbh;
    my $txn = $dbh->txn_scope();

    my $target_item = $dbh->select_row('SELECT * FROM `items` WHERE `id` = ? FOR UPDATE', $item_id);
    if (!$target_item) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'item not found');
    }
    if ($target_item->{seller_id} != $user->{id}) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, '自分の商品以外は編集できません');
    }

    my $seller = $dbh->select_row('SELECT * FROM `users` WHERE `id` = ? FOR UPDATE', $user->{id});
    if (!$seller) {
        return $self->error_with_msg($c, HTTP_NOT_FOUND, 'user not found');
    }

    my $now = time;
    if (unix_from_mysql_datetime($seller->{last_bump}) + $BUMP_CHARGE_SECONDS > $now) {
        return $self->error_with_msg($c, HTTP_FORBIDDEN, 'Bump not allowed')
    }

    $dbh->query(
        'UPDATE `items` SET `created_at`=?, `updated_at`=? WHERE id=?',
        mysql_datetime_from_unix($now),
        mysql_datetime_from_unix($now),
        $target_item->{id},
    );

    $dbh->query(
        'UPDATE `users` SET `last_bump`=? WHERE id=?',
        mysql_datetime_from_unix($now),
        $seller->{id},
    );

    my $item = $dbh->select_row('SELECT * FROM `items` WHERE `id` = ?', $target_item->{id});

    $txn->commit();
    $c->render_json({
        item_id => $item->{id},
        item_price => $item->{price},
        item_created_at => unix_from_mysql_datetime($item->{created_at}),
        item_updated_at => unix_from_mysql_datetime($item->{updated_at}),
    });
};

# getSettings
get '/settings' => sub {
    my ($self, $c) = @_;

    my $response = {};
    my $user = $self->getUser($c);
    if ($user) {
        $response->{user} = {
            id => $user->{id},
            account_name => $user->{account_name},
            address => $user->{address},
            num_sell_items => $user->{num_sell_items},
        };
    }

    $response->{payment_service_url} = $self->getPaymentServiceURL();

    my $csrf_token = $self->getCSRFToken($c);
    $response->{csrf_token} = $csrf_token;

    my $categories = $self->dbh->select_all('SELECT * FROM `categories`');
    $response->{categories} = $categories;

    $c->render_json($response);
};

# postLogin
post '/login' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;
    my $account_name = $c->req->body_parameters->get('account_name') // "";
    my $password = $c->req->body_parameters->get('password') // "";

    if (!$account_name || !$password) {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST,'all parameters are required');
    }

    my $user = $self->dbh->select_row('SELECT * FROM `users` WHERE `account_name` = ?', $account_name);
    if (!$user) {
        return $self->error_with_msg($c, HTTP_UNAUTHORIZED, "アカウント名かパスワードが間違えています");
    }

    my $res = eval {
        check_password($password, $user->{hashed_password});
    };
    if ($@) {
        warn $@;
        return $self->error_with_msg($c, 500, 'crypt error');
    }
    if (!$res) {
        return $self->error_with_msg($c, 401, "アカウント名かパスワードが間違えています");
    }

    my $session = Plack::Session->new($c->env);
    $session->set('user_id' => $user->{id});
    $session->set('csrf_token' => secure_random_str(20));

    $c->render_json({
        id => $user->{id},
        account_name => $user->{account_name},
        address => $user->{address},
        num_sell_items => $user->{num_sell_items},
    });
};

# postRegister
post '/register' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;

    my $account_name = $c->req->body_parameters->get('account_name') // "";
    my $address = $c->req->body_parameters->get('address') // "";
    my $password = $c->req->body_parameters->get('password') // "";

    if ($account_name eq "" || $password eq "" || $address eq "") {
        return $self->error_with_msg($c, HTTP_BAD_REQUEST, 'all parameters are required');
    }

    my $hashed_password = encrypt_password($password);

    $self->dbh->query(
        'INSERT INTO `users` (`account_name`, `hashed_password`, `address`) VALUES (?, ?, ?)',
        $account_name,
        $hashed_password,
        $address,
    );
    my $user_id = $self->dbh->last_insert_id();

    my $session = Plack::Session->new($c->env);
    $session->set('user_id' => $user_id);
    $session->set('csrf_token' => secure_random_str(20));

    $c->render_json({
        id => $user_id,
        account_name => $account_name,
        Address => $address
    });
};

# getReports
get '/reports.json' => [qw/allow_json_request/] => sub {
    my ($self, $c) = @_;

    my $tes = $self->dbh->select_all(
        'SELECT * FROM `transaction_evidences` WHERE `id` > 15007'
    );

    my @transaction_evidences = ();
    for my $te (@$tes) {
        delete $te->{created_at};
        delete $te->{uppdated_at};
        push @transaction_evidences, $te;
    }

    $c->render_json(\@transaction_evidences);
};

1;
