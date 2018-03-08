var serviceApp = angular.module('serviceApp', ['ngRoute']);

serviceApp.config(['$routeProvider',
    function($routeProvider) {
        $routeProvider.
            when('/status/:resource', {
                templateUrl: 'view/status.html',
            }).
            when('/logs/:file/:id', {
                templateUrl: 'view/session.html',
            }).
            when('/logs/', {
                templateUrl: 'view/session.html',
            }).
            when('/settings/', {
                templateUrl: 'view/settings.html',
            }).
            when('/pprof/', {
                templateUrl: 'view/pprof.html',
            }).
            otherwise({
                redirectTo: '/status/self',
            });
    }]);

serviceApp.controller('TabCtrl', ['$scope', '$http', '$routeParams',
    function($scope, $http, $routeParams) {
        var resourceMap = {};
        var tab = $routeParams.resource;

        var setTab = function (tabId) {
            tab = tabId;
            if (resourceMap[tabId]) {
                $scope.resource = tabId;
            }
        };

        this.setTab = setTab;

        this.isSet = function (tabId) {
            return tab === tabId;
        };

        $http.get('status/resources').then(function(response) {
            $scope.resources = response.data.resources.sort();
            $scope.self_name = response.data.caption;

            resourceMap['self'] = true;
            $scope.resources.forEach(function(resource) {
                resourceMap[resource] = true;
            });

            setTab($routeParams.resource);
        }, function() {});
    }
]);

serviceApp.controller('NavCtrl', ['$scope', '$location', '$http',
    function($scope, $location, $http) {
        $scope.isActive = function(path) {
            return $location.url().startsWith(path);
        };

        $http.get('meta').then(function(response) {
            document.title = response.data["service_id"] + ' [' + response.data["venture"] +
                            ' - ' + response.data["env"] + ']';
            $scope.meta = response.data;
        }, function() {});
    }
]);

serviceApp.controller('StatusCtrl', ['$scope',
    function($scope) {
        $scope.$watch('resource', function() {
            if ($scope.resource == $scope.resourceName) {
                connect();
            } else {
                disconnect();
            }
        });

        $scope.$on('$destroy', function() {
            disconnect();
        });

        var ws;

        function disconnect() {
            if (ws) {
                ws.close();
                ws = null;
            }
        }

        function connect() {
            if (ws) {
                return;
            }

            $scope.status = 'connecting';
            if ($scope.resourceName == 'self') {
                ws = new WebSocket("ws://" + location.host + "/status/ws-status");
            } else {
                ws = new WebSocket("ws://" + location.host + "/status/ws-resource?resource="+encodeURIComponent($scope.resource));
            }
            ws.onopen = function() {
                $scope.status = 'connected';
                $scope.$apply();
            };

            ws.onclose = function() {
                $scope.status = 'disconnected';
                ws = null;
                $scope.$apply();
            };

            ws.onerror = function() {
                $scope.status = 'disconnected';
                ws = null;
                $scope.$apply();
            };

            ws.onmessage = function(e) {
                $scope.stats = JSON.parse(e.data);
                $scope.$apply();
            };
        }
        $scope.connect = connect;
    }
]);


serviceApp.controller('SessionCtrl', ['$scope', '$http', '$routeParams', '$location', '$sce',
    function($scope, $http, $routeParams, $location, $sce) {
        $scope.sessionNav = {
            file: $routeParams.file,
            id: $routeParams.id
        };

        $scope.updateFiles = function() {
            $scope.updatingFiles = true;
            $http.get("logs/files").then(function(response) {
                $scope.files = response.data.sort().reverse();
                $scope.updatingFiles = false;
                $scope.sessionNav.file = $scope.files[0];
            }, function(response) {
                $scope.updatingFiles = false;
            });
        };

        $scope.setActive = function() {
            $scope.curSession = this.session;
            $scope.reqVar = $scope.curSession.request_dump;
        };

        $scope.navigateSubmit = function() {
            $location.path('logs/' + this.sessionNav.file + '/' + this.sessionNav.id);
        };

        $scope.updateFiles();

        if ($routeParams.file && $routeParams.id) {
            $scope.state = 'loading';
            $scope.error = null;

            $http.get("logs/session", {
                params: {
                    file: $routeParams.file,
                    id: $routeParams.id
                }
            }).then(function(response) {
                fixData($sce, response.data);
                $scope.curSession = $scope.session = response.data;
                $scope.state = 'loaded';
            }, function(response) {
                $scope.state = 'error';
                $scope.error = response.data;
            });
        } else {
            $scope.state = 'no_session';
        }
    }
]);


