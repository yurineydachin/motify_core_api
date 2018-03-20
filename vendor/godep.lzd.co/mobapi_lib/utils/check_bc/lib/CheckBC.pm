package CheckBC;

require Exporter;
@ISA = qw(Exporter);
@EXPORT_OK = qw(CheckBC);

use strict;
use warnings FATAL => 'all';

use constant TRUE => !!1;
use constant FALSE => !!0;

use Data::Dumper qw(Dumper);

sub CheckBC {
        my ($old, $new) = @_;

        my @errors;

        my $interfaces = {
                old => $old,
                new => $new,
        };
        foreach my $interfaceName (sort keys(%$interfaces)) {
                my $interface = $interfaces->{$interfaceName};
                foreach my $field (qw(produces consumes definitions basePath paths)) {
                        unless (exists($interface->{$field})) {
                                push @errors, "$interfaceName: No field '$field'.";
                        }
                }
        }

        my %streams = (
                produces => 'produce',
                consumes => 'consume',
        );

        foreach my $stream (keys(%streams)) {
                if (exists($old->{$stream}) && exists($new->{$stream})) {
                        foreach my $missed (array_difference($old->{$stream}, $new->{$stream})) {
                                push @errors, "new: Interface must $streams{$stream} '$missed'.";
                        }
                }
        }

        if (exists($old->{basePath}) && exists($new->{basePath}) && $old->{basePath} ne $new->{basePath}) {
                push @errors, "new: basePath must be '$old->{basePath}'.";
        }

        if (exists($old->{paths}) && exists($new->{paths})) {
                foreach my $missed (array_difference([keys($old->{paths})], [keys($new->{paths})])) {
                        push @errors, "new: Lost path '$missed'.";
                }

                foreach my $path (keys($old->{paths})) {
                        next unless exists($new->{paths}{$path});

                        foreach my $missed (array_difference([keys($old->{paths}{$path})], [keys($new->{paths}{$path})])) {
                                push @errors, "new: Lost method '$missed' for path '$path'.";
                        }

                        foreach my $method (keys($old->{paths}{$path})) {
                                next unless exists($new->{paths}{$path}{$method});

                                my $oldMethod = $old->{paths}{$path}{$method};
                                my $newMethod = $new->{paths}{$path}{$method};

                                foreach my $missed (array_difference($oldMethod->{produces}, $newMethod->{produces})) {
                                        push @errors, "new: Lost format '$missed' for path '$path' method '$method'.";
                                }

                                foreach my $missed (array_difference([keys($old->{paths}{$path}{$method}{responses})], [keys($new->{paths}{$path}{$method}{responses})])) {
                                        push @errors, "new: Lost code '$missed' for path '$path' method '$method'.";
                                }

                                foreach my $code (keys($oldMethod->{responses})) {
                                        next unless exists($newMethod->{responses}{$code});

                                        my @answerErrors = checkTypes($oldMethod->{responses}{$code}{schema}, $newMethod->{responses}{$code}{schema}, $old, $new);
                                        my $errorText = "new: Changed response type for path '$path' method '$method' code '$code'.";
                                        foreach my $error (@answerErrors) {
                                                push @errors, "$errorText $error";
                                        }
                                }

                                {
                                        my @oldErrorCodes = $oldMethod->{description} =~ m'<li>Code: "<code>(.+?)</code>'g;
                                        my @newErrorCodes = $newMethod->{description} =~ m'<li>Code: "<code>(.+?)</code>'g;
                                        foreach my $code (array_difference(\@newErrorCodes, \@oldErrorCodes)) {
                                                push @errors, "new: Added error '$code' for path '$path' method '$method'. New errors aren't allowed."
                                        }
                                }

                                {
                                        my %oldParameters = map {$_->{name} => $_} @{$oldMethod->{parameters}};
                                        my %newParameters = map {$_->{name} => $_} @{$newMethod->{parameters}};
                                        foreach my $parameter (keys(%oldParameters)) {
                                                next unless exists($newParameters{$parameter});

                                                my $oldParameter = $oldParameters{$parameter};
                                                my $newParameter = $newParameters{$parameter};
                                                if ($oldParameter->{in} ne $newParameter->{in}) {
                                                        push @errors, "new: Changed parameter '$parameter' 'in' value for path '$path' method '$method'. Must be '$oldParameter->{in}'.";
                                                }
                                                my @parameterErrors = checkTypes($oldParameter, $newParameter, $old, $new);
                                                my $errorText = "new: Changed parameter '$parameter' type for path '$path' method '$method'.";
                                                foreach my $error (@parameterErrors) {
                                                        push @errors, "$errorText $error";
                                                }
                                        }

                                        foreach my $parameter (array_difference([keys(%newParameters)], [keys(%oldParameters)])) {
                                                my $newParameter = $newParameters{$parameter};
                                                if ($newParameter->{required}) {
                                                        push @errors, "new: Added required parameter '$newParameter->{name}' for path '$path' method '$method'. New required fields aren't allowed.";
                                                }
                                        }

                                        foreach my $parameter (array_difference([keys(%oldParameters)], [keys(%newParameters)])) {
                                                my $oldParameter = $oldParameters{$parameter};
                                                if ($oldParameter->{required}) {
                                                        push @errors, "new: Removed parameter '$oldParameter->{name}' for path '$path' method '$method'. Parameters musn't be removed.";
                                                }
                                        }
                                }
                        }
                }
        }

        return @errors;
}

