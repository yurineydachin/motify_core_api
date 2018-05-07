#!/usr/bin/perl -w

use Test::Most;
use Clone qw(clone);
use Data::Dumper qw(Dumper);

use FindBin qw($Bin);
use lib "$Bin/../lib";
use CheckBC qw(CheckBC);

use JSON;

die_on_fail();

require_ok("CheckBC");

my %methodTemplate = (description => '', responses => {'200' => {'schema' => {type => 'string'}}});

subtest('empty interfaces' => sub {
        my $old = {};
        my $new = {};

        my @errors = CheckBC($old, $new);
        my @expectedErrors;
        foreach my $interface (qw(new old)) {
                foreach my $field (qw(produces consumes definitions basePath paths)) {
                        push @expectedErrors, "$interface: No field '$field'.";
                }
        }
        is_deeply(\@errors, \@expectedErrors, 'Empty interfaces')
});

subtest("streams" => sub {
        my %streams = (
                produces => 'produce',
                consumes => 'consume',
        );
        foreach my $stream (keys(%streams)) {
                subtest("new interface must $streams{$stream} the all formats of old one' " => sub {
                        my $template = {$stream => ['application/json']};
                        my $addition = 'application/xml';

                        subtest('the same value' => sub {
                                my $old = clone($template);
                                my $new = clone($template);

                                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                                is_deeply(\@errors, [], 'no specific errors')
                        });

                        subtest('more values in new' => sub {
                                my $old = clone($template);
                                my $new = clone($template);

                                push $new->{$stream}, $addition;

                                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                                is_deeply(\@errors, [], 'no specific errors')
                        });

                        subtest('more values in old' => sub {
                                my $old = clone($template);
                                my $new = clone($template);

                                push $old->{$stream}, $addition;

                                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                                is_deeply(\@errors, ["new: Interface must $streams{$stream} '$addition'."], "Output format '$addition' is lost");
                        });
                });
        }
});

