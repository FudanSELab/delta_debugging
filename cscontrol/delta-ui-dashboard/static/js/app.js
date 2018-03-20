var app = angular.module('app', ['ngRoute', 'app.instance-controller','app.node-controller', 'app.service-controller']);

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
        .otherwise({redirectTo:'/service'});
}

]);

app.controller('NavController', ['$scope', '$location', function($scope, $location) {
        $scope.isActive = function(destination) {
            // console.log( $location.path());
            return destination === $location.path();
        }
}]);

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
        // var data = {
        //     "status": true,
        //     "message": "Successfully get the pod info list!",
        //     "pods": [
        //         {
        //             "name": "redis-596cfc856c-pn2mn",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.174",
        //             "startTime": "2018-03-19T03:02:13Z"
        //         },
        //         {
        //             "name": "ts-account-mongo-7db75b4948-nvlfh",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.35",
        //             "startTime": "2018-03-19T03:02:13Z"
        //         },
        //         {
        //             "name": "ts-admin-basic-info-service-6cdfc758cf-8wcr5",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.188",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-admin-order-service-7b67cd8c94-g8vv5",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.46",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-admin-route-service-77b44b976d-2s2qj",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.185",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-admin-travel-service-56cdd9897b-lc97q",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.48",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-admin-user-service-7b7f4947fd-d4lwg",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.186",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-assurance-mongo-65d4bd64-hlx65",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.42",
        //             "startTime": "2018-03-19T03:02:15Z"
        //         },
        //         {
        //             "name": "ts-assurance-service-84c6b45f89-w5dpd",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.47",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-basic-service-d56bb447c-q6xpj",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.187",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-cancel-service-6d46bf5b89-7chvb",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.49",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-config-mongo-d47fbc78b-tmb9q",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.177",
        //             "startTime": "2018-03-19T03:02:13Z"
        //         },
        //         {
        //             "name": "ts-config-service-7d49ddbfb9-znq4j",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.189",
        //             "startTime": "2018-03-19T03:04:20Z"
        //         },
        //         {
        //             "name": "ts-consign-mongo-6f76c788b8-clvfh",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.44",
        //             "startTime": "2018-03-19T03:02:16Z"
        //         },
        //         {
        //             "name": "ts-consign-price-mongo-766fdf6f7b-qpfjr",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.183",
        //             "startTime": "2018-03-19T03:02:16Z"
        //         },
        //         {
        //             "name": "ts-consign-price-service-7586dcbb5-wszhl",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.50",
        //             "startTime": "2018-03-19T03:04:21Z"
        //         },
        //         {
        //             "name": "ts-consign-service-5c86d47488-jxzn6",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.190",
        //             "startTime": "2018-03-19T03:04:21Z"
        //         },
        //         {
        //             "name": "ts-contacts-mongo-688c557cb4-dbm6f",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.36",
        //             "startTime": "2018-03-19T03:02:13Z"
        //         },
        //         {
        //             "name": "ts-contacts-service-66c84c56d5-zxbrw",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.51",
        //             "startTime": "2018-03-19T03:04:21Z"
        //         },
        //         {
        //             "name": "ts-execute-service-59f88cb84-4dn9b",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.192",
        //             "startTime": "2018-03-19T03:04:21Z"
        //         },
        //         {
        //             "name": "ts-food-map-mongo-7667f798b8-48vrv",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.181",
        //             "startTime": "2018-03-19T03:02:16Z"
        //         },
        //         {
        //             "name": "ts-food-map-service-6b9d46b7-sr8rr",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.52",
        //             "startTime": "2018-03-19T03:04:21Z"
        //         },
        //         {
        //             "name": "ts-food-mongo-7f59665bc4-xxpvp",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.45",
        //             "startTime": "2018-03-19T03:02:16Z"
        //         },
        //         {
        //             "name": "ts-food-service-7db787f64c-zvncq",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.191",
        //             "startTime": "2018-03-19T03:04:22Z"
        //         },
        //         {
        //             "name": "ts-inside-payment-mongo-8cdc4b748-2jdc7",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.41",
        //             "startTime": "2018-03-19T03:02:15Z"
        //         },
        //         {
        //             "name": "ts-inside-payment-service-b77666797-l6wvq",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.53",
        //             "startTime": "2018-03-19T03:04:22Z"
        //         },
        //         {
        //             "name": "ts-login-service-558f589486-v7dj8",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.193",
        //             "startTime": "2018-03-19T03:04:22Z"
        //         },
        //         {
        //             "name": "ts-news-service-655f7fcbff-6h7t2",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.54",
        //             "startTime": "2018-03-19T03:04:22Z"
        //         },
        //         {
        //             "name": "ts-notification-service-75cff6d8bd-6hsl4",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.194",
        //             "startTime": "2018-03-19T03:04:22Z"
        //         },
        //         {
        //             "name": "ts-order-mongo-5db5b7d864-9bcpm",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.176",
        //             "startTime": "2018-03-19T03:02:13Z"
        //         },
        //         {
        //             "name": "ts-order-other-mongo-79bd5dcbd4-rqdh9",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.37",
        //             "startTime": "2018-03-19T03:02:13Z"
        //         },
        //         {
        //             "name": "ts-order-other-service-78d95488fd-cqxxc",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.55",
        //             "startTime": "2018-03-19T03:04:22Z"
        //         },
        //         {
        //             "name": "ts-order-service-f88d5c588-pcf7h",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.195",
        //             "startTime": "2018-03-19T03:04:23Z"
        //         },
        //         {
        //             "name": "ts-payment-mongo-b66487b6d-zdlf8",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.184",
        //             "startTime": "2018-03-19T03:02:15Z"
        //         },
        //         {
        //             "name": "ts-payment-service-6b4d58fcb6-4xbwb",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.56",
        //             "startTime": "2018-03-19T03:04:23Z"
        //         },
        //         {
        //             "name": "ts-preserve-other-service-887cfb447-5znn2",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.196",
        //             "startTime": "2018-03-19T03:04:23Z"
        //         },
        //         {
        //             "name": "ts-preserve-service-d85dbd75c-plrxk",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.57",
        //             "startTime": "2018-03-19T03:04:23Z"
        //         },
        //         {
        //             "name": "ts-price-mongo-6d7dd5c476-5chbc",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.40",
        //             "startTime": "2018-03-19T03:02:14Z"
        //         },
        //         {
        //             "name": "ts-price-service-6d6757874b-8pz58",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.197",
        //             "startTime": "2018-03-19T03:04:23Z"
        //         },
        //         {
        //             "name": "ts-rebook-service-7d46b45676-r6vb5",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.58",
        //             "startTime": "2018-03-19T03:04:24Z"
        //         },
        //         {
        //             "name": "ts-register-service-68969d67f8-4jvp8",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.198",
        //             "startTime": "2018-03-19T03:04:24Z"
        //         },
        //         {
        //             "name": "ts-route-mongo-5c7cb85cb4-czhkv",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.175",
        //             "startTime": "2018-03-19T03:02:13Z"
        //         },
        //         {
        //             "name": "ts-route-plan-service-d998f8b5c-8zmjg",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.59",
        //             "startTime": "2018-03-19T03:04:24Z"
        //         },
        //         {
        //             "name": "ts-route-service-6f48c4999d-q94bq",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.202",
        //             "startTime": "2018-03-19T03:04:25Z"
        //         },
        //         {
        //             "name": "ts-seat-service-865c86df4d-lqvcd",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.60",
        //             "startTime": "2018-03-19T03:04:24Z"
        //         },
        //         {
        //             "name": "ts-security-mongo-54b85db674-rc4k6",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.180",
        //             "startTime": "2018-03-19T03:02:15Z"
        //         },
        //         {
        //             "name": "ts-security-service-77fc96989d-fs6w5",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.203",
        //             "startTime": "2018-03-19T03:04:25Z"
        //         },
        //         {
        //             "name": "ts-sso-service-67496fbb87-wwtlf",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.63",
        //             "startTime": "2018-03-19T03:04:25Z"
        //         },
        //         {
        //             "name": "ts-station-mongo-5d8cdb754d-qjrhw",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.38",
        //             "startTime": "2018-03-19T03:02:13Z"
        //         },
        //         {
        //             "name": "ts-station-service-7cc6758c86-8gbn2",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.200",
        //             "startTime": "2018-03-19T03:04:25Z"
        //         },
        //         {
        //             "name": "ts-ticket-office-mongo-54bdc8d996-9tfl9",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.182",
        //             "startTime": "2018-03-19T03:02:15Z"
        //         },
        //         {
        //             "name": "ts-ticket-office-service-6f4d8c5d4-nfhd8",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.61",
        //             "startTime": "2018-03-19T03:04:25Z"
        //         },
        //         {
        //             "name": "ts-ticketinfo-service-57cddcd894-ls62s",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.199",
        //             "startTime": "2018-03-19T03:04:25Z"
        //         },
        //         {
        //             "name": "ts-train-mongo-7c6c4b478d-hgm4d",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.178",
        //             "startTime": "2018-03-19T03:02:14Z"
        //         },
        //         {
        //             "name": "ts-train-service-68f46c6cb8-c767z",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.62",
        //             "startTime": "2018-03-19T03:04:26Z"
        //         },
        //         {
        //             "name": "ts-travel-mongo-56c7b5b95d-trdpm",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.39",
        //             "startTime": "2018-03-19T03:02:14Z"
        //         },
        //         {
        //             "name": "ts-travel-plan-service-555f8dd6b9-vchs5",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.64",
        //             "startTime": "2018-03-19T03:04:26Z"
        //         },
        //         {
        //             "name": "ts-travel-service-79cc6ff645-m9ngq",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.205",
        //             "startTime": "2018-03-19T03:04:26Z"
        //         },
        //         {
        //             "name": "ts-travel2-mongo-77db8785d8-84c4l",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.179",
        //             "startTime": "2018-03-19T03:02:14Z"
        //         },
        //         {
        //             "name": "ts-travel2-service-5689d4f4c-mlkrv",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.201",
        //             "startTime": "2018-03-19T03:04:26Z"
        //         },
        //         {
        //             "name": "ts-ui-dashboard-66f6d5d5d8-9jfxw",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.66",
        //             "startTime": "2018-03-19T06:12:07Z"
        //         },
        //         {
        //             "name": "ts-verification-code-service-7d67d56fff-svmrc",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.65",
        //             "startTime": "2018-03-19T03:04:26Z"
        //         },
        //         {
        //             "name": "ts-voucher-mysql-6874d55496-4hkkh",
        //             "nodeName": "centos-minion-1",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.179",
        //             "podIP": "10.244.3.43",
        //             "startTime": "2018-03-19T03:02:15Z"
        //         },
        //         {
        //             "name": "ts-voucher-service-7c8689d746-25fvv",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.204",
        //             "startTime": "2018-03-19T03:04:27Z"
        //         },
        //         {
        //             "name": "zipkin-5988589cc6-m5tc7",
        //             "nodeName": "centos-minion-2",
        //             "status": "Running",
        //             "nodeIP": "10.141.211.180",
        //             "podIP": "10.244.2.206",
        //             "startTime": "2018-03-19T06:12:15Z"
        //         }
        //     ]
        // };
        // deferred.resolve(data);
        return promise;
    };
    return service;
});


app.filter('trustHtml', function ($sce) {
    return function (input) {
        // console.log("23333");
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