sub checkTypes {
        my ($oldAnswer, $newAnswer, $old, $new, $typePrefix) = @_;
        $typePrefix //= '';

        my ($oldType, $oldSubAnswer) = getType($oldAnswer, $old);
        my ($newType, $newSubAnswer) = getType($newAnswer, $new);

        my @errors;
        if ($oldType ne $newType) {
                if ($typePrefix) {
                        push @errors, "'$typePrefix' must be '$oldType'.";
                } else {
                        push @errors, "Must be '$oldType'.";
                }
        } else {
                $typePrefix &&= "$typePrefix.";
                if (exists($oldSubAnswer->{type}) && $oldSubAnswer->{type} eq 'object') {
                        foreach my $missed (array_difference($oldSubAnswer->{required}, $newSubAnswer->{required})) {
                                my $fieldPath = getFieldPath($typePrefix, $oldType, $missed);
                                push @errors, "'$fieldPath' must be required.";
                        }
                        foreach my $field (keys($oldSubAnswer->{properties})) {
                                my $fieldPath = getFieldPath($typePrefix, $oldType, $field);
                                if (exists($newSubAnswer->{properties}{$field})) {
                                        my @fieldErrors = checkTypes($oldSubAnswer->{properties}{$field}, $newSubAnswer->{properties}{$field}, $old, $new, $fieldPath);
                                        push @errors, @fieldErrors;
                                } else {
                                        push @errors, "'$fieldPath' is lost.";
                                }
                        }
                }
        }

        return @errors;
}

sub getType {
        my ($answer, $interface) = @_;

        if (exists($answer->{'$ref'})) {
                my $ref = $answer->{'$ref'};
                my @path = split('/', $ref);
                @path = @path[1, $#path];
                my $value = $interface;
                for my $step (@path) {
                        $value = $value->{$step};
                }
                return getType($value, $interface);
        }

        my $type = $answer->{type};
        if ($type eq 'array') {
                return arrayType(getType($answer->{items}, $interface));
        }

        return $type, $answer;
}

sub getFieldPath {
        my ($typePrefix, $type, $field) = @_;
        return "$typePrefix$type.{field:$field}";
}

sub arrayType {
        my ($type, $answer) = @_;

        return "array[$type]", $answer;
}

sub objectType {
        my ($type, $answer) = @_;

        return "object{$type}", $answer;
}

sub array_difference {
        my ($minuend, $subtrahend) = @_;

        my %subtrahend = map {$_ => TRUE} @$subtrahend;
        my @difference = grep {!$subtrahend{$_}} @$minuend;

        return @difference;
}

1;
