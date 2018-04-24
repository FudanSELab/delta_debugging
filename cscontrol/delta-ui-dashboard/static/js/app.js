var app = angular.module('app', ['ngRoute', 'app.instance-controller','app.node-controller', 'app.service-controller', 'app.config-controller', 'app.sequence-controller','app.mixer-controller']);

app.config(['$routeProvider', function($routeProvider){
    $routeProvider
        .when('/instance', {
            templateUrl: 'templates/instance.html',
            controller: 'InstanceCtrl',
            cache: true
        })
        .when('/node',{
            templateUrl: 'templates/node.html',
            controller: 'NodeCtrl',
            cache: true
        })
        .when('/service',{
            templateUrl: 'templates/service.html',
            controller: 'ServiceCtrl',
            cache: true
        })
        .when('/config',{
            templateUrl: 'templates/config.html',
            controller: 'ConfigCtrl',
            cache: true
        })
        .when('/sequence',{
            templateUrl: 'templates/sequence.html',
            controller: 'SequenceCtrl',
            cache: true
        })
        .when('/mixer',{
            templateUrl: 'templates/mixer.html',
            controller: 'MixerCtrl',
            cache: true
        })
        .otherwise({redirectTo:'/mixer'});
}

]);


app.constant('defaultCluster', 'cluster4');

app.controller('NavController', ['$scope', '$location', function($scope, $location) {
        $scope.isActive = function(destination) {
            return destination === $location.path();
        }
}]);

app.factory('loadNodeList', function ($http, $q, defaultCluster) {
    var service = {};
    //获取并返回数据
    service.load = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;

        $http({
            method: "get",
            url: "/api/getNodesList/" + defaultCluster,
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
            cache:true,
            withCredentials: true
        }).success(function (data, status, headers, config) {
            if (data) {
                deferred.resolve(data);
            }
            else{
                alert("Request the order list fail!" + data.message);
            }
        });

        // var data = {
        //     "Delta Test":[
        //         "TestBookErrorConfig",
        //         "TestLoginErrorInstance",
        //         "TestLoginErrorInstance2",
        //         "TestCancelErrorSequence",
        //         "TestMiniSequenceError"
        //     ]
        //
        // };
        // deferred.resolve(data);
        return promise;
    };
    return service;
});


