#!/usr/bin/perl -w

use strict;
use warnings FATAL => 'all';

use v5.14;

use JSON::XS qw(decode_json);
use Data::Dumper;

my $dir;
BEGIN {
        use Cwd qw(realpath);
        use File::Basename qw(dirname);
        $dir = dirname(realpath($0))
}
use lib "$dir/../lib";
use CheckBC qw(CheckBC);

unless (scalar(@ARGV) == 2) {
        help();
        exit;
}

my $oldFile = $ARGV[0];
my $newFile = $ARGV[1];
if ($oldFile eq $newFile) {
        die("The same file is given for both old and new interfaces. Died.\n")
}

foreach my $file ($oldFile, $newFile) {
        unless ($file eq '-' || -f $file) {
                die("File '$file' doesn't exist. Died.")
        }
}

my $old = decode_json(readFile($oldFile));
my $new = decode_json(readFile($newFile));

my @errors = CheckBC($old, $new);
if (@errors) {
        say join "\n", @errors;
        exit(1);
}

exit 0;

sub help {
print <<HELP
Compares two service interfaces for backward compability.

Usage:
\t$0 <old> <new>

To generate interface from service use flag --print-interface. I. e.
\tgo run example/main.go -config=example/app.ini --print-interface

HELP
}

sub readFile {
        my ($file) = @_;

        open(FILE, $file) || die "Can't open file '$file'";

        local $/ = undef;
        my $content = <FILE>;

        close(FILE);

        return $content;
}
