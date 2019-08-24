package Isucari::API;

use strict;
use warnings;
use utf8;
use JSON::XS 3.00;
use JSON::Types;
use LWP::UserAgent;

our $ISUCARI_API_TOKEN = "Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0";

sub new {
    my $class = shift;
    my $ua  = LWP::UserAgent->new(
        agent => "isucon9-qualify-webapp",
    );
    bless {ua => $ua}, $class;
}

sub ua {
    my $self = shift;
    $self->{ua};
}

sub payment_token {
    my ($self,$payment_url,$param) = @_;
    my $json = JSON::encode_json($param);

    my $req = HTTP::Request->new(POST => $payment_url . "/token");
    $req->header("Content-Type", "application/json");
    $req->content($json);

    my $res = $self->ua->request($req);
    if ($res->code != 200) {
        my $msg = $res->code . ':' . $res->content;
        $msg =~ s/\n$//gms;
        die $msg;
    }

    return JSON::decode_json($res->content);
}

sub shipment_create {
    my ($self,$shipment_url,$param) = @_;
    my $json = JSON::encode_json($param);

    my $req = HTTP::Request->new(POST => $shipment_url . "/create");
    $req->header("Content-Type", "application/json");
    $req->header("Authorization", $ISUCARI_API_TOKEN);
    $req->content($json);

    my $res = $self->ua->request($req);
    if ($res->code != 200) {
        my $msg = $res->code . ':' . $res->content;
        $msg =~ s/\n$//gms;
        die $msg;
    }

    return JSON::decode_json($res->content);
}


sub shipment_request {
    my ($self,$shipment_url,$param) = @_;
    my $json = JSON::encode_json($param);

    my $req = HTTP::Request->new(POST => $shipment_url . "/request");
    $req->header("Content-Type", "application/json");
    $req->header("Authorization", $ISUCARI_API_TOKEN);
    $req->content($json);

    my $res = $self->ua->request($req);
    if ($res->code != 200) {
        my $msg = $res->code . ':' . $res->content;
        $msg =~ s/\n$//gms;
        die $msg;
    }

    return $res->content;
}

sub shipment_status {
    my ($self,$shipment_url,$param) = @_;
    my $json = JSON::encode_json($param);

    my $req = HTTP::Request->new(GET => $shipment_url . "/status");
    $req->header("Content-Type", "application/json");
    $req->header("Authorization", $ISUCARI_API_TOKEN);
    $req->content($json);

    my $res = $self->ua->request($req);
    if ($res->code != 200) {
        my $msg = $res->code . ':' . $res->content;
        $msg =~ s/\n$//gms;
        die $msg;
    }

    return JSON::decode_json($res->content);
}
