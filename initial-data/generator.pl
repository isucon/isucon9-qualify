#!/usr/bin/perl

use strict;
use warnings;
use utf8;
use Crypt::Eksblowfish::Bcrypt qw/bcrypt/;
use Crypt::OpenSSL::Random;
use Digest::SHA;
use JSON;
use JSON::Types;
use List::Util;

open(my $sql_fh, ">:utf8", "result/initial.sql") or die $!;
open(my $users_fh, ">", "result/users_json.txt") or die $!;
open(my $active_sellers_fh, ">", "result/active_sellers_json.txt") or die $!;
open(my $items_fh, ">", "result/items_json.txt") or die $!;
open(my $te_fh, ">", "result/transaction_evidences_json.txt") or die $!;
open(my $shippings_fh, ">", "result/shippings_json.txt") or die $!;

my $BASE_PRICE = 100;
my $NUM_USER_GENERATE = 4000;
my $NUM_ITEM_GENERATE = 50000;
my $RATE_OF_SOLDOUT = 31; # sold out商品の割合
my $RATE_OF_ACTIVE_SELLER = 10; # 出品が多いユーザの割合
my $RATE_OF_ACTIVE_SELLER_RATE = 90; # 出品が多いユーザに割り振る割合。90%の商品が10%のユーザから出品されている
my $CLAUSE_IN_DESCRIPTION = 100; # description中の文節の数

my $PASSWORD_SALT = 'Oi87WbXmCRnFZATUm4fXUJUE8VLdiI4tGk17M1K3SmS';
my @ADDTIONAL_ADDREDSS = qw/
青葉区
泉区
太白区
宮城野区
若林区
東区
白石区
厚別区
豊平区
清田区
南区
西区
手稲区
秋葉区
江南区
西蒲区
川崎区
幸区
中原区
高津区
宮前区
多摩区
麻生区
旭区
磯子区
神奈川区
金沢区
港南区
栄区
瀬谷区
都筑区
鶴見区
戸塚区
保土ケ谷区
緑区
伊洲根区
清水区
葵区
駿河区
浜北区
天竜区
熱田区
昭和区
千種区
天白区
中川区
中村区
瑞穂区
名東区
守山区
左京区
上京区
右京区
中京区
東山区
山科区
伏見区
西京区
下京区
中京区
大宮区
見沼区
桜区
浦和区
岩槻区
/;
my @CATEGOREIS_WEIGHT = (
[[2,1],1],
[[3,1],2],
[[4,1],1],
[[5,1],1],
[[6,1],1],
[[11,10],3],
[[12,10],2],
[[13,10],2],
[[14,10],3],
[[15,10],1],
[[21,20],3],
[[22,20],5],
[[23,20],3],
[[24,20],2],
[[31,30],4],
[[32,30],3],
[[33,30],2],
[[34,30],1],
[[35,30],1],
[[41,40],4],
[[42,40],2],
[[43,40],2],
[[44,40],2],
[[45,40],2],
[[51,50],1],
[[52,50],2],
[[53,50],1],
[[54,50],2],
[[55,50],1],
[[56,50],1],
[[61,60],3],
[[62,60],2],
[[63,60],1],
[[64,60],3],
[[65,60],3],
[[66,60],4]
);
my @CATEGOREIS=();
for my $cw (@CATEGOREIS_WEIGHT) {
    for (my $i=0;$i<$cw->[1];$i++) {
        push @CATEGOREIS, $cw->[0];
    }
}

sub format_mysql {
    my $time = shift;
    my @lt = localtime($time);
    sprintf("%04d-%02d-%02d %02d:%02d:%02d", $lt[5]+1900,$lt[4]+1,$lt[3],$lt[2],$lt[1],$lt[0]);
}

#  パスワードの生成
# account_name + salt の hmac + sha256 + base64
sub gen_passwd {
    my $id = shift;
    Digest::SHA::hmac_sha256_base64($id,$PASSWORD_SALT)
}