app.factory('loadServiceList', function ($http, $q, defaultCluster) {
    var service = {};
    //获取并返回数据
    service.loadServiceList = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;
        $http({
            method: "get",
            url: "/api/getServicesList/" + defaultCluster,
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
        // var data = {
        //     "status": true,
        //     "message": "Get the service list successfully!",
        //     "services": [
        //         {
        //             "serviceName": "redis",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-account-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-admin-basic-info-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-admin-order-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-admin-route-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-admin-travel-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-admin-user-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-assurance-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-assurance-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-basic-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-cancel-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-config-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-config-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-consign-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-consign-price-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-consign-price-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-consign-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-contacts-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-contacts-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-execute-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-food-map-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-food-map-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-food-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-food-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-inside-payment-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-inside-payment-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-login-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-news-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-notification-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-order-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-order-other-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-order-other-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-order-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-payment-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-payment-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-preserve-other-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-preserve-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-price-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-price-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-rebook-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-register-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-route-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-route-plan-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-route-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-seat-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-security-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-security-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-sso-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-station-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-station-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-ticket-office-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-ticket-office-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-ticketinfo-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-train-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-train-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-travel-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-travel-plan-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-travel-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-travel2-mongo",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-travel2-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-ui-dashboard",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-verification-code-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-voucher-mysql",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "ts-voucher-service",
        //             "numOfReplicas": 1
        //         },
        //         {
        //             "serviceName": "zipkin",
        //             "numOfReplicas": 1
        //         }
        //     ]
        // };
        // deferred.resolve(data);
        return promise;
    };
    return service;
});


app.factory('getConfigService', function ($http, $q, defaultCluster) {
    var service = {};
    service.load = function (podString) {
        var deferred = $q.defer();
        var promise = deferred.promise;
        $http({
            method: "get",
            url: "/api/getServicesAndConfig/" + defaultCluster,
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

        // var data = {
        //     "status": true,
        //     "message": "Get the services and the corresponding config successfully!",
        //     "services": [
        //         {
        //             "serviceName": "config-example",
        //             "limits": {
        //                 "memory": "128Mi",
        //                 "cpu": "500m"
        //             },
        //             "requests": {
        //                 "memory": "90Mi",
        //                 "cpu": "250m"
        //             }
        //         },
        //         {
        //             "serviceName": "redis",
        //             "limits": {
        //                 "memory": "128Mi",
        //                 "cpu": "500m"
        //             },
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-account-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-admin-basic-info-service",
        //             "limits": {
        //                 "memory": "128Mi",
        //                 "cpu": "500m"
        //             },
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-admin-order-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-admin-route-service",
        //             "limits": {
        //                 "memory": "128Mi",
        //                 "cpu": "500m"
        //             },
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-admin-travel-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-admin-user-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-assurance-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-assurance-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-basic-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-cancel-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-config-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-config-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-consign-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-consign-price-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-consign-price-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-consign-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-contacts-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-contacts-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-execute-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-food-map-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-food-map-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-food-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-food-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-inside-payment-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-inside-payment-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-login-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-news-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-notification-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-order-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-order-other-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-order-other-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-order-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-payment-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-payment-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-preserve-other-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-preserve-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-price-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-price-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-rebook-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-register-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-route-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-route-plan-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-route-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-seat-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-security-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-security-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-sso-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-station-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-station-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-ticket-office-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-ticket-office-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-ticketinfo-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-train-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-train-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-travel-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-travel-plan-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-travel-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-travel2-mongo",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-travel2-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-ui-dashboard",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-verification-code-service",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-voucher-mysql",
        //             "limits": null,
        //             "requests": null
        //         },
        //         {
        //             "serviceName": "ts-voucher-service",
        //             "limits": null,
        //             "requests": null
        //         }
        //     ]
        // };
        // deferred.resolve(data);

        return promise;
    };
    return service;
});


app.factory('getPodLogService', function ($http, $q, defaultCluster) {
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
                clusterName: defaultCluster,
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


app.factory('refreshPodsService', function ($http, $q, defaultCluster) {
    var service = {};
    service.load = function () {
        var deferred = $q.defer();
        var promise = deferred.promise;
        $http({
            method: "get",
            url: "/api/getPodsList/" + defaultCluster,
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


//---------------------------------------------------------------------------
//---------------------------- Add by jichao --------------------------------
//---------------------------------------------------------------------------
app.factory('suspendServiceWithSource', function ($http, $q, defaultCluster) {
    var service = {};
    service.load = function (sourceSvcName, svcName) {
        $http({
            method: "post",
            url: "/api/setServiceRequestSuspendWithSourceSvc",
            contentType: "application/json",
            dataType: "json",
            data:{
                clusterName: defaultCluster,
                sourceSvcName: sourceSvcName,
                svc: svcName
            },
            withCredentials: true
        }).success(function (data) {
            if (data.status) {
                alert(data.message);
            } else{
                alert("SuspendServiceWithSource fail!" + data.message);
            }
        });
    };
    return service;
});

app.factory('unsuspendServiceWithSource', function ($http, $q, defaultCluster) {
    var service = {};
    service.load = function (sourceSvcName, svcName) {
        $http({
            method: "post",
            url: "/api/unsetServiceRequestSuspend",
            contentType: "application/json",
            dataType: "json",
            data:{
                clusterName: defaultCluster,
                sourceSvcName: sourceSvcName,
                svc: svcName
            },
            withCredentials: true
        }).success(function (data) {
            if (data.status) {
                alert(data.message);
            } else{
                alert("UnsuspendServiceWithSource fail!" + data.message);
            }
        });
    };
    return service;
});

app.factory('setAsyncRequestSequenceWithSrc', function ($http, $q, defaultCluster) {
    var service = {};
    service.load = function (sourceSvcName, svcListString) {
        $http({
            method: "post",
            url: "/api/setAsyncRequestSequenceWithSrc",
            contentType: "application/json",
            dataType: "json",
            data:{
                clusterName: defaultCluster,
                sourceSvcName: sourceSvcName,
                svc: svcListString.split(",")
            },
            withCredentials: true
        }).success(function (data) {
            if (data.status) {
                alert(data.message);
            } else{
                alert("Sequence Control fail!" + data.message);
            }
        });
    };
    return service;
});