subtest("basePath must be the same" => sub {
        my $first = '/';
        my $second = '/another/';
        my $template = {basePath => $first};

        subtest('the same value' => sub {
                my $old = clone($template);
                my $new = clone($template);

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('different values' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{basePath} = $second;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: basePath must be '$first'."], 'must be the same')
        });
});

subtest("paths must include the all old ones" => sub {
        my $content = {};
        my $template = { paths => { '/hello/world/v1/' => $content, }, };
        my $addition = '/search/google/v1/';

        subtest('the same value' => sub {
                my $old = clone($template);
                my $new = clone($template);

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('more values in new' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{paths}{$addition} = $content;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('more values in old' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $old->{paths}{$addition} = $content;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Lost path '$addition'."], "Path '$addition' is lost");
        });
});

subtest("path methods must include the all old ones" => sub {
        my $content = {%methodTemplate};
        my $path = '/hello/world/v1/';
        my $template = {paths => {$path => {get => $content}}};
        my $addition = 'post';

        subtest('the same value' => sub {
                my $old = clone($template);
                my $new = clone($template);

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('more values in new' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{paths}{$path}{$addition} = $content;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('more values in old' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $old->{paths}{$path}{$addition} = $content;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Lost method '$addition' for path '$path'."], "Path '$path' method '$addition' is lost");
        });
});

subtest("path method must produce the all formats old one does" => sub {
        my $path = '/hello/world/v1/';
        my $method = 'get';
        my $template = {paths => {$path => {$method => {produces => ['application/json'], %methodTemplate}}}};
        my $addition = 'application/xml';

        subtest('the same value' => sub {
                my $old = clone($template);
                my $new = clone($template);

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('more values in new' => sub {
                my $old = clone($template);
                my $new = clone($template);

                push $new->{paths}{$path}{$method}{produces}, $addition;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('more values in old' => sub {
                my $old = clone($template);
                my $new = clone($template);

                push $old->{paths}{$path}{$method}{produces}, $addition;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Lost format '$addition' for path '$path' method '$method'."], "Path '$path' method '$method' format '$addition' is lost");
        });
});

subtest("path method musn't have new errors" => sub {
        my $path = '/search/google/v1/';
        my $method = 'get';
        my $content = 'Search things in Lazada using Google';
        my $errorsIntro = "<br>Handler can return these error messages:\n";
        my $firstCode = 'EMPTY_RESULT';
        my $secondCode = 'INVALID_PARAMETER';
        my $first = '<li>Code: "<code>'.$firstCode.'</code>", Data: "<code>Nothing was found</code>"</li>"';
        my $second = '<li>Code: "<code>'.$secondCode.'</code>", Data: "<code>Invalid parameter got</code>"</li>"';
        my $template = {paths => {$path => {$method => {
                %methodTemplate,
                'description' => $content.$errorsIntro.'<ul>'.$first.'</ul>',
        }}}};

        subtest('the same value' => sub {
                my $old = clone($template);
                my $new = clone($template);

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest("errors addition is prohibited" => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{paths}{$path}{$method}{description} = $content.$errorsIntro.'<ul>'.$first.$second.'</ul>';

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Added error '$secondCode' for path '$path' method '$method'. New errors aren't allowed."], "Path '$path' method '$method' error '$secondCode' is added");
        });

        subtest('errors removal is allowed' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{paths}{$path}{$method}{description} = $content;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });
});

subtest("path method must response the all codes old one does" => sub {
        my $content = {};
        my $path = '/hello/world/v1/';
        my $method = 'get';
        my $template = {paths => {$path => {$method => {%methodTemplate, responses => {'200' => $methodTemplate{responses}{'200'}}}}}};
        my $addition = '302';

        subtest('the same value' => sub {
                my $old = clone($template);
                my $new = clone($template);

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('more values in new' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{paths}{$path}{$method}{responses}{$addition} = $content;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('more values in old' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $old->{paths}{$path}{$method}{responses}{$addition} = $content;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Lost code '$addition' for path '$path' method '$method'."], "Path '$path' method '$method' code '$addition' is lost");
        });
});

subtest("path method must keep backward compability response" => sub {
        subtest("responses string" => sub {
                my $content = 'string';
                my $path = '/hello/world/v1/';
                my $method = 'get';
                my $code = '200';
                my $template = {paths => {$path => {$method => {%methodTemplate, responses => {$code => {schema => { 'type' => $content}}}}}}};
                my $change = 'int';

                subtest('the same value' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, [], 'no specific errors')
                });

                subtest('another value in new' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        $new->{paths}{$path}{$method}{responses}{$code}{schema}{type} = $change;

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, ["new: Changed response type for path '$path' method '$method' code '$code'. Must be '$content'."], "Path '$path' method '$method' code '$code' method '$content' is changed");
                });
        });

        subtest("responses array[array[string]]" => sub {
                my $content = {
                        'type' => 'array',
                        'items' => {
                                'type' => 'array',
                                'items' => {
                                        type => 'string',
                                }
                        },
                };
                my $path = '/hello/world/v1/';
                my $method = 'get';
                my $code = '200';
                my $template = {paths => {$path => {$method => {%methodTemplate, responses => {$code => {schema => $content}}}}}};
                my $change = {
                        'type' => 'array',
                        'items' => {
                                type => 'string',
                        },
                };

                subtest('the same value' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, [], 'no specific errors')
                });

                subtest('another value in new' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        $new->{paths}{$path}{$method}{responses}{$code}{schema} = $change;

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, ["new: Changed response type for path '$path' method '$method' code '$code'. Must be 'array[array[string]]'."], "Path '$path' method '$method' code '$code' response is changed");
                });
        });

        subtest("responses array[google.V1Res]" => sub {
                my $content = {
                        'required' => [
                                'title',
                                'url'
                        ],
                        'type' => 'object',
                        'properties' => {
                                'url' => {
                                        'type' => 'string'
                                },
                                'title' => {
                                        'type' => 'string'
                                }
                        }
                };
                my $path = '/hello/world/v1/';
                my $method = 'get';
                my $code = '200';
                my $template = {
                        paths => {$path => {$method => {%methodTemplate, responses => {$code => { schema => {
                                'type' => 'array',
                                'items' => {
                                        required => [qw(reference)],
                                        type => 'object',
                                        properties => {
                                                reference => {
                                                        '$ref' => '#/definitions/google.V1Res'
                                                }
                                        }
                                }
                        }}}}}},
                        'definitions' => { 'google.V1Res' => $content }
                };
                my $change = {
                        'type' => 'array',
                        'items' => {
                                type => 'string',
                        },
                };

                subtest('the same value' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, [], 'no specific errors')
                });

                subtest('ref is renamed' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        $new->{definitions} = {'google.V2Res' => $content};
                        $new->{paths}{$path}{$method}{responses}{$code}{schema}{items}{properties}{reference}{'$ref'} = '#/definitions/google.V2Res';

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, [], 'no specific errors')
                });

                subtest("Ref object field type musn't be changed" => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        $new->{definitions}{'google.V1Res'}{properties}{title}{type} = 'int';

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, ["new: Changed response type for path '$path' method '$method' code '$code'. 'array[object].{field:reference}.object.{field:title}' must be 'string'."], "Path '$path' method '$method' code '$code' response is changed");
                });

                subtest('More fields' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        $new->{definitions}{'google.V1Res'}{properties}{newfield}{type} = 'int';

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, [], 'no specific errors')
                });

                subtest('More required fields' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        $new->{definitions}{'google.V1Res'}{properties}{newfield}{type} = 'int';
                        push($new->{definitions}{'google.V1Res'}{required}, 'newfield');

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, [], 'no specific errors')
                });

                subtest('All old fields must be in the new interface' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        $old->{definitions}{'google.V1Res'}{properties}{newfield}{type} = 'int';

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, ["new: Changed response type for path '$path' method '$method' code '$code'. 'array[object].{field:reference}.object.{field:newfield}' is lost."], "Path '$path' method '$method' code '$code' response is changed");
                });

                subtest('All old required fields must be required in the new interface' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        my $optional_field = shift($new->{definitions}{'google.V1Res'}{required});

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, ["new: Changed response type for path '$path' method '$method' code '$code'. 'array[object].{field:reference}.object.{field:$optional_field}' must be required."], "Path '$path' method '$method' code '$code' response is changed");
                });

                subtest('Changed type in ref' => sub {
                        my $old = clone($template);
                        my $new = clone($template);

                        $new->{definitions}{'google.V1Res'} = {type => 'string'};

                        my @errors = filterNoFieldErrors(CheckBC($old, $new));
                        is_deeply(\@errors, ["new: Changed response type for path '$path' method '$method' code '$code'. 'array[object].{field:reference}' must be 'object'."], "Path '$path' method '$method' code '$code' response is changed");
                });
        });
});