sub encrypt_password {
    my $password = shift;
    my $hex = Digest::SHA::hmac_sha256_hex($password);
    if (-f "pwcache/$hex.txt") {
        open(my $fh, "pwcache/$hex.txt");
        my $encrypted = do {local $/; <$fh>};
        chomp $encrypted;
        return $encrypted;
    }
    my $salt = shift || Crypt::Eksblowfish::Bcrypt::en_base64(Crypt::OpenSSL::Random::random_bytes(16));
    my $settings = '$2a$10$'.$salt;
    my $encrypted = Crypt::Eksblowfish::Bcrypt::bcrypt($password, $settings);
    open(my $fh, ">", "pwcache/$hex.txt") or die $!;
    print $fh $encrypted;
    return $encrypted;
}

# Check if the passwords match
sub check_password {
    my ($plain_password, $hashed_password) = @_;
    if ($hashed_password =~ m!^\$2a\$\d{2}\$([A-Za-z0-9+\\.]{22})!) {
        return encrypt_password($plain_password, $1) eq $hashed_password;
    } else {
        return;
    }
}

my %users = ();
my @active_seller = ();
sub create_user {
    my ($id, $name, $passwd, $address, $created_at, $category_id, $parent_category_id) = @_;
    # 出品が多いユーザ
    if (rand(100) < $RATE_OF_ACTIVE_SELLER) {
        push @active_seller, $id;
    }
    $users{$id} = +{
        id           => number $id,
        account_name => string $name,
        plain_passwd => string $passwd,
        address      => string $address,
        created_at   => number $created_at,
        num_sell_items => number 0,
        buy_category_id => number $category_id,
        buy_parent_category_id => number $parent_category_id,
        num_buy_items => number 0,
    };
}

sub flush_users {
    my @insert_users = ();
    for my $user (values %users) {
        delete $user->{buy_category_id};
        print $users_fh JSON::encode_json($user)."\n";
        push @insert_users,
            sprintf(q!(%d,'%s','%s','%s', %d,'%s')!,
            $user->{id},
            $user->{account_name},
            encrypt_password($user->{plain_passwd}),
            $user->{address},
            $user->{num_sell_items},
            format_mysql($user->{created_at})
        );
        if (@insert_users > 500) {
            print $sql_fh q!INSERT INTO `users` (`id`,`account_name`,`hashed_password`,`address`,`num_sell_items`,`created_at`) VALUES ! . join(", ", @insert_users) . ";\n";
            @insert_users = ();
        }
    }
    print $sql_fh q!INSERT INTO `users` (`id`,`account_name`,`hashed_password`,`address`,`num_sell_items`,`created_at`) VALUES ! . join(", ", @insert_users) . ";\n";

    for my $active_seller_id (@active_seller) {
        my $user = $users{$active_seller_id};
        print $active_sellers_fh JSON::encode_json({id => $user->{id}, num_sell_items => $user->{num_sell_items}})."\n";
    }
}

{
    print $sql_fh q!use `isucari`;!."\n\n";
}

{
    open(my $fh, "<:utf8", "users.tsv") or die $!;
    my @dummy_users = map { chomp $_; [ split /\t/, $_, 3] } <$fh>;

    # For demo
    create_user(1, 'isudemo1', 'isudemo1', '東京都港区6-11-1', 1565398800, 2,1);
    create_user(2, 'isudemo2', 'isudemo2', '東京都新宿区4-1-6', 1565398801, 11,10);
    create_user(3, 'isudemo3', 'isudemo3', '東京都伊洲根9-4000', 1565398802, 21,20);

    my $base_time = 1565398803; #2019-08-10 10:00:03
    srand(1565458009);
    for (my $i=4;$i<=$NUM_USER_GENERATE;$i++) {
        my $dummy_user = $dummy_users[$i];
        my $id = $dummy_user->[1];
        $id =~ s/@.+$//g;
        my $ad1 = int(rand(5))+1;
        my $ad2 = int(rand(50))+1;
        my $address = $dummy_user->[2] . $ADDTIONAL_ADDREDSS[$i % (scalar @ADDTIONAL_ADDREDSS)] . $ad1 . "-" . $ad2;
        $users{$i} = [$id,$address];
        my $category = $CATEGOREIS[int(rand(scalar @CATEGOREIS))];
        create_user(
            $i,
            $id,
            gen_passwd($id),
            $address,
            $base_time+$i,
            $category->[0],
            $category->[1]
        );
    }
}

