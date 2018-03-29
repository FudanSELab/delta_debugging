var app = angular.module('app', ['ngRoute', 'app.instance-controller','app.node-controller', 'app.service-controller', 'app.config-controller']);

app.config(['$routeProvider', function($routeProvider){
    $routeProvider
        .when('/instance', {
            templateUrl: 'templates/instance.html',
            controller: 'InstanceCtrl'
        })
        .when('/node',{
            templateUrl: 'templates/node.html',
            controller: 'NodeCtrl'
        })
        .when('/service',{
            templateUrl: 'templates/service.html',
            controller: 'ServiceCtrl'
        })
        .when('/config',{
            templateUrl: 'templates/config.html',
            controller: 'ConfigCtrl'
        })
        .otherwise({redirectTo:'/config'});
}

]);

app.controller('NavController', ['$scope', '$location', function($scope, $location) {
        $scope.isActive = function(destination) {
            return destination === $location.path();
        }
}]);

app.factory('loadNodeList', function ($http, $q) {
    var service = {};
    //获取并返回数据
    service.load = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;

        $http({
            method: "get",
            url: "/api/getNodesList",
            contentType: "application/json",
            dataType: "json",
            withCredentials: true
        }).success(function (data, status, headers, config) {
            if (data) {
                deferred.resolve(data);
            }
            else{
                alert("Request the order list fail!" + data.message);
            }
        });
        return promise;
    };
    return service;
});


app.factory('loadTestCases', function ($http, $q) {

    var service = {};
    //获取并返回数据
    service.loadTestList = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;

        $http({
            method: "get",
            url: "/testBackend/getFileTree",
            contentType: "application/json",
            dataType: "json",
            withCredentials: true
        }).success(function (data, status, headers, config) {
            if (data) {
                deferred.resolve(data);
            }
            else{
                alert("Request the order list fail!" + data.message);
            }
        });
        return promise;
    };
    return service;
});


app.factory('loadServiceList', function ($http, $q) {
    var service = {};
    //获取并返回数据
    service.loadServiceList = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;
        $http({
            method: "get",
            url: "/api/getServicesList",
            contentType: "application/json",
            dataType: "json",
            withCredentials: true
        }).success(function (data, status, headers, config) {
            if (data) {
                deferred.resolve(data);
            }
            else{
                alert("Request the order list fail!" + data.message);
            }
        });
        return promise;
    };
    return service;
});


app.factory('getConfigService', function ($http, $q) {
    var service = {};
    service.load = function (podString) {
        var deferred = $q.defer();
        var promise = deferred.promise;
        // $http({
        //     method: "post",
        //     url: "/api/getSinglePodLog",
        //     contentType: "application/json",
        //     dataType: "json",
        //     data:{
        //         podName: podString
        //     },
        //     withCredentials: true
        // }).success(function (data) {
        //     if (data) {
        //         deferred.resolve(data);
        //     } else{
        //         alert("Get logs fail!" + data.message);
        //     }
        // });

        var data = {
            status:true,
            message:"2333",
            data:{
                "clusterConfig":[
                    {
                        "system": "k8s",
                        "configs": [
                            {
                                "configName": "networks overlay",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "statefulset",
                                "value": "xxxxxx"
                            }
                        ]
                    },
                    {
                        "system": "istio",
                        "configs": [
                            {
                                "configName": "request routing",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "traffic shifting",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "request timeout",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "circuit breaking",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "mutual TLS authentication",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "role-based access control",
                                "value": "xxxxxx"
                            }
                        ]
                    }
                ],

                "serviceConfig": [
                    {
                        "serviceName": "service1",
                        "configs": [
                            {
                                "configName": "memory limit",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "CPU core",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "resource quota",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "pod disruption budget",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "limit range",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "pod preset",
                                "value": "xxxxxx"
                            }
                        ]
                    },
                    {
                        "serviceName": "service2",
                        "configs": [
                            {
                                "configName": "memory limit",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "CPU core",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "resource quota",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "pod disruption budget",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "limit range",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "pod preset",
                                "value": "xxxxxx"
                            }
                        ]
                    },
                    {
                        "serviceName": "service3",
                        "configs": [
                            {
                                "configName": "memory limit",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "CPU core",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "resource quota",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "pod disruption budget",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "limit range",
                                "value": "xxxxxx"
                            },
                            {
                                "configName": "pod preset",
                                "value": "xxxxxx"
                            }
                        ]
                    }
                ]
            }

        };
        deferred.resolve(data);
        return promise;
    };
    return service;
});


app.factory('getPodLogService', function ($http, $q) {
    var service = {};
    service.load = function (podString) {
        var deferred = $q.defer();
        var promise = deferred.promise;
        $http({
            method: "post",
            url: "/api/getSinglePodLog",
            contentType: "application/json",
            dataType: "json",
            data:{
                podName: podString
            },
            withCredentials: true
        }).success(function (data) {
            if (data) {
                deferred.resolve(data);
            } else{
                alert("Get logs fail!" + data.message);
            }
        });
        return promise;
    };
    return service;
});


app.factory('refreshPodsService', function ($http, $q) {
    var service = {};
    service.load = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;
        $http({
            method: "get",
            url: "/api/getPodsList",
            contentType: "application/json",
            dataType: "json",
            withCredentials: true
        }).success(function (data) {
            if (data) {
                deferred.resolve(data);
            } else{
                alert("Get logs fail!" + data.message);
            }
        });
        return promise;
    };
    return service;
});


app.filter('trustHtml', function ($sce) {
    return function (input) {
        return $sce.trustAsHtml(input);
    }
});


app.directive('icheck', ['$timeout', '$parse', function($timeout, $parse) {
    return {
        restrict: 'A',
        require: '?ngModel',
        link: function(scope, element, attr, ngModel) {
            $timeout(function() {
                var value = attr.value;

                // function update(checked) {
                //     if(attr.type==='radio') {
                //         ngModel.$setViewValue(value);
                //     } else {
                //         ngModel.$setViewValue(checked);
                //     }
                // }

                $(element).iCheck({
                    checkboxClass: attr.checkboxClass || 'icheckbox_square-blue',
                    radioClass: attr.radioClass || 'iradio_square-blue'
                }).on('ifChanged', function(e) {
                    // if ($(element).attr('type') === 'checkbox' && attr['ngModel']) {
                    //     scope.$apply(function() {
                    //         return ngModel.$setViewValue(e.target.checked);
                    //     });
                    // }
                    // if ($(element).attr('type') === 'radio' && attr['ngModel']) {
                    //     return scope.$apply(function() {
                    //         return ngModel.$setViewValue(value);
                    //     });
                    // }
                });

                // scope.$watch(attr.ngChecked, function(checked) {
                //     if(typeof checked === 'undefined') checked = !!ngModel.$viewValue;
                //     update(checked)
                // }, true);
                //
                // scope.$watch(attr.ngModel, function(model) {
                //     $(element).iCheck('update');
                // }, true);

            })
        }
    }
}]);


