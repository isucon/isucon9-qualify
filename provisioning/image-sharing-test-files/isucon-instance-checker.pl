#!/usr/bin/perl

use strict;
use warnings;
use JSON::PP;
use Test::More;

sub cmd_json {
    my @cmd_args = @_;
    my @cmd = qw!/usr/local/bin/aliyun --mode EcsRamRole --region=ap-northeast1!;
    push @cmd, @cmd_args;
    open(my $pipe, "-|", @cmd) or die $!;
    my $buffer="";
    while (<$pipe>) {
        $buffer .= $_;
    }
    close($pipe) or die $!;
    JSON::PP::decode_json($buffer);
}

# AccountId
my $identity = eval {
    cmd_json('sts','GetCallerIdentity');
};
die "Failed to retrieve Instance information. RAM role is not setup correctly: $@" if $@;

printf("AccountId: %s\n", $identity->{AccountId});

# Instances
my $instaces = cmd_json(qw/ecs DescribeInstances --RegionId ap-northeast-1/);
my @instaces = @{$instaces->{Instances}->{Instance}};

for my $instance (@instaces) {
    next if $instance->{Status} ne "Running";
    
    subtest sprintf("InstanceId: %s\n",$instance->{InstanceId}) => sub {
        my $disks = cmd_json(qw/ecs DescribeDisks --RegionId ap-northeast-1 --InstanceId/,$instance->{InstanceId});
        my @disks = @{$disks->{Disks}->{Disk}};

        is($instance->{InstanceChargeType},'PostPaid','InstanceChargeType should be PostPaid');
        is($instance->{ZoneId},'ap-northeast-1a','ZoneId should be ap-northeast-1a');
        is($instance->{InstanceType},'ecs.sn1ne.large','InstanceType should be ecs.sn1ne.large');
        is($instance->{Cpu},'2','Cpu should be 2 vCPU');
        is($instance->{Memory},'4096','Memory should be 4096 MB');
        is($instance->{InternetChargeType},'PayByTraffic','InternetChargeType should be PayByTraffic');
        is($instance->{InternetMaxBandwidthOut},'100','InternetMaxBandwidthOut should be 100');

        is(scalar @disks, 1, 'number of Disks should be 1');

        for my $disk (@disks) {
            is($disk->{Type}, 'system', 'Disk Type should be system');
            is($disk->{Size}, '40', 'Disk Size should be 40 GiB');
            is($disk->{Category}, 'cloud_efficiency', 'Disk Category should be cloud_efficiency(Ultra Disk)');
        }
    }
}

done_testing();