subtest("path method parameters must keep backward compability" => sub {
        my $path = '/hello/world/v1/';
        my $method = 'get';
        my $parameter = 'query';

        my $content = {
                'required' => JSON::true,
                'in' => 'query',
                'name' => $parameter,
                'type' => 'string',
                'description' => 'Search query'
        };

        my $template = {paths => {$path => {$method => {%methodTemplate, parameters => [ $content ]}}}};

        my %addition = (
                'in' => 'query',
                'name' => 'lang',
                'type' => 'string',
                'description' => 'Search query'
        );

        subtest('the same value' => sub {
                my $old = clone($template);
                my $new = clone($template);

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest("parameter musn't be removed" => sub {
                my $old = clone($template);
                my $new = clone($template);

                my $removed = shift $new->{paths}{$path}{$method}{parameters};

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Removed parameter '$parameter' for path '$path' method '$method'. Parameters musn't be removed."], "Path '$path' method '$method' parameter is removed");
        });

        subtest('parameter can become optional' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{paths}{$path}{$method}{parameters}[0]{required} = JSON::false;

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest("parameter 'in' value musnt be changed" => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{paths}{$path}{$method}{parameters}[0]{in} = 'form';

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Changed parameter '$parameter' 'in' value for path '$path' method '$method'. Must be '$content->{in}'."], "Path '$path' method '$method' parameter 'in' value is changed");
        });

        subtest('parameter type musnt be changed' => sub {
                my $old = clone($template);
                my $new = clone($template);

                $new->{paths}{$path}{$method}{parameters}[0]{type} = 'int';

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Changed parameter '$parameter' type for path '$path' method '$method'. Must be '$content->{type}'."], "Path '$path' method '$method' parameter type is changed");
        });

        subtest('new optional parameters are allowed' => sub {
                my $old = clone($template);
                my $new = clone($template);

                push $new->{paths}{$path}{$method}{parameters}, {%addition, required => JSON::false};

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, [], 'no specific errors')
        });

        subtest('no new required parameters are allowed' => sub {
                my $old = clone($template);
                my $new = clone($template);

                push $new->{paths}{$path}{$method}{parameters}, {%addition, required => JSON::true};

                my @errors = filterNoFieldErrors(CheckBC($old, $new));
                is_deeply(\@errors, ["new: Added required parameter '$addition{name}' for path '$path' method '$method'. New required fields aren't allowed."], "Path '$path' method '$method' required parameter is added");
        });
});

done_testing();

sub filterNoFieldErrors {
        return grep {!/^\w+: No field '\w+'./} @_;
}

__DATA__

Interface example:

{
        'info' => {
                'version' => '1.0.0',
                'title' => 'HTTP JSON RPC for Go',
                'description' => '<h2>Description</h2>
                <p>HTTPS RPC server.</p>
                <h2>Protocol</h2>
                <p>It supports "GET" or "POST" methods for requests and returns a JSON in response.</p>
                <h3>Response</h3>
                <p>Response is a JSON object that contains 3 fields:
                <ul>
                <li><strong>result: </strong><code>OK</code>, <code>ERROR</code></li>
                <li><strong>data: </strong>response payload, it is error description if <code>result</code> is <code>ERROR</code></li>
                <li><strong>error: </strong>error code, it is an empty string if <code>result</code> is <code>OK</code></li>
                </ul>
                </p>
                <h3>Response compression</h3>
                <p>API compress a respone using gzip if the header "Accept-Encoding" contains "gzip" and a response is bigger or equal 1Kb.
                If a response is compressed then server sends the header "Content-Encoding: gzip".</p>'
        },
        'swagger' => '2.0',
        'produces' => [
                'application/json'
        ],
        'consumes' => [
                'application/json'
        ],
        'definitions' => {
                'google.V1Res' => {
                        'required' => [
                                'title',
                                'url'
                        ],
                        'type' => 'object',
                        'properties' => {
                                'url' => {
                                        'type' => 'string'
                                },
                                'title' => {
                                        'type' => 'string'
                                }
                        }
                }
        },
        'tags' => [
                {
                        'name' => 'hello'
                },
                {
                        'name' => 'search'
                }
        ],
        'paths' => {
                '/hello/world/v1/' => {
                        'get' => {
                                'summary' => 'Example handler',
                                'produces' => [
                                        'application/json'
                                ],
                                'responses' => {
                                        '200' => {
                                                'schema' => {
                                                        'type' => 'string'
                                                },
                                                'description' => 'Successful result'
                                        }
                                },
                                'description' => 'Returns \'Hello world\' string.<br/>Handler caches response.',
                                'tags' => [
                                        'hello'
                                ]
                        }
                },
                '/search/google/v1/' => {
                        'get' => {
                                'parameters' => [
                                        {
                                                'required' => bless( do{\(my $o = 1)}, 'JSON::PP::Boolean' ),
                                                'in' => 'query',
                                                'name' => 'query',
                                                'type' => 'string',
                                                'description' => 'Search query'
                                        }
                                ],
                                'summary' => 'Google search',
                                'produces' => [
                                        'application/json'
                                ],
                                'responses' => {
                                        '200' => {
                                                'schema' => {
                                                        'type' => 'array',
                                                        'items' => {
                                                                '$ref' => '#/definitions/google.V1Res'
                                                        }
                                                },
                                                'description' => 'Successful result'
                                        }
                                },
                                'description' => 'Search things in Lazada using Google<br>Handler can return these error messages:
                                <ul><li>Code: "<code>EMPTY_RESULT</code>", Data: "<code>Nothing was found</code>"</li></ul>',
                                'tags' => [
                                        'search'
                                ]
                        }
                }
        },
        'basePath' => '/'
};