open(my $md5fh, "<:utf8", "image_files.txt") or die $!;
my @IMAGES = map { chomp $_; $_ } <$md5fh>;

open(my $fh, "<:utf8", "keywords.tsv") or die $!;
my @KEYWORDS = map { chomp $_; $_ } <$fh>;

sub gen_text {
    my ($length, $return) = @_;
    my @text;
    for (my $i=0;$i<$length;$i++) {
        my $r = int(rand(scalar @KEYWORDS));
        my $t = $KEYWORDS[$r];
        chomp($t);
        if ($t eq "#") {
            $t = "\n" if $return;
            $t = " " if !$return;
        }
        push @text, $t;
    }
    my $text = join "", @text;
    $text =~ s/^(\s|\n)+//gs;
    return $text;
}

my @insert_items;
my @insert_te;
my @insert_shippings;
sub insert_items {
    my ($id, $seller_id, $buyer_id, $status, $name, $price, $description, $image_name, $category_id, $created_at, $updated_at) = @_;

    print $items_fh JSON::encode_json({
        id        => number $id,
        seller_id => number $seller_id,
        buyer_id  => number $buyer_id,
        status    => string $status,
        name      => string $name,
        price     => number $price,
        description => string $description,
        image_name  => string $image_name,
        category_id => number $category_id,
        created_at  => number $created_at,
        updated_at  => number $updated_at
    })."\n";

    $description =~ s/\n/\\n/g;
    push @insert_items, sprintf(q!(%d, %d, %d, '%s', '%s', %d, '%s', '%s', %d, '%s', '%s')!, $id, $seller_id, $buyer_id, $status, $name, $price, $description, $image_name, $category_id, format_mysql($created_at), format_mysql($updated_at));

    $users{$seller_id}->{num_sell_items}++;

    if (@insert_items > 500) {
        flush_items();
    }

}
sub flush_items {
    print $sql_fh q!INSERT INTO `items` (`id`,`seller_id`,`buyer_id`,`status`,`name`,`price`,`description`,`image_name`,`category_id`,`created_at`,`updated_at`) VALUES ! . join(", ", @insert_items) . ";\n";
    @insert_items = ();
    flush_te();
    flush_shippings();
}

sub insert_te {
    my ($id, $seller_id, $buyer_id, $status, $item_id, $item_name, $item_price, $item_description, $item_category_id, $item_root_category_id, $created_at, $updated_at) = @_;

    print $te_fh JSON::encode_json({
        id         => number $id,
        seller_id  => number $seller_id,
        buyer_id   => number $buyer_id,
        status     => string $status,
        item_id    => number $item_id,
        item_name  => string $item_name,
        item_price => number $item_price,
        item_description      => string $item_description,
        item_category_id      => number $item_category_id,
        item_root_category_id => number $item_root_category_id,
        created_at            => number $created_at,
        updated_at            => number $updated_at
    })."\n";

    push @insert_te, sprintf(q!(%d, %d, %d, '%s', %d, '%s', %d, '%s', %d, %d, '%s', '%s')!, $id, $seller_id, $buyer_id, $status, $item_id, $item_name, $item_price, $item_description, $item_category_id, $item_root_category_id, format_mysql($created_at), format_mysql($updated_at));
}

sub flush_te {
    return unless @insert_te;
    print $sql_fh q!INSERT INTO `transaction_evidences` (`id`,`seller_id`,`buyer_id`,`status`,`item_id`,`item_name`,`item_price`,`item_description`,`item_category_id`,`item_root_category_id`,`created_at`,`updated_at`) VALUES ! . join(", ", @insert_te) . ";\n";
    @insert_te = ();
}