serviceApp.controller('SettingsCtrl', ['$scope', '$http',
    function($scope, $http) {
        $scope.curUser = {email: 'test@example.com'};

        $scope.connectWS = function() {
            var ws = $scope.ws;
            if (ws) {
                $scope.ws = null;
                ws.onclose = function() {};
                ws.close();
            }

            $scope.status = 'connecting';
            ws = new WebSocket("ws://" + location.host + "/settings/ws");
            $scope.ws = ws;

            $scope.ws.onopen = function() {
                $scope.status = 'connected';
                $scope.$apply();
            };

            $scope.ws.onmessage = function(evt) {
                var message = JSON.parse(evt.data);

                var type = message['type'];
                if (!type) {
                    return;
                }

                switch (type) {
                    case 'SETTINGS_LIST':
                        $scope.settings = sortSettings(message.settings);
                        break;
                    case 'SETTING_CHANGE':
                        var settingExists = false;
                        for (var i = 0; i < $scope.settings.length; i++) {
                            if ($scope.settings[i].key == message.setting.key) {
                                $scope.settings[i] = message.setting;
                                settingExists = true;
                                break;
                            }
                        }

                        if (!settingExists) {
                            $scope.settings.push(message.setting);
                            $scope.settings = sortSettings($scope.settings);
                        }
                        break;
                }

                $scope.$apply();
            };

            $scope.ws.onclose = function() {
                $scope.status = 'disconnected';
                $scope.$apply();
            };

            $scope.ws.onerror = function() {
                $scope.status = 'disconnected';
                $scope.$apply();
            }
        };

        $scope.editSettingClick = function() {
            $scope.editingSetting = {
                key: this.setting.key,
                value: this.setting.value
            };
            angular.element('#editSettingModal').modal('show');
        };

        $scope.editSettingModalFormSubmit = function() {
            $http.get("/settings/edit", {
                params: $scope.editingSetting
            }).then(function(response) {
                if (response.data.result == "OK") {
                    angular.element('#editSettingModal').modal('hide');
                    $scope.editSettingFormError = null;
                } else {
                    $scope.editSettingFormError = data.error;
                }
            }, function(response) {
                $scope.editSettingFormError = response.data;
            });
        };

        $scope.historyClick = function() {
            $scope.historySetting = this.setting;
            angular.element('#showHistoryModal').modal('show');
        };

        $scope.connectWS();
    }
]);

function fixData($sce, session) {
    if (!session) {
        return;
    }

    var maxResponseDate;

    if (session.raw_request_dump) {
        session.request_dump = $sce.trustAsHtml(session.raw_request_dump)
    }

    if (session.responses) {
        for (var i = 0; i < session.responses.length; i++) {
            if (session.responses[i].raw_dump) {
                session.responses[i].dump = $sce.trustAsHtml(session.responses[i].raw_dump);
            }
            if (session.responses[i].error_message) {
                session.responses[i].error_message = $sce.trustAsHtml(session.responses[i].error_message)
            }
            var respDate = new Date(session.responses[i].time);
            if (!maxResponseDate || maxResponseDate < respDate) {
                maxResponseDate = respDate;
            }
        }
    }

    if (session.errors) {
        for (var i = 0; i < session.errors.length; i++) {
            if (session.errors[i].raw_dump) {
                session.errors[i].dump = $sce.trustAsHtml(session.errors[i].raw_dump);
            }
            var errDate = new Date(session.errors[i].time);
            if (!maxResponseDate || maxResponseDate < errDate) {
                maxResponseDate = errDate;
            }
        }
    }

    if (maxResponseDate) {
        session.duration = (maxResponseDate - new Date(session.request_time)) / 1000;
    }

    if (session.children) {
        for (var i = 0; i < session.children.length; i++) {
            fixData($sce, session.children[i]);
        }
    }
}

function sortSettings(settings) {
    return settings.sort(function(a, b) {
        if (a.key.toUpperCase() < b.key.toUpperCase()) {
            return -1;
        } else if (a.key.toUpperCase() == b.key.toUpperCase()) {
            return 0;
        } else {
            return 1;
        }
    })
}

serviceApp.filter('percentage', ['$filter', function ($filter) {
    return function (input, decimals) {
        return $filter('number')(input * 100, decimals) + '%';
    };
}]);

serviceApp.filter('duration', [function () {
    return function (input) {
        var b = ['ns', 'µs', 'ms', 'm', 's', 'h'];
        var n = [1000, 1000, 1000, 60, 60, 0];
        for (var i = 0; i < b.length - 1; i++) {
            var k = n[i];
            n[i] = input % k;
            input = parseInt(input / k);
        }
        n[i] = input;
        for (; i > 0; i--) {
            if (n[i] > 0) {
                var d = n[i] + b[i];
                if (n[i - 1] > 0) {
                    d += ' ' + n[i - 1] + b[i - 1]
                }
                return d
            }
        }
        return n[0] + b[0]
    };
}]);

