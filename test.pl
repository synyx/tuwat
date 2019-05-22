#!/usr/bin/perl
use strict;
use warnings;

use Data::Dumper;

use Encode qw/decode/;
use JSON; #libjson-perl
use LWP;  # sollte: libwww-perl
use utf8;
use Sys::Syslog; #libsys-syslog-perl
use LockFile::Simple qw(lock trylock unlock);# liblockfile-simple-perl

# Simple locking using default settings
lock("/var/lock/test.pl.pid") || die "can't lock /var/lock/test.pl.pid\n";
warn "already locked\n" unless trylock("/var/lock/test.pl.pid");

my $redmine_user	= "AdminBot";
my $redmine_key 	= "20eb22c240dad7b38e6a1b9721d9292ab96a61f5";

my $contargo_mail = 'ferstl@synyx.de';

 (my $sec,my $min,my $hour,my $mday,my $mon,my $year,my $wday,my $yday,my $isdst) = localtime();

								
#								my $dat->{"issue"}->{project_id}= "a";
#								$dat->{"issue"}->{subject}= $subject_message;
#								$dat->{"issue"}->{priority}= 5;
#								$dat->{"issue"}->{description}= $body;
#								$dat->{"issue"}->{notes}= "notes";
#								$dat->{"issue"}->{custom_field_values}->{3} = $standort;

#								my $json_text = JSON->new->encode($dat);
#								print $json_text."\r\n";
							
#								my $uri = "https://project.synyx.de/issues.json?project_id=a&status_id=2";
                my $uri = "https://project.synyx.de/issues.json?project_id=a&status_id=2&limit=100&sort=updated_on:desc";               
								my $req = HTTP::Request->new( 'GET', $uri );
								$req->header( 'Content-Type' => 'application/json' );
#								$req->header( 'X-ChiliProject-API-Key' => $redmine_key );
								$req->header( 'X-Redmine-API-Key' => $redmine_key );
#								$req->content( $json_text );

								my $lwp = LWP::UserAgent->new;
								my $ret = $lwp->request( $req );


								print Dumper($ret);
								print "\r\n\r".$ret->{_content};

open(my $fh, '>', 'progress.json');
print $fh $ret->{_content};
close $fh;

$mon = $mon +1;
$year = $year+1900;

my $today = "$year-$mon-$mday";

#                my $uri = "https://project.synyx.de/issues.json?project_id=a&status_id=5&updated_on=%3E%3D$today";               
                
#                my $uri = "https://project.synyx.de/issues.json?project_id=a&updated_on=%3E%3D$today";               
                #
                my $uri = "https://project.synyx.de/issues.json?project_id=a&status_id=5&sort=updated_on:desc&limit=10";               
								my $req = HTTP::Request->new( 'GET', $uri );
								$req->header( 'Content-Type' => 'application/json' );
#								$req->header( 'X-ChiliProject-API-Key' => $redmine_key );
								$req->header( 'X-Redmine-API-Key' => $redmine_key );
#								$req->content( $json_text );

								my $lwp = LWP::UserAgent->new;
								my $ret = $lwp->request( $req );


								print Dumper($ret);
								print "\r\n\r".$ret->{_content};

open(my $fh, '>', 'done.json');
print $fh $ret->{_content};
close $fh;
unlock("/var/lock/test.pl.pid");