sub insert_shippings {
    my ($transaction_evidence_id, $status, $item_name, $item_id, $reserve_id, $reserve_time, $to_address, $to_name, $from_address, $from_name, $img_binary, $created_at, $updated_at) = @_;
    {
        print $shippings_fh JSON::encode_json({
            transaction_evidence_id => number $transaction_evidence_id,
            status => string $status,
            item_name => string $item_name,
            item_id => number $item_id,
            reserve_id => string $reserve_id,
            reserve_time => number $reserve_time,
            to_address => string $to_address,
            to_name => string $to_name,
            from_address => string $from_address,
            from_name => string $from_name,
            img_binary => string "", # delete from JSON
            created_at => number $created_at,
            updated_at => number $updated_at,
        })."\n";
    }
    {
        push @insert_shippings, sprintf(q!(%d, '%s', '%s', %d, '%s', %d, '%s', '%s', '%s', '%s', '%s', '%s', '%s')!, $transaction_evidence_id, $status, $item_name, $item_id, $reserve_id, $reserve_time, $to_address, $to_name, $from_address, $from_name, $img_binary, format_mysql($created_at), format_mysql($updated_at));
    }
}
sub flush_shippings {
    return unless @insert_shippings;
    print $sql_fh q!INSERT INTO `shippings` (`transaction_evidence_id`,`status`,`item_name`,`item_id`,`reserve_id`,`reserve_time`,`to_address`,`to_name`,`from_address`,`from_name`,`img_binary`,`created_at`,`updated_at`) VALUES ! . join(", ", @insert_shippings) . ";\n";
    @insert_shippings = ();
}

{
    my $base_time = 1565575207; #2019-08-12 11:00:07
    srand(1565358009);

    my $te_id = 0;
    my $active_seller_rr = 0;
    my $images_rr = 0;
    my $buyer_rr = 0;
    my @buyers = (1..$NUM_USER_GENERATE);
    @buyers = List::Util::shuffle @buyers;
    for (my $i=1;$i<=$NUM_ITEM_GENERATE;$i++) {
        my $t_sell = $base_time+int(rand(10))-5;
        my $t_buy = $t_sell + int(rand(10)) + 60;
        my $t_done = $t_buy + 10;

        my $name = gen_text(8,0),;
        my $description = gen_text($CLAUSE_IN_DESCRIPTION,1);
        my $category = $CATEGOREIS[int(rand(scalar @CATEGOREIS))];

        my $seller = int(rand($NUM_USER_GENERATE))+1;
        if (rand(100) < $RATE_OF_ACTIVE_SELLER_RATE) {
            $seller = $active_seller[$active_seller_rr % scalar @active_seller];
            $active_seller_rr++;
        }
        my $status = 'on_sale';
        my $buyer = 0;

        if (rand(100) < $RATE_OF_SOLDOUT && $te_id < 15007) { # TODO
            $status = 'sold_out';
            $te_id++;
            $buyer = $buyers[$buyer_rr % scalar @buyers];
            $buyer_rr++;
            while ($buyer == $seller) {
                $buyer = int(rand($NUM_USER_GENERATE))+1;
            }

            #buyerが決まっている場合、buyerのcategoryを使う
            $category = [
                $users{$buyer}->{buy_category_id},
                $users{$buyer}->{buy_parent_category_id}
            ];

            $users{$buyer}->{num_buy_items}++;

            insert_te(
                $te_id,
                $seller,
                $buyer,
                'done',
                $i,
                $name,
                $BASE_PRICE,
                $description,
                $category->[0],
                $category->[1],
                $t_buy,
                $t_done
            );

            insert_shippings(
                $te_id,
                'done',
                $name,
                $i,
                sprintf("%010d", int(rand(10000000000))),
                $t_buy,
                $users{$buyer}->{address},
                $users{$buyer}->{account_name},
                $users{$seller}->{address},
                $users{$seller}->{account_name},
                "", # img_binary null ok
                $t_buy,
                $t_done
            );
        }

        insert_items(
            $i,
            $seller,
            $buyer,
            $status,
            $name,
            $BASE_PRICE,
            $description,
            $IMAGES[$images_rr % scalar @IMAGES],
            $category->[0],
            $t_sell,
            $t_done
        );
        $images_rr++;
        $base_time++;
    }
}

flush_items();
flush_users();
